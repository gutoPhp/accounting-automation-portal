package app

import "time"

type User struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	PasswordHash string   `json:"password_hash"`
	Role         string   `json:"role"`
	Modules      []string `json:"modules"`
}

type Client struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CNPJ          string `json:"cnpj"`
	Delinquent    bool   `json:"delinquent"`
	Blocked       bool   `json:"blocked"`
	TaskOwner     string `json:"task_owner"`
	DefaultOwner  string `json:"default_owner"`
	PreviousOwner string `json:"previous_owner,omitempty"`
}

type SystemAction struct {
	System string `json:"system"`
	Action string `json:"action"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type Execution struct {
	ID          string         `json:"id"`
	Module      string         `json:"module"`
	StartedAt   time.Time      `json:"started_at"`
	FinishedAt  time.Time      `json:"finished_at"`
	TriggeredBy string         `json:"triggered_by"`
	Trigger     string         `json:"trigger"`
	Status      string         `json:"status"`
	Summary     string         `json:"summary"`
	Error       string         `json:"error,omitempty"`
	Actions     []SystemAction `json:"actions,omitempty"`
}

type BriefingRule struct {
	ID       string   `json:"id"`
	Field    string   `json:"field"`
	Operator string   `json:"operator"`
	Value    string   `json:"value"`
	Requires []string `json:"requires"`
	Label    string   `json:"label"`
}

type Briefing struct {
	ID        string            `json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	CreatedBy string            `json:"created_by"`
	Status    string            `json:"status"`
	Answers   map[string]string `json:"answers"`
}

type Certificate struct {
	ID             string     `json:"id"`
	ClientID       string     `json:"client_id"`
	ClientName     string     `json:"client_name"`
	Kind           string     `json:"kind"`
	ExpiresAt      time.Time  `json:"expires_at"`
	LastNotifiedAt *time.Time `json:"last_notified_at,omitempty"`
	LastNotifiedTo string     `json:"last_notified_to,omitempty"`
}

type FiscalCheck struct {
	ID         string    `json:"id"`
	ClientID   string    `json:"client_id"`
	ClientName string    `json:"client_name"`
	Agency     string    `json:"agency"`
	Status     string    `json:"status"`
	CheckedAt  time.Time `json:"checked_at"`
	Error      string    `json:"error,omitempty"`
	Attempts   int       `json:"attempts"`
}

type database struct {
	Users        []User         `json:"users"`
	Clients      []Client       `json:"clients"`
	Executions   []Execution    `json:"executions"`
	Rules        []BriefingRule `json:"rules"`
	Briefings    []Briefing     `json:"briefings"`
	Certificates []Certificate  `json:"certificates"`
	FiscalChecks []FiscalCheck  `json:"fiscal_checks"`
}
