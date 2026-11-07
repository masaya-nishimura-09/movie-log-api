package record

import "testing"

func TestCheckAttachmentLimit(t *testing.T) {
	tests := []struct {
		name    string
		input   uint
		wantErr bool
	}{
		{"no attachments", 0, false},
		{"one below the limit", 9, false},

		{"at the limit", 10, true},
		{"over the limit", 11, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAttachmentLimit(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAttachmentLimit(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
