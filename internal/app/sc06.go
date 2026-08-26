package app

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

var baseBriefingFields = []string{
	"client_name",
	"service_type",
	"state",
	"partner_name",
	"partner_marital_status",
}

type createBriefingRequest struct {
	Answers map[string]string `json:"answers"`
}

func (a *App) registerSC06Routes(mux *http.ServeMux) {
	mux.HandleFunc(
		"GET /api/sc-06/rules",
		a.requireModule("SC-06", a.rules),
	)
	mux.HandleFunc(
		"GET /api/sc-06/briefings",
		a.requireModule("SC-06", a.briefings),
	)
	mux.HandleFunc(
		"POST /api/sc-06/briefings",
		a.requireModule("SC-06", a.createBriefing),
	)
}

func (a *App) rules(w http.ResponseWriter, _ *http.Request) {
	rules := a.store.snapshot().Rules
	if rules == nil {
		rules = make([]BriefingRule, 0)
	}

	writeJSON(w, http.StatusOK, rules)
}

func (a *App) briefings(w http.ResponseWriter, _ *http.Request) {
	databaseSnapshot := a.store.snapshot()
	if databaseSnapshot.Briefings == nil {
		databaseSnapshot.Briefings = make([]Briefing, 0)
	}

	sort.Slice(databaseSnapshot.Briefings, func(i, j int) bool {
		return databaseSnapshot.Briefings[i].CreatedAt.After(
			databaseSnapshot.Briefings[j].CreatedAt,
		)
	})

	writeJSON(w, http.StatusOK, databaseSnapshot.Briefings)
}

func (a *App) createBriefing(w http.ResponseWriter, r *http.Request) {
	user, _ := a.currentUser(r)

	var input createBriefingRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Preencha o briefing.")
		return
	}

	rules := a.store.snapshot().Rules
	missing := requiredBriefingFields(input.Answers, rules)
	if len(missing) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":   "O briefing ainda possui campos obrigatórios.",
			"missing": missing,
		})
		return
	}

	briefing := newBriefing(user.Name, input.Answers)
	if err := a.saveBriefing(briefing); err != nil {
		writeError(w, http.StatusInternalServerError, "Não foi possível salvar o briefing.")
		return
	}

	writeJSON(w, http.StatusCreated, briefing)
}

func (a *App) saveBriefing(briefing Briefing) error {
	return a.store.update(func(db *database) error {
		db.Briefings = append(db.Briefings, briefing)
		db.Executions = append(db.Executions, Execution{
			ID:          newID("exec"),
			Module:      "SC-06",
			StartedAt:   briefing.CreatedAt,
			FinishedAt:  briefing.CreatedAt,
			TriggeredBy: briefing.CreatedBy,
			Trigger:     "manual",
			Status:      "success",
			Summary:     "Briefing validado e concluído.",
		})
		return nil
	})
}

func newBriefing(createdBy string, answers map[string]string) Briefing {
	return Briefing{
		ID:        newID("brief"),
		CreatedAt: time.Now().UTC(),
		CreatedBy: createdBy,
		Status:    "complete",
		Answers:   answers,
	}
}

func requiredBriefingFields(answers map[string]string, rules []BriefingRule) []string {
	missing := missingFields(answers, baseBriefingFields)

	for _, rule := range rules {
		if ruleMatches(answers[rule.Field], rule.Operator, rule.Value) {
			missing = append(missing, missingFields(answers, rule.Requires)...)
		}
	}

	return uniqueStrings(missing)
}

func ruleMatches(actual, operator, expected string) bool {
	switch operator {
	case "eq":
		return strings.EqualFold(actual, expected)
	case "neq":
		return actual != "" && !strings.EqualFold(actual, expected)
	default:
		return false
	}
}

func missingFields(values map[string]string, fields []string) []string {
	missing := make([]string, 0)
	for _, field := range fields {
		if strings.TrimSpace(values[field]) == "" {
			missing = append(missing, field)
		}
	}
	return missing
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	unique := make([]string, 0, len(values))

	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			unique = append(unique, value)
		}
	}
	return unique
}
