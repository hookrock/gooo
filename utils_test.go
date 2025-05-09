package gooo

import (
	"os"
	"testing"
)

func TestResolveAddress(t *testing.T) {
	addr := resolveAddress([]string{"127.0.0.1:8080"})
	if addr != "127.0.0.1:8080" {
		t.Error("resolveAddress failed")
	}
}

func TestResolveAddressWithEnv(t *testing.T) {
	os.Setenv("PORT", "8080")
	addr := resolveAddress([]string{})
	if addr != ":8080" {
		t.Error("resolveAddress failed")
	}
}

func TestResolveAddressWithTooManyParams(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("resolveAddress failed")
		}
	}()
	resolveAddress([]string{"127.0.0.1:8080", "127.0.0.1:8081"})
}
