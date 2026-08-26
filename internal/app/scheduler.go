package app

import "time"

const schedulerInterval = 6 * time.Hour

func (a *App) runScheduler() {
	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()

	for range ticker.C {
		a.ensureMonthlySC02()
		a.ensureMonthlySC20()
	}
}

func (a *App) ensureMonthlySC20() {
	now := time.Now().UTC()
	databaseSnapshot := a.store.snapshot()

	if alreadyExecutedThisMonth(databaseSnapshot.Executions, now) {
		return
	}

	expiringCount := countCertificatesWithinDays(
		databaseSnapshot.Certificates,
		now,
		60,
	)
	execution := newCertificateScanExecution(
		"Agendador",
		"scheduled",
		now,
		expiringCount,
	)

	_ = a.store.update(func(db *database) error {
		db.Executions = append(db.Executions, execution)
		return nil
	})
}

func alreadyExecutedThisMonth(executions []Execution, reference time.Time) bool {
	for _, execution := range executions {
		isCurrentMonth := execution.StartedAt.Year() == reference.Year() &&
			execution.StartedAt.Month() == reference.Month()

		if execution.Module == "SC-20" && execution.Trigger == "scheduled" && isCurrentMonth {
			return true
		}
	}
	return false
}
