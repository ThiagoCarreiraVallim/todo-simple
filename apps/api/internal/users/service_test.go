package users

import "testing"

func TestNormalizeUsername(t *testing.T) {
	cases := []struct {
		in    string
		want  string
		valid bool
	}{
		{"  Joao ", "joao", true}, // trim + lowercase
		{"MARIA_99", "maria_99", true},
		{"ab", "", false}, // curto demais
		{"a", "", false},  // curto demais
		{"comespaço no meio", "", false},
		{"tres", "tres", true},
		{"averylongusername_123", "", false}, // > 20
		{"joão", "", false},                  // acento fora do padrão
		{"user-name", "user-name", true},
	}
	for _, c := range cases {
		got, ok := normalizeUsername(c.in)
		if ok != c.valid || (ok && got != c.want) {
			t.Errorf("normalizeUsername(%q) = (%q, %v), want (%q, %v)", c.in, got, ok, c.want, c.valid)
		}
	}
}
