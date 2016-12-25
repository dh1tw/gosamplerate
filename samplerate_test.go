package samplerate

import (
	"strings"
	"testing"
)

func TestGetConverterName(t *testing.T) {
	name, err := GetName(SRC_LINEAR)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Linear Interpolator" {
		t.Fatal("Unexpected string")
	}
}

func TestGetConverterNameError(t *testing.T) {
	_, err := GetName(5)
	if err == nil {
		t.Fatal("expected Error")
	}
	if err.Error() != "unknown samplerate converter" {
		t.Fatal("unexpected string")
	}
}

func TestGetConverterDescription(t *testing.T) {
	desc, err := GetDescription(SRC_LINEAR)
	if err != nil {
		t.Fatal(err)
	}
	if desc != "Linear interpolator, very fast, poor quality." {
		t.Fatal("Unexpected string")
	}
}

func TestGetConverterDescriptionError(t *testing.T) {
	_, err := GetDescription(5)
	if err == nil {
		t.Fatal("expected Error")
	}
	if err.Error() != "unknown samplerate converter" {
		t.Fatal("unexpected string")
	}
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if !strings.Contains(version, "libsamplerate-") {
		t.Fatal("Unexpected string")
	}
}
