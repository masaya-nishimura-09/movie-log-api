package record

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestNewNote(t *testing.T) {
	ofSize := func(size int) json.RawMessage {
		return json.RawMessage(`{"type":"doc","a":"` + strings.Repeat("a", size-21) + `"}`)
	}

	tests := []struct {
		name    string
		input   json.RawMessage
		want    Note
		wantErr bool
	}{
		{"typical", json.RawMessage(`{"type":"doc","content":[{"type":"paragraph"}]}`), Note(`{"type":"doc","content":[{"type":"paragraph"}]}`), false},
		{"max size", ofSize(NoteMaxBytes), Note(ofSize(NoteMaxBytes)), false},
		{"whitespace is removed", json.RawMessage("{\n  \"type\": \"doc\",\n  \"content\": [{\"type\": \"paragraph\"}]\n}"), Note(`{"type":"doc","content":[{"type":"paragraph"}]}`), false},
		{"whitespace does not count toward max size", json.RawMessage(" " + string(ofSize(NoteMaxBytes)) + " "), Note(ofSize(NoteMaxBytes)), false},

		{"empty", nil, nil, true},
		{"invalid json", json.RawMessage(`{"type":`), nil, true},
		{"not json", json.RawMessage(`hello`), nil, true},
		{"over max size", ofSize(NoteMaxBytes + 1), nil, true},
		{"array", json.RawMessage(`[1,2]`), nil, true},
		{"string", json.RawMessage(`"doc"`), nil, true},
		{"null", json.RawMessage(`null`), nil, true},
		{"object without type", json.RawMessage(`{}`), nil, true},
		{"object with other type", json.RawMessage(`{"type":"paragraph"}`), nil, true},
		{"content not array", json.RawMessage(`{"type":"doc","content":"abc"}`), nil, true},
		{"child not object", json.RawMessage(`{"type":"doc","content":[1]}`), nil, true},
		{"child type not string", json.RawMessage(`{"type":"doc","content":[{"type":1}]}`), nil, true},
		{"text not string", json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":5}]}]}`), nil, true},
		{"nested content not array", json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","content":{}}]}`), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewNote(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewNote(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("NewNote(len=%d) = %s, want %s", len(tt.input), got, tt.want)
			}
		})
	}
}
