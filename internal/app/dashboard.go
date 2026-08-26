package app

import "net/http"

type moduleSummary struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Nature      string `json:"nature"`
	Description string `json:"description"`
}

type dashboardMetrics struct {
	Clients      int `json:"clients"`
	Executions   int `json:"executions"`
	Failures     int `json:"failures"`
	Certificates int `json:"certificates"`
}

type dashboardResponse struct {
	Modules []moduleSummary  `json:"modules"`
	Metrics dashboardMetrics `json:"metrics"`
}

var availableModules = []moduleSummary{
	{
		Code:        "SC-02",
		Name:        "Situação fiscal dos clientes",
		Nature:      "RPA",
		Description: "Consulta órgãos em lote e diferencia irregularidade de falha operacional.",
	},
	{
		Code:        "SC-05",
		Name:        "Bloqueio de inadimplentes",
		Nature:      "RPA",
		Description: "Orquestra bloqueio e desbloqueio em sistemas isolados.",
	},
	{
		Code:        "SC-06",
		Name:        "Briefing societário",
		Nature:      "Controle",
		Description: "Formulário dinâmico com regras condicionais configuráveis.",
	},
	{
		Code:        "SC-20",
		Name:        "Certificados digitais",
		Nature:      "Controle",
		Description: "Monitora vencimentos e evita alertas repetidos.",
	},
}

func (a *App) registerDashboardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/dashboard", a.requireAuthentication(a.dashboard))
	mux.HandleFunc("GET /api/executions", a.requireAuthentication(a.executions))
	mux.HandleFunc("GET /api/clients", a.requireAuthentication(a.clients))
}

func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := a.currentUser(r)
	databaseSnapshot := a.store.snapshot()

	response := dashboardResponse{
		Modules: modulesForUser(user),
		Metrics: dashboardMetrics{
			Clients:      len(databaseSnapshot.Clients),
			Executions:   len(databaseSnapshot.Executions),
			Failures:     countFailedExecutions(databaseSnapshot.Executions),
			Certificates: len(databaseSnapshot.Certificates),
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *App) executions(w http.ResponseWriter, r *http.Request) {
	user, _ := a.currentUser(r)
	databaseSnapshot := a.store.snapshot()
	visibleExecutions := make([]Execution, 0, len(databaseSnapshot.Executions))

	for index := len(databaseSnapshot.Executions) - 1; index >= 0; index-- {
		execution := databaseSnapshot.Executions[index]
		if user.Role == "admin" || sliceContains(user.Modules, execution.Module) {
			visibleExecutions = append(visibleExecutions, execution)
		}
	}

	writeJSON(w, http.StatusOK, visibleExecutions)
}

func (a *App) clients(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.store.snapshot().Clients)
}

func modulesForUser(user User) []moduleSummary {
	if user.Role == "admin" {
		return availableModules
	}

	modules := make([]moduleSummary, 0, len(user.Modules))
	for _, module := range availableModules {
		if sliceContains(user.Modules, module.Code) {
			modules = append(modules, module)
		}
	}
	return modules
}

func countFailedExecutions(executions []Execution) int {
	count := 0
	for _, execution := range executions {
		if execution.Status == "failed" {
			count++
		}
	}
	return count
}
