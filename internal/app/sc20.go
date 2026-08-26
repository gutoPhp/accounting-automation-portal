package app

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type notifyCertificateRequest struct {
	ID        string `json:"id"`
	Recipient string `json:"recipient"`
}

type scheduleStatus struct {
	Enabled    bool       `json:"enabled"`
	Frequency  string     `json:"frequency"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`
	NextRunAt  time.Time  `json:"next_run_at"`
	LastStatus string     `json:"last_status,omitempty"`
}

func (a *App) registerSC20Routes(mux *http.ServeMux) {
	mux.HandleFunc(
		"GET /api/sc-20/certificates",
		a.requireModule("SC-20", a.certificates),
	)
	mux.HandleFunc(
		"POST /api/sc-20/run",
		a.requireModule("SC-20", a.runSC20),
	)
	mux.HandleFunc(
		"GET /api/sc-20/schedule",
		a.requireModule("SC-20", a.sc20Schedule),
	)
	mux.HandleFunc(
		"POST /api/sc-20/notify",
		a.requireModule("SC-20", a.notifyCertificate),
	)
}

func (a *App) sc20Schedule(w http.ResponseWriter, _ *http.Request) {
	now := time.Now().UTC()
	status := scheduleStatus{
		Enabled:   true,
		Frequency: "Mensal",
		NextRunAt: time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC),
	}

	executions := a.store.snapshot().Executions
	for index := len(executions) - 1; index >= 0; index-- {
		execution := executions[index]
		if execution.Module == "SC-20" && execution.Trigger == "scheduled" {
			status.LastRunAt = &execution.FinishedAt
			status.LastStatus = execution.Status
			break
		}
	}

	writeJSON(w, http.StatusOK, status)
}

func (a *App) certificates(w http.ResponseWriter, _ *http.Request) {
	databaseSnapshot := a.store.snapshot()

	sort.Slice(databaseSnapshot.Certificates, func(i, j int) bool {
		return databaseSnapshot.Certificates[i].ExpiresAt.Before(
			databaseSnapshot.Certificates[j].ExpiresAt,
		)
	})

	writeJSON(w, http.StatusOK, databaseSnapshot.Certificates)
}

func (a *App) runSC20(w http.ResponseWriter, r *http.Request) {
	user, _ := a.currentUser(r)
	startedAt := time.Now().UTC()
	expiringCount := countCertificatesWithinDays(
		a.store.snapshot().Certificates,
		startedAt,
		60,
	)

	execution := newCertificateScanExecution(
		user.Name,
		"manual",
		startedAt,
		expiringCount,
	)

	_ = a.store.update(func(db *database) error {
		db.Executions = append(db.Executions, execution)
		return nil
	})

	writeJSON(w, http.StatusOK, execution)
}

func (a *App) notifyCertificate(w http.ResponseWriter, r *http.Request) {
	user, _ := a.currentUser(r)

	var input notifyCertificateRequest
	if err := decodeJSON(w, r, &input); err != nil || strings.TrimSpace(input.Recipient) == "" {
		writeError(w, http.StatusBadRequest, "Informe o certificado e o destinatário.")
		return
	}

	certificate, err := a.registerCertificateNotification(input, user.Name)
	if err != nil {
		writeError(w, http.StatusNotFound, readableError(err))
		return
	}

	writeJSON(w, http.StatusOK, certificate)
}

func (a *App) registerCertificateNotification(
	input notifyCertificateRequest,
	triggeredBy string,
) (Certificate, error) {
	var notifiedCertificate Certificate

	err := a.store.update(func(db *database) error {
		for index := range db.Certificates {
			certificate := &db.Certificates[index]
			if certificate.ID != input.ID {
				continue
			}

			now := time.Now().UTC()
			certificate.LastNotifiedAt = &now
			certificate.LastNotifiedTo = input.Recipient
			notifiedCertificate = *certificate

			db.Executions = append(db.Executions, Execution{
				ID:          newID("exec"),
				Module:      "SC-20",
				StartedAt:   now,
				FinishedAt:  now,
				TriggeredBy: triggeredBy,
				Trigger:     "manual",
				Status:      "success",
				Summary:     "Aviso registrado para " + certificate.ClientName + ".",
			})

			return nil
		}
		return errNotFound
	})

	return notifiedCertificate, err
}

func countCertificatesWithinDays(
	certificates []Certificate,
	reference time.Time,
	days int,
) int {
	threshold := reference.AddDate(0, 0, days)
	count := 0

	for _, certificate := range certificates {
		if !certificate.ExpiresAt.After(threshold) {
			count++
		}
	}
	return count
}

func newCertificateScanExecution(
	triggeredBy string,
	trigger string,
	startedAt time.Time,
	expiringCount int,
) Execution {
	label := "Varredura concluída"
	if trigger == "scheduled" {
		label = "Varredura mensal concluída"
	}

	return Execution{
		ID:          newID("exec"),
		Module:      "SC-20",
		StartedAt:   startedAt,
		FinishedAt:  executionFinishedAt(startedAt),
		TriggeredBy: triggeredBy,
		Trigger:     trigger,
		Status:      "success",
		Summary:     certificateScanSummary(label, expiringCount),
	}
}

func certificateScanSummary(label string, expiringCount int) string {
	return fmt.Sprintf(
		"%s: %d certificado(s) vencem em até 60 dias.",
		label,
		expiringCount,
	)
}
