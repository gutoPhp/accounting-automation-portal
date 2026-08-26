package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	path string
	db   database
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if raw, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(raw, &s.db); err != nil {
			return nil, fmt.Errorf("ler base: %w", err)
		}
		s.normalizeClients()
		s.normalizeUsers()
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else {
		s.db = seedData()
		if err := s.saveLocked(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) normalizeUsers() {
	for index := range s.db.Users {
		user := &s.db.Users[index]
		if user.Role == "admin" && !sliceContains(user.Modules, "SC-02") {
			user.Modules = append([]string{"SC-02"}, user.Modules...)
		}
	}
}

func (s *Store) normalizeClients() {
	defaultOwners := map[string]string{
		"cli-1": "Marina",
		"cli-2": "Carlos",
		"cli-3": "Lívia",
	}

	for index := range s.db.Clients {
		client := &s.db.Clients[index]
		if client.DefaultOwner == "" {
			client.DefaultOwner = defaultOwners[client.ID]
		}
		if client.PreviousOwner == "BLOQUEADO" {
			client.PreviousOwner = client.DefaultOwner
		}
	}
}

func hashPassword(value string) string {
	sum := sha256.Sum256([]byte("sheepcontabil:" + value))
	return hex.EncodeToString(sum[:])
}

func seedData() database {
	now := time.Now().UTC()

	return database{
		Users:        seedUsers(),
		Clients:      seedClients(),
		Executions:   make([]Execution, 0),
		FiscalChecks: make([]FiscalCheck, 0),
		Rules:        seedBriefingRules(),
		Briefings:    make([]Briefing, 0),
		Certificates: seedCertificates(now),
	}
}

func seedUsers() []User {
	return []User{
		{
			ID:           "usr-admin",
			Name:         "Ana Administradora",
			Email:        "admin@sheepcontabil.com",
			PasswordHash: hashPassword("Sheep@2026"),
			Role:         "admin",
			Modules:      []string{"SC-02", "SC-05", "SC-06", "SC-20"},
		},
		{
			ID:           "usr-operador",
			Name:         "Bruno Operador",
			Email:        "operador@sheepcontabil.com",
			PasswordHash: hashPassword("Sheep@2026"),
			Role:         "operator",
			Modules:      []string{"SC-06", "SC-20"},
		},
	}
}

func seedClients() []Client {
	return []Client{
		{
			ID:           "cli-1",
			Name:         "Clínica Horizonte Ltda.",
			CNPJ:         "12.345.678/0001-95",
			Delinquent:   true,
			TaskOwner:    "Marina",
			DefaultOwner: "Marina",
		},
		{
			ID:           "cli-2",
			Name:         "Comercial Riacho Ltda.",
			CNPJ:         "48.721.036/0001-42",
			TaskOwner:    "Carlos",
			DefaultOwner: "Carlos",
		},
		{
			ID:            "cli-3",
			Name:          "Agro Vale Verde Ltda.",
			CNPJ:          "73.946.281/0001-10",
			Delinquent:    true,
			Blocked:       true,
			TaskOwner:     "BLOQUEADO",
			DefaultOwner:  "Lívia",
			PreviousOwner: "Lívia",
		},
	}
}

func seedBriefingRules() []BriefingRule {
	return []BriefingRule{
		{
			ID:       "rule-state",
			Field:    "state",
			Operator: "neq",
			Value:    "AL",
			Requires: []string{"foreign_state_registration", "foreign_state_city"},
			Label:    "Cliente de outro estado",
		},
		{
			ID:       "rule-married",
			Field:    "partner_marital_status",
			Operator: "eq",
			Value:    "casado",
			Requires: []string{"marriage_regime"},
			Label:    "Sócio casado",
		},
		{
			ID:       "rule-change",
			Field:    "service_type",
			Operator: "eq",
			Value:    "alteracao",
			Requires: []string{"change_description"},
			Label:    "Alteração contratual",
		},
	}
}

func seedCertificates(reference time.Time) []Certificate {
	return []Certificate{
		{
			ID:         "cert-1",
			ClientID:   "cli-1",
			ClientName: "Clínica Horizonte Ltda.",
			Kind:       "A1",
			ExpiresAt:  reference.AddDate(0, 0, 18),
		},
		{
			ID:         "cert-2",
			ClientID:   "cli-2",
			ClientName: "Comercial Riacho Ltda.",
			Kind:       "A3",
			ExpiresAt:  reference.AddDate(0, 0, 47),
		},
		{
			ID:         "cert-3",
			ClientID:   "cli-3",
			ClientName: "Agro Vale Verde Ltda.",
			Kind:       "A1",
			ExpiresAt:  reference.AddDate(0, 0, 94),
		},
	}
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s.db, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) authenticate(email, password string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.db.Users {
		if u.Email == email && u.PasswordHash == hashPassword(password) {
			u.PasswordHash = ""
			return u, true
		}
	}
	return User{}, false
}

func (s *Store) snapshot() database {
	s.mu.RLock()
	defer s.mu.RUnlock()
	raw, _ := json.Marshal(s.db)
	var out database
	_ = json.Unmarshal(raw, &out)
	return out
}

func (s *Store) update(fn func(*database) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(&s.db); err != nil {
		return err
	}
	return s.saveLocked()
}
