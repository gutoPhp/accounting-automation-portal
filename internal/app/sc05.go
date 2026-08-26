package app

import (
	"net/http"
	"time"
)

const (
	blockAction   = "block"
	unblockAction = "unblock"
)

type sc05RunRequest struct {
	ClientID        string `json:"client_id"`
	Action          string `json:"action"`
	SimulateFailure bool   `json:"simulate_failure"`
}

type simulatedSystemState struct {
	System string `json:"system"`
	State  string `json:"state"`
	Detail string `json:"detail"`
}

func (a *App) registerSC05Routes(mux *http.ServeMux) {
	mux.HandleFunc(
		"GET /api/sc-05/systems",
		a.requireModule("SC-05", a.sc05Systems),
	)
	mux.HandleFunc(
		"POST /api/sc-05/run",
		a.requireModule("SC-05", a.runSC05),
	)
}

func (a *App) sc05Systems(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	client := findClient(a.store.snapshot().Clients, clientID)
	if client == nil {
		writeError(w, http.StatusNotFound, "Cliente não encontrado.")
		return
	}

	writeJSON(w, http.StatusOK, systemStatesForClient(*client))
}

func systemStatesForClient(client Client) []simulatedSystemState {
	accessState := "Ativo"
	financialState := "Operação normal"
	if client.Blocked {
		accessState = "Bloqueado"
		financialState = "Acesso suspenso"
	}

	return []simulatedSystemState{
		{
			System: "Gestão contábil",
			State:  accessState,
			Detail: "Permissão de acesso do cliente",
		},
		{
			System: "Portal financeiro",
			State:  financialState,
			Detail: "Situação operacional da conta",
		},
		{
			System: "Sistema de tarefas",
			State:  client.TaskOwner,
			Detail: "Responsável atual pelas tarefas",
		},
	}
}

func (a *App) runSC05(w http.ResponseWriter, r *http.Request) {
	user, _ := a.currentUser(r)

	var input sc05RunRequest
	if err := decodeJSON(w, r, &input); err != nil || !validSC05Action(input.Action) {
		writeError(w, http.StatusBadRequest, "Informe cliente e ação válidos.")
		return
	}

	execution := newSC05Execution(user.Name)
	err := a.store.update(func(db *database) error {
		client := findClient(db.Clients, input.ClientID)
		if client == nil {
			return errNotFound
		}

		execution.Actions = runSystemAdapters(input)
		completeSC05Execution(&execution, client, input.Action)
		db.Executions = append(db.Executions, execution)
		return nil
	})

	if err != nil {
		writeError(w, http.StatusNotFound, readableError(err))
		return
	}

	writeJSON(w, http.StatusOK, execution)
}

func newSC05Execution(triggeredBy string) Execution {
	return Execution{
		ID:          newID("exec"),
		Module:      "SC-05",
		StartedAt:   time.Now().UTC(),
		TriggeredBy: triggeredBy,
		Trigger:     "manual",
		Status:      "success",
	}
}

func runSystemAdapters(input sc05RunRequest) []SystemAction {
	systems := []string{
		"Gestão contábil",
		"Portal financeiro",
		"Sistema de tarefas",
	}
	actions := make([]SystemAction, 0, len(systems))

	for index, system := range systems {
		action := SystemAction{
			System: system,
			Action: adapterAction(system, input.Action),
			Status: "success",
			Detail: "Operação confirmada pelo adaptador simulado.",
		}

		if input.SimulateFailure && index == 1 {
			action.Status = "failed"
			action.Detail = "Sistema indisponível após 3 tentativas."
			actions = append(actions, action)
			break
		}

		actions = append(actions, action)
	}

	return actions
}

func adapterAction(system, action string) string {
	if system != "Sistema de tarefas" {
		return action
	}
	if action == blockAction {
		return "trocar responsável para BLOQUEADO"
	}
	return "restaurar responsável anterior"
}

func completeSC05Execution(execution *Execution, client *Client, action string) {
	execution.FinishedAt = executionFinishedAt(execution.StartedAt)

	if hasFailedAction(execution.Actions) {
		execution.Status = "failed"
		execution.Error = "Portal financeiro não respondeu. As demais etapas foram interrompidas."
		execution.Summary = "Execução parcial; nenhuma alteração local foi confirmada."
		return
	}

	if action == blockAction {
		blockClient(client)
		execution.Summary = "Cliente bloqueado nos sistemas e tarefas preservadas."
		return
	}

	unblockClient(client)
	execution.Summary = "Cliente desbloqueado e responsável das tarefas restaurado."
}

func blockClient(client *Client) {
	if client.Blocked {
		return
	}

	client.Blocked = true
	if client.TaskOwner != "BLOQUEADO" {
		client.PreviousOwner = client.TaskOwner
	}
	client.TaskOwner = "BLOQUEADO"
}

func unblockClient(client *Client) {
	client.Blocked = false
	ownerToRestore := client.PreviousOwner
	if ownerToRestore == "" || ownerToRestore == "BLOQUEADO" {
		ownerToRestore = client.DefaultOwner
	}
	client.TaskOwner = ownerToRestore
	client.PreviousOwner = ""
}

func findClient(clients []Client, clientID string) *Client {
	for index := range clients {
		if clients[index].ID == clientID {
			return &clients[index]
		}
	}
	return nil
}

func validSC05Action(action string) bool {
	return action == blockAction || action == unblockAction
}

func hasFailedAction(actions []SystemAction) bool {
	for _, action := range actions {
		if action.Status == "failed" {
			return true
		}
	}
	return false
}
