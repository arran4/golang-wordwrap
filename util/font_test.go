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
	tests := []string{
		"goregular",
		"basicfont",
		"inconsolata",
		"inconsolata-bold",
		"gomono",
	}

	for _, name := range tests {
		face, err := GetFace(name, 12, 72)
		if err != nil {
			t.Errorf("GetFace(%q) returned error: %v", name, err)
		}
		if face == nil {
			t.Errorf("GetFace(%q) returned nil face", name)
		}
	}

	// Test invalid font
	_, err := GetFace("invalid", 12, 72)
	if err == nil {
		t.Error("GetFace(\"invalid\") expected error, got nil")
	}
}
