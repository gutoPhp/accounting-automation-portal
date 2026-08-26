package app

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSeedAuthentication(t *testing.T) {
	s, err := NewStore(t.TempDir() + "/db.json")
	if err != nil {
		t.Fatal(err)
	}
	u, ok := s.authenticate("admin@sheepcontabil.com", "Sheep@2026")
	if !ok || u.Role != "admin" {
		t.Fatal("admin deveria autenticar")
	}
	if _, ok := s.authenticate("admin@sheepcontabil.com", "errada"); ok {
		t.Fatal("senha inválida foi aceita")
	}
}
func TestOperatorCannotAccessSC05(t *testing.T) {
	a, err := New(t.TempDir() + "/db.json")
	if err != nil {
		t.Fatal(err)
	}
	loginBody := strings.NewReader(`{
		"email": "operador@sheepcontabil.com",
		"password": "Sheep@2026"
	}`)
	req := httptest.NewRequest("POST", "/api/login", loginBody)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	cookie := w.Result().Cookies()[0]
	runBody := strings.NewReader(`{
		"client_id": "cli-1",
		"action": "block"
	}`)
	req = httptest.NewRequest("POST", "/api/sc-05/run", runBody)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("esperava 403, recebeu %d", w.Code)
	}
}

func TestRepeatedBlockPreservesOriginalTaskOwner(t *testing.T) {
	client := Client{
		ID:           "cli-test",
		TaskOwner:    "Marina",
		DefaultOwner: "Marina",
	}

	blockClient(&client)
	blockClient(&client)
	unblockClient(&client)

	if client.TaskOwner != "Marina" {
		t.Fatalf("esperava restaurar Marina, recebeu %q", client.TaskOwner)
	}
}

func TestFiscalScanKeepsFailedConsultationVisible(t *testing.T) {
	clients := seedClients()
	checks := buildFiscalChecks(clients, time.Now().UTC(), true)

	expectedChecks := len(clients) * len(fiscalAgencies)
	if len(checks) != expectedChecks {
		t.Fatalf("esperava %d consultas, recebeu %d", expectedChecks, len(checks))
	}
	if countFiscalFailures(checks) != 1 {
		t.Fatal("esperava uma falha registrada")
	}
}
