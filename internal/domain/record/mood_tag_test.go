package record

import (
	"slices"
	"testing"
)

func TestNewMoodTag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MoodTag
		wantErr bool
	}{
		{"defined value", "moving", MoodTagMoving, false},

		{"empty", "", "", true},
		{"undefined value", "sad", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMoodTag(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewMoodTag(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewMoodTag(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewMoodTags(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []MoodTag
		wantErr bool
	}{
		{"multiple values", []string{"moving", "dark"}, []MoodTag{MoodTagMoving, MoodTagDark}, false},
		{"empty slice", []string{}, []MoodTag{}, false},

		{"duplicate value", []string{"moving", "moving"}, nil, true},
		{"undefined element", []string{"moving", "sad"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMoodTags(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewMoodTags(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("NewMoodTags(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
