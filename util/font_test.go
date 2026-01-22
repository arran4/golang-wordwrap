package util

import (
	"testing"
)

func TestFontByName(t *testing.T) {
	tests := []string{
		"goregular",
		"gobold",
		"gobolditalic",
		"goitalic",
		"gomedium",
		"gomediumitalic",
		"gomono",
		"gomonobold",
		"gomonobolditalic",
		"gomonoitalic",
		"gosmallcaps",
		"gosmallcapsitalic",
	}

	for _, name := range tests {
		b, err := FontByName(name)
		if err != nil {
			t.Errorf("FontByName(%q) returned error: %v", name, err)
		}
		if len(b) == 0 {
			t.Errorf("FontByName(%q) returned empty bytes", name)
		}
	}

	// Test invalid font
	_, err := FontByName("invalid")
	if err == nil {
		t.Error("FontByName(\"invalid\") expected error, got nil")
	}
}

func TestGetFace(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"basicfont"},
		{"inconsolata-bold"},
		{"inconsolata-regular"},
		{"inconsolata"},
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

	// Test invalid font
	_, err := GetFace("invalid", 12, 72)
	if err == nil {
		t.Error("GetFace(\"invalid\") expected error, got nil")
	}
}
