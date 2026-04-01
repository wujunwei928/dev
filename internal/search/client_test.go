package search

import (
	"testing"
)

func TestGetHTTPClient_NotNil(t *testing.T) {
	client := GetHTTPClient()
	if client == nil {
		t.Fatal("expected non-nil HTTP client")
	}
}

func TestGetHTTPClient_Singleton(t *testing.T) {
	client1 := GetHTTPClient()
	client2 := GetHTTPClient()
	if client1 != client2 {
		t.Error("expected same client instance (singleton)")
	}
}
