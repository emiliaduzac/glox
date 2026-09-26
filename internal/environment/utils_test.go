package environment

import (
	"strings"
	"testing"
)

func SetUpEnvironment(t *testing.T) *Environment {
	t.Helper()
	return NewEnvironment()
}

func AssertValue(t *testing.T, got any, expected any) {
	t.Helper()
	if got != expected {
		t.Errorf("got %v, expected %v", got, expected)
	}
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func AssertErrorMessage(t *testing.T, err error, expected string) {
	t.Helper()
	AssertError(t, err)
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %q", expected, err.Error())
	}
}
