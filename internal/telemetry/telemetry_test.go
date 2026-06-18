package telemetry

import (
	"testing"
)

func TestInitTelemetry_DoesNotPanic(t *testing.T) {
	shutdown, err := Init("rssembly-test", "0.1.0")
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected non-nil shutdown function")
	}
	shutdown()
}

func TestInitTelemetry_ShutdownIdempotent(t *testing.T) {
	shutdown, err := Init("rssembly-test", "0.1.0")
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}
	shutdown()
	shutdown() // second call must not panic
}

func TestMetricsHandler_IsRegistered(t *testing.T) {
	shutdown, err := Init("rssembly-test", "0.1.0")
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}
	defer shutdown()

	if MetricsHandler() == nil {
		t.Fatal("expected non-nil MetricsHandler")
	}
}