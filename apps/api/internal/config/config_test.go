package config

import (
	"net/url"
	"testing"
)

func TestDatabaseURLFromPartsEscapesPassword(t *testing.T) {
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "todo")
	t.Setenv("DB_PASSWORD", "hhKmKDZ6pLsl+4izeuNzGUtN/Dy9Lf/Q")
	t.Setenv("DB_NAME", "todo")
	t.Setenv("DB_SSLMODE", "disable")

	raw := databaseURLFromParts()

	// A URL montada precisa ser parseável e devolver a senha original.
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("URL inválida: %v (%q)", err, raw)
	}
	if u.Hostname() != "postgres" || u.Port() != "5432" {
		t.Errorf("host/porta errados: %q:%q", u.Hostname(), u.Port())
	}
	pass, _ := u.User.Password()
	if pass != "hhKmKDZ6pLsl+4izeuNzGUtN/Dy9Lf/Q" {
		t.Errorf("senha não sobreviveu ao round-trip: %q", pass)
	}
	if u.Query().Get("sslmode") != "disable" {
		t.Errorf("sslmode errado: %q", u.Query().Get("sslmode"))
	}
}

func TestDatabaseURLFromPartsEmptyWhenIncomplete(t *testing.T) {
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_NAME", "todo")

	if got := databaseURLFromParts(); got != "" {
		t.Errorf("esperava string vazia com partes faltando, veio %q", got)
	}
}
