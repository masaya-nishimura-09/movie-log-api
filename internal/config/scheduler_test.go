package config

import (
	"testing"
	"time"
)

func TestCleanupEnable(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    bool
		wantErr bool
	}{
		{name: "returns an error when CLEANUP_ENABLE is empty", value: "", wantErr: true},
		{name: "returns an error when CLEANUP_ENABLE is not a boolean", value: "abc", wantErr: true},
		{name: "returns true when CLEANUP_ENABLE is true", value: "true", want: true},
		{name: "returns false when CLEANUP_ENABLE is false", value: "false", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLEANUP_ENABLE", tt.value)

			got, err := CleanupEnable()
			if (err != nil) != tt.wantErr {
				t.Fatalf("CleanupEnable() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CleanupEnable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanupInterval(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "returns an error when CLEANUP_INTERVAL_SECOND is empty", value: "", wantErr: true},
		{name: "returns an error when CLEANUP_INTERVAL_SECOND is not a number", value: "abc", wantErr: true},
		{name: "returns an error when CLEANUP_INTERVAL_SECOND is zero", value: "0", wantErr: true},
		{name: "returns an error when CLEANUP_INTERVAL_SECOND is negative", value: "-1", wantErr: true},
		{name: "returns the seconds as a duration when CLEANUP_INTERVAL_SECOND is positive", value: "60", want: 60 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLEANUP_INTERVAL_SECOND", tt.value)

			got, err := CleanupInterval()
			if (err != nil) != tt.wantErr {
				t.Fatalf("CleanupInterval() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CleanupInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
