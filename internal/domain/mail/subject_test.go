package mail

import "testing"

func TestNewSubject(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Subject
		wantErr bool
	}{
		{"typical", "メールアドレスの確認", "メールアドレスの確認", false},

		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSubject(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSubject(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewSubject(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
