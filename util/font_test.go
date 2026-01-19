package util

import (
	"testing"
)

func TestGetFace(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"basicfont"},
		{"inconsolata-bold"},
		{"inconsolata-regular"},
		{"goregular"},
		{"gobold"},
		{"gobolditalic"},
		{"goitalic"},
		{"gomedium"},
		{"gomediumitalic"},
		{"gomono"},
		{"gomonobold"},
		{"gomonobolditalic"},
		{"gomonoitalic"},
		{"gosmallcaps"},
		{"gosmallcapsitalic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			face, err := GetFace(tt.name, 12, 72)
			if err != nil {
				t.Errorf("GetFace(%q) error = %v", tt.name, err)
				return
			}
			if face == nil {
				t.Errorf("GetFace(%q) returned nil face", tt.name)
			}
		})
	}
}
