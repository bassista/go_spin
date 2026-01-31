package runtime

import (
	"context"
	"testing"

	"github.com/bassista/go_spin/internal/repository"
)

func TestNewRuntimeFromConfig_Memory(t *testing.T) {
	rt, err := NewRuntimeFromConfig(RuntimeTypeMemory, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rt == nil {
		t.Fatal("expected runtime to be created")
	}
	if _, ok := rt.(*MemoryRuntime); !ok {
		t.Error("expected MemoryRuntime type")
	}
}

func TestNewRuntimeFromConfig_MemoryWithDocument(t *testing.T) {
	doc := &repository.DataDocument{
		Containers: []repository.Container{
			{Name: "c1", Running: boolPtr(true)},
		},
	}

	rt, err := NewRuntimeFromConfig(RuntimeTypeMemory, doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mr, ok := rt.(*MemoryRuntime)
	if !ok {
		t.Fatal("expected MemoryRuntime type")
	}

	// Verify container state was loaded
	running, _ := mr.IsRunning(context.TODO(), "c1")
	if !running {
		t.Error("expected c1 to be running from document")
	}
}

func TestNewRuntimeFromConfig_Docker(t *testing.T) {
	// This test may fail if Docker is not available
	// We just check that it doesn't return an unknown runtime error
	_, err := NewRuntimeFromConfig(RuntimeTypeDocker, nil)
	// If Docker is not available, we expect an error, but not "unknown runtime type"
	if err != nil {
		if err.Error() == "unknown runtime type: docker (supported: docker, memory)" {
			t.Error("docker should be a recognized runtime type")
		}
		// Other errors (like Docker not available) are acceptable in test environment
		t.Logf("Docker runtime error (may be expected if Docker not running): %v", err)
	}
}

func TestNewRuntimeFromConfig_EmptyString(t *testing.T) {
	// Empty string should default to Docker
	_, err := NewRuntimeFromConfig("", nil)
	if err != nil {
		// If Docker is not available, we expect an error, but not "unknown runtime type"
		if err.Error() == "unknown runtime type:  (supported: docker, memory)" {
			t.Error("empty string should default to docker")
		}
		t.Logf("Docker runtime error (may be expected if Docker not running): %v", err)
	}
}

func TestNewRuntimeFromConfig_UnknownType(t *testing.T) {
	_, err := NewRuntimeFromConfig("unknown-runtime", nil)
	if err == nil {
		t.Error("expected error for unknown runtime type")
	}
}
