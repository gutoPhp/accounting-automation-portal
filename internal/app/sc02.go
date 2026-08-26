package app

import (
	"fmt"
	"net/http"
	"sort"
	"time"
)

var fiscalAgencies = []string{
	"Receita Federal",
	"FGTS",
	"Secretaria Estadual",
}

type sc02RunRequest struct {
	SimulateFailure bool `json:"simulate_failure"`
}

func (a *App) registerSC02Routes(mux *http.ServeMux) {
	mux.HandleFunc(
		"GET /api/sc-02/checks",
		a.requireModule("SC-02", a.fiscalChecks),
	)
	mux.HandleFunc(
		"POST /api/sc-02/run",
		a.requireModule("SC-02", a.runSC02),
	)
	mux.HandleFunc(
		"GET /api/sc-02/schedule",
		a.requireModule("SC-02", a.sc02Schedule),
	)
}

func (a *App) fiscalChecks(w http.ResponseWriter, _ *http.Request) {
	checks := a.store.snapshot().FiscalChecks
	if checks == nil {
		checks = make([]FiscalCheck, 0)
	}

	sort.Slice(checks, func(i, j int) bool {
		return checks[i].CheckedAt.After(checks[j].CheckedAt)
	})

	writeJSON(w, http.StatusOK, checks)
}

func (a *App) runSC02(w http.ResponseWriter, r *http.Request) {
	user, _ := a.currentUser(r)

	var input sc02RunRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Configuração de consulta inválida.")
		return
	}

	execution := a.executeFiscalScan(user.Name, "manual", input.SimulateFailure)
	writeJSON(w, http.StatusOK, execution)
}

func (a *App) executeFiscalScan(
	triggeredBy string,
	trigger string,
	simulateFailure bool,
) Execution {
	startedAt := time.Now().UTC()
	databaseSnapshot := a.store.snapshot()
	checks := buildFiscalChecks(databaseSnapshot.Clients, startedAt, simulateFailure)
	failures := countFiscalFailures(checks)
	irregularities := countFiscalIrregularities(checks)

	execution := Execution{
		ID:          newID("exec"),
		Module:      "SC-02",
		StartedAt:   startedAt,
		FinishedAt:  executionFinishedAt(startedAt),
		TriggeredBy: triggeredBy,
		Trigger:     trigger,
		Status:      "success",
		Summary:     fiscalScanSummary(irregularities, failures),
	}
	if failures > 0 {
		execution.Status = "failed"
		execution.Error = "Uma ou mais fontes não responderam; as tentativas foram preservadas."
	}

	_ = a.store.update(func(db *database) error {
		db.FiscalChecks = checks
		db.Executions = append(db.Executions, execution)
		return nil
	})

	return execution
}

func buildFiscalChecks(
	clients []Client,
	checkedAt time.Time,
	simulateFailure bool,
) []FiscalCheck {
	checks := make([]FiscalCheck, 0, len(clients)*len(fiscalAgencies))

	for clientIndex, client := range clients {
		for agencyIndex, agency := range fiscalAgencies {
			check := FiscalCheck{
				ID:         newID("check"),
				ClientID:   client.ID,
				ClientName: client.Name,
				Agency:     agency,
				Status:     fiscalStatus(clientIndex, agencyIndex),
				CheckedAt:  checkedAt,
				Attempts:   1,
			}

			if simulateFailure && clientIndex == 0 && agency == "FGTS" {
				check.Status = "failed"
				check.Error = "Portal indisponível após 3 tentativas."
				check.Attempts = 3
			}

			checks = append(checks, check)
		}
	}

	return checks
}

func fiscalStatus(clientIndex, agencyIndex int) string {
	if (clientIndex+agencyIndex)%4 == 0 {
		return "irregular"
	}
	return "regular"
}

func countFiscalFailures(checks []FiscalCheck) int {
	return countChecksWithStatus(checks, "failed")
}

func countFiscalIrregularities(checks []FiscalCheck) int {
	return countChecksWithStatus(checks, "irregular")
}

func countChecksWithStatus(checks []FiscalCheck, status string) int {
	count := 0
	for _, check := range checks {
		if check.Status == status {
			count++
		}
	}
	return count
}

func fiscalScanSummary(irregularities, failures int) string {
	return fmt.Sprintf(
		"Consulta concluída: %d irregularidade(s) e %d falha(s) de fonte.",
		irregularities,
		failures,
	)
}

func (a *App) ensureMonthlySC02() {
	now := time.Now().UTC()
	executions := a.store.snapshot().Executions

	for _, execution := range executions {
		isCurrentMonth := execution.StartedAt.Year() == now.Year() &&
			execution.StartedAt.Month() == now.Month()
		if execution.Module == "SC-02" && execution.Trigger == "scheduled" && isCurrentMonth {
			return
		}
	}

	a.executeFiscalScan("Agendador", "scheduled", false)
}

func (a *App) sc02Schedule(w http.ResponseWriter, _ *http.Request) {
	now := time.Now().UTC()
	status := scheduleStatus{
		Enabled:   true,
		Frequency: "Mensal",
		NextRunAt: time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC),
	}

	executions := a.store.snapshot().Executions
	for index := len(executions) - 1; index >= 0; index-- {
		execution := executions[index]
		if execution.Module == "SC-02" && execution.Trigger == "scheduled" {
			status.LastRunAt = &execution.FinishedAt
			status.LastStatus = execution.Status
			break
		}
	}

	writeJSON(w, http.StatusOK, status)
}
