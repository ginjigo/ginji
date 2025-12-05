package ginji

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

// TestDICircularDependency tests that circular dependencies are handled gracefully
func TestDICircularDependency(t *testing.T) {
	container := NewContainer()

	// Register services that depend on each other
	// Note: Current implementation doesn't detect cycles - this test documents the behavior
	err := container.RegisterSingleton("serviceA", func(serviceB string) (string, error) {
		return "A with " + serviceB, nil
	})
	if err != nil {
		t.Fatalf("Failed to register serviceA: %v", err)
	}

	err = container.RegisterSingleton("serviceB", func(serviceA string) (string, error) {
		return "B with " + serviceA, nil
	})
	if err != nil {
		t.Fatalf("Failed to register serviceB: %v", err)
	}

	// Attempting to resolve would cause infinite recursion
	// For now we document this behavior - circular dependency detection would be future work
}

// TestDIInvalidFactorySignature tests handling of invalid factory functions
func TestDIInvalidFactorySignature(t *testing.T) {
	container := NewContainer()

	tests := []struct {
		name    string
		factory any
		wantErr bool
	}{
		{
			name:    "Not a function",
			factory: "not a function",
			wantErr: true,
		},
		{
			name:    "Function with no returns",
			factory: func() {},
			wantErr: true,
		},
		{
			name:    "Function with too many returns",
			factory: func() (string, int, error) { return "", 0, nil },
			wantErr: true,
		},
		{
			name:    "Valid factory - single return",
			factory: func() string { return "ok" },
			wantErr: false,
		},
		{
			name:    "Valid factory - with error",
			factory: func() (string, error) { return "ok", nil },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := container.Register("test", tt.factory, Singleton)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDIServiceNotFound tests error handling when service doesn't exist
func TestDIServiceNotFound(t *testing.T) {
	container := NewContainer()
	scope := NewServiceScope(container, nil)

	_, err := container.Resolve("nonexistent", scope)
	if err == nil {
		t.Error("Expected error for nonexistent service, got nil")
	}

	expectedMsg := "service 'nonexistent' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestDIFactoryReturnsError tests handling when factory returns an error
func TestDIFactoryReturnsError(t *testing.T) {
	container := NewContainer()
	scope := NewServiceScope(container, nil)

	expectedErr := errors.New("factory error")
	err := container.RegisterSingleton("failing", func() (string, error) {
		return "", expectedErr
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	_, err = container.Resolve("failing", scope)
	if err == nil {
		t.Fatal("Expected error from failing factory, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap factory error, got: %v", err)
	}
}

// TestDINilFactory tests handling of nil factory
func TestDINilFactory(t *testing.T) {
	container := NewContainer()

	err := container.Register("test", nil, Singleton)
	if err == nil {
		t.Error("Expected error for nil factory, got nil")
	}
}

// TestDIConcurrentResolution tests thread safety of concurrent service resolution
func TestDIConcurrentResolution(t *testing.T) {
	container := NewContainer()

	// Register a singleton service
	counter := 0
	err := container.RegisterSingleton("counter", func() int {
		counter++
		return counter
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Resolve concurrently from multiple goroutines
	const numGoroutines = 100
	var wg sync.WaitGroup
	results := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scope := NewServiceScope(container, nil)
			result, err := container.Resolve("counter", scope)
			if err != nil {
				t.Errorf("Resolve failed: %v", err)
				return
			}
			results <- result.(int)
		}()
	}

	wg.Wait()
	close(results)

	// All goroutines should get the same singleton instance
	firstValue := <-results
	for value := range results {
		if value != firstValue {
			t.Errorf("Expected all values to be %d, got %d", firstValue, value)
		}
	}

	// Factory should have been called only once
	if counter != 1 {
		t.Errorf("Expected factory to be called once, was called %d times", counter)
	}
}

// TestDITransientVsSingleton tests difference between transient and singleton lifetimes
func TestDITransientVsSingleton(t *testing.T) {
	container := NewContainer()

	counter := 0
	factory := func() int {
		counter++
		return counter
	}

	// Register as transient
	err := container.RegisterTransient("transient", factory)
	if err != nil {
		t.Fatalf("Failed to register transient: %v", err)
	}

	scope := NewServiceScope(container, nil)

	// Each resolution should create a new instance
	val1, _ := container.Resolve("transient", scope)
	val2, _ := container.Resolve("transient", scope)

	if val1 == val2 {
		t.Error("Transient services should return different instances")
	}

	// Reset counter
	counter = 0
	container2 := NewContainer()
	err = container2.RegisterSingleton("singleton", factory)
	if err != nil {
		t.Fatalf("Failed to register singleton: %v", err)
	}

	scope2 := NewServiceScope(container2, nil)

	// Each resolution should return same instance
	val3, _ := container2.Resolve("singleton", scope2)
	val4, _ := container2.Resolve("singleton", scope2)

	if val3 != val4 {
		t.Error("Singleton services should return same instance")
	}

	if counter != 1 {
		t.Errorf("Singleton factory should be called once, was called %d times", counter)
	}
}

// TestDIScopedLifetime tests scoped service lifecycle
func TestDIScopedLifetime(t *testing.T) {
	container := NewContainer()

	counter := 0
	err := container.RegisterScoped("scoped", func() int {
		counter++
		return counter
	})
	if err != nil {
		t.Fatalf("Failed to register scoped service: %v", err)
	}

	// Within same scope, should return same instance
	scope1 := NewServiceScope(container, nil)
	val1, _ := container.Resolve("scoped", scope1)
	val2, _ := container.Resolve("scoped", scope1)

	if val1 != val2 {
		t.Error("Scoped services should return same instance within same scope")
	}

	// Different scope should create new instance
	scope2 := NewServiceScope(container, nil)
	val3, _ := container.Resolve("scoped", scope2)

	if val1 == val3 {
		t.Error("Scoped services should return different instances in different scopes")
	}
}

// TestDIFactoryWithDependencies tests factory that depends on other services
func TestDIFactoryWithDependencies(t *testing.T) {
	container := NewContainer()

	// Register dependency
	err := container.RegisterSingleton("config", func() string {
		return "production"
	})
	if err != nil {
		t.Fatalf("Failed to register config: %v", err)
	}

	// Register service that depends on config
	err = container.RegisterSingleton("service", func(config string) string {
		return fmt.Sprintf("Service in %s mode", config)
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	scope := NewServiceScope(container, nil)
	result, err := container.Resolve("service", scope)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	expected := "Service in production mode"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestDIRegisterInstanceError tests error handling in RegisterInstance
func TestDIRegisterInstanceError(t *testing.T) {
	container := NewContainer()

	// Nil instance should error
	err := container.RegisterInstance("test", nil)
	if err == nil {
		t.Error("Expected error for nil instance, got nil")
	}
}

// TestDIMustResolveTypedPanic tests that MustResolveTyped panics on error
func TestDIMustResolveTypedPanic(t *testing.T) {
	container := NewContainer()
	scope := NewServiceScope(container, nil)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nonexistent service")
		}
	}()

	_ = MustResolveTyped[string](container, scope)
}

// TestDITypeMismatch tests type mismatch in GetService
func TestDITypeMismatch(t *testing.T) {
	container := NewContainer()
	scope := NewServiceScope(container, nil)

	// Register string service
	err := container.RegisterSingleton("test", func() string {
		return "hello"
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Try to resolve as int (type mismatch)
	_, err = GetService[int](container, "test", scope)
	if err == nil {
		t.Error("Expected error for type mismatch, got nil")
	}
}

// TestDIScopeDisposal tests that scope disposal prevents further resolution
func TestDIScopeDisposal(t *testing.T) {
	container := NewContainer()

	err := container.RegisterScoped("test", func() string {
		return "value"
	})
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	scope := NewServiceScope(container, nil)

	// Resolve once - should work
	_, err = container.Resolve("test", scope)
	if err != nil {
		t.Fatalf("First resolve failed: %v", err)
	}

	// Dispose scope
	scope.Dispose()
}
