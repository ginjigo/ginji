package ginji

import (
	"net/http/httptest"
	"testing"
)

// Test service for DI tests
type TestUserService struct {
	Name string
}

func NewTestUserService() *TestUserService {
	return &TestUserService{Name: "TestService"}
}

func TestRegisterTyped(t *testing.T) {
	container := NewContainer()

	// Register service with type safety
	err := RegisterSingletonTyped[*TestUserService](container, NewTestUserService)
	if err != nil {
		t.Fatalf("Failed to register typed service: %v", err)
	}

	// Resolve service
	service, err := ResolveTyped[*TestUserService](container, nil)
	if err != nil {
		t.Fatalf("Failed to resolve typed service: %v", err)
	}

	if service.Name != "TestService" {
		t.Errorf("Expected service name 'TestService', got '%s'", service.Name)
	}
}

func TestMustResolveTyped(t *testing.T) {
	container := NewContainer()

	// Register service
	err := RegisterSingletonTyped[*TestUserService](container, NewTestUserService)
	if err != nil {
		t.Fatalf("Failed to register typed service: %v", err)
	}

	// Must resolve (should not panic)
	service := MustResolveTyped[*TestUserService](container, nil)
	if service.Name != "TestService" {
		t.Errorf("Expected service name 'TestService', got '%s'", service.Name)
	}
}

func TestMustResolveTypedPanic(t *testing.T) {
	container := NewContainer()

	// Test panic on non-existent service
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for non-existent service")
		}
	}()

	_ = MustResolveTyped[*TestUserService](container, nil)
}

func TestRegisterInstanceTyped(t *testing.T) {
	container := NewContainer()

	// Create and register instance
	instance := &TestUserService{Name: "InstanceService"}
	err := RegisterInstanceTyped(container, instance)
	if err != nil {
		t.Fatalf("Failed to register instance: %v", err)
	}

	// Resolve and verify it's the same instance
	resolved, err := ResolveTyped[*TestUserService](container, nil)
	if err != nil {
		t.Fatalf("Failed to resolve instance: %v", err)
	}

	if resolved != instance {
		t.Error("Expected same instance")
	}

	if resolved.Name != "InstanceService" {
		t.Errorf("Expected service name 'InstanceService', got '%s'", resolved.Name)
	}
}

func TestRegisterScopedTyped(t *testing.T) {
	engine := New()
	container := engine.Container()

	// Register scoped service
	err := RegisterScopedTyped[*TestUserService](container, NewTestUserService)
	if err != nil {
		t.Fatalf("Failed to register scoped service: %v", err)
	}

	// Create request context and scope
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	ctx := NewContext(rec, req, engine)

	// Resolve from scope
	service1, err := ResolveTyped[*TestUserService](container, ctx.services)
	if err != nil {
		t.Fatalf("Failed to resolve scoped service: %v", err)
	}

	// Resolve again from same scope - should be same instance
	service2, err := ResolveTyped[*TestUserService](container, ctx.services)
	if err != nil {
		t.Fatalf("Failed to resolve scoped service second time: %v", err)
	}

	if service1 != service2 {
		t.Error("Expected same instance within same scope")
	}

	// Create new scope - should get different instance
	ctx2 := NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), engine)
	service3, err := ResolveTyped[*TestUserService](container, ctx2.services)
	if err != nil {
		t.Fatalf("Failed to resolve scoped service in new scope: %v", err)
	}

	if service1 == service3 {
		t.Error("Expected different instances in different scopes")
	}
}

func TestRegisterTransientTyped(t *testing.T) {
	container := NewContainer()

	// Register transient service
	err := RegisterTransientTyped[*TestUserService](container, NewTestUserService)
	if err != nil {
		t.Fatalf("Failed to register transient service: %v", err)
	}

	// Resolve twice - should get different instances
	service1, err := ResolveTyped[*TestUserService](container, nil)
	if err != nil {
		t.Fatalf("Failed to resolve transient service: %v", err)
	}

	service2, err := ResolveTyped[*TestUserService](container, nil)
	if err != nil {
		t.Fatalf("Failed to resolve transient service second time: %v", err)
	}

	if service1 == service2 {
		t.Error("Expected different instances for transient service")
	}
}
