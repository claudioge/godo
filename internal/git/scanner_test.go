package git

import (
	"reflect"
	"testing"
)

func TestParseAIResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     []FileChange
		wantErr  bool
	}{
		{
			name: "Standard format",
			response: "Here is the code:\n\n```\nFILE: main.go\nCONTENT:\npackage main\n\nfunc main() {}\n```",
			want: []FileChange{
				{Path: "main.go", Content: "package main\n\nfunc main() {}", Action: "create"},
			},
			wantErr: false,
		},
		{
			name: "Multiple files",
			response: "```\nFILE: a.go\nCONTENT:\npackage a\n```\n\n```\nFILE: b.go\nCONTENT:\npackage b\n```",
			want: []FileChange{
				{Path: "a.go", Content: "package a", Action: "create"},
				{Path: "b.go", Content: "package b", Action: "create"},
			},
			wantErr: false,
		},
		{
			name: "No changes needed",
			response: "NO_CHANGES_NEEDED",
			want:     nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAIResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAIResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAIResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
