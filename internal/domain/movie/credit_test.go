package movie

import (
	"slices"
	"testing"
)

func TestNewCreditRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CreditRole
		wantErr bool
	}{
		{"defined value", "director", CreditRoleDirector, false},

		{"empty", "", "", true},
		{"undefined value", "producer", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCreditRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCreditRole(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewCreditRole(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewCredits(t *testing.T) {
	credit := func(tmdbID *PersonTMDBID, name string, role CreditRole) Credit {
		return Credit{
			Person: Person{TMDBID: tmdbID, Name: PersonName(name)},
			Role:   role,
		}
	}
	id := func(v PersonTMDBID) *PersonTMDBID { return &v }

	tarantinoDirector := credit(id(138), "Quentin Tarantino", CreditRoleDirector)
	tarantinoWriter := credit(id(138), "Quentin Tarantino", CreditRoleWriter)
	avaryWriter := credit(id(1015), "Roger Avary", CreditRoleWriter)

	tests := []struct {
		name    string
		input   []Credit
		wantErr bool
	}{
		{"different people", []Credit{tarantinoWriter, avaryWriter}, false},
		{"same person different roles", []Credit{tarantinoDirector, tarantinoWriter}, false},
		{"empty slice", []Credit{}, false},
		{
			"same name different tmdb id",
			[]Credit{credit(id(1), "John Smith", CreditRoleCast), credit(id(2), "John Smith", CreditRoleCast)},
			false,
		},

		{
			"same tmdb id same role",
			[]Credit{tarantinoDirector, credit(id(138), "Quentin Tarantino", CreditRoleDirector)},
			true,
		},
		{
			"same name same role without tmdb id",
			[]Credit{credit(nil, "Quentin Tarantino", CreditRoleDirector), credit(nil, "Quentin Tarantino", CreditRoleDirector)},
			true,
		},
		{
			"same name same role with and without tmdb id",
			[]Credit{tarantinoDirector, credit(nil, "Quentin Tarantino", CreditRoleDirector)},
			true,
		},
		{"empty person name", []Credit{credit(id(138), "", CreditRoleDirector)}, true},
		{"undefined role", []Credit{credit(id(138), "Quentin Tarantino", "producer")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCredits(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCredits() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !slices.Equal(got, tt.input) {
				t.Errorf("NewCredits() = %v, want %v", got, tt.input)
			}
		})
	}
}
