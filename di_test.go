package ginji

import (
	"fmt"
	"testing"
)

// Test services
type ILogger interface {
	Log(message string)
}

type simpleLogger struct {
	messages []string
}

func (l *simpleLogger) Log(message string) {
	l.messages = append(l.messages, message)
}

type UserService struct {
	logger ILogger
}

func NewUserService(logger ILogger) *UserService {
	return &UserService{logger: logger}
}

func (s *UserService) CreateUser(name string) string {
	s.logger.Log(fmt.Sprintf("Creating user: %s", name))
	return fmt.Sprintf("User created: %s", name)
}

// Test repository
type Repository struct {
	connectionString string
}

func NewRepository() *Repository {
	return &Repository{connectionString: "mongodb://localhost"}
}

func TestRegisterSingleton(t *testing.T) {
	container := NewContainer()

	// Register singleton
	err := container.RegisterSingleton("logger", func() ILogger {
		return &simpleLogger{messages: make([]string, 0)}
	})

	if err != nil {
		t.Fatalf("Failed to register singleton: %v", err)
	}

	// Resolve twice - should be the same instance
	instance1, err := container.Resolve("logger", nil)
	if err != nil {
		t.Fatalf("Failed to resolve logger: %v", err)
	}

	instance2, err := container.Resolve("logger", nil)
	if err != nil {
		t.Fatalf("Failed to resolve logger second time: %v", err)
	}

	logger1 := instance1.(ILogger)

	// Verify it's the same instance
	logger1.Log("test")
	simpleLogger1 := instance1.(*simpleLogger)
	simpleLogger2 := instance2.(*simpleLogger)

	if len(simpleLogger1.messages) != 1 || len(simpleLogger2.messages) != 1 {
		t.Error("Singleton instances are not the same")
	}
}

func TestRegisterTransient(t *testing.T) {
	container := NewContainer()

	// Register transient
	err := container.RegisterTransient("repo", func() *Repository {
		return NewRepository()
	})

	if err != nil {
		t.Fatalf("Failed to register transient: %v", err)
	}

	// Resolve twice - should be different instances
	instance1, err := container.Resolve("repo", nil)
	if err != nil {
		t.Fatalf("Failed to resolve repo: %v", err)
	}

	instance2, err := container.Resolve("repo", nil)
	if err != nil {
		t.Fatalf("Failed to resolve repo second time: %v", err)
	}

	repo1 := instance1.(*Repository)
	repo2 := instance2.(*Repository)

	// Modify first instance
	repo1.connectionString = "modified"

	// Second instance should not be affected
	if repo2.connectionString == "modified" {
		t.Error("Transient instances should be different")
	}
}

func TestRegisterScoped(t *testing.T) {
	container := NewContainer()

	// Register scoped service
	err := container.RegisterScoped("logger", func() ILogger {
		return &simpleLogger{messages: make([]string, 0)}
	})

	if err != nil {
		t.Fatalf("Failed to register scoped service: %v", err)
	}

	// Create first scope
	scope1 := NewServiceScope(container, nil)
	instance1a, err := container.Resolve("logger", scope1)
	if err != nil {
		t.Fatalf("Failed to resolve logger in scope1: %v", err)
	}

	instance1b, err := container.Resolve("logger", scope1)
	if err != nil {
		t.Fatalf("Failed to resolve logger second time in scope1: %v", err)
	}

	// Within same scope, should be same instance
	logger1a := instance1a.(*simpleLogger)
	logger1b := instance1b.(*simpleLogger)

	logger1a.Log("test")
	if len(logger1b.messages) != 1 {
		t.Error("Scoped instances within same scope should be the same")
	}

	// Create second scope
	scope2 := NewServiceScope(container, nil)
	instance2, err := container.Resolve("logger", scope2)
	if err != nil {
		t.Fatalf("Failed to resolve logger in scope2: %v", err)
	}

	logger2 := instance2.(*simpleLogger)

	// Different scope should have different instance
	if len(logger2.messages) != 0 {
		t.Error("Different scopes should have different instances")
	}
}

func TestConstructorInjection(t *testing.T) {
	container := NewContainer()

	// Register logger with the correct type name that DI will look for
	if err := container.RegisterSingleton("ginji.ILogger", func() ILogger {
		return &simpleLogger{messages: make([]string, 0)}
	}); err != nil {
		t.Fatalf("Failed to register logger: %v", err)
	}

	// Register user service with dependency injection
	if err := container.RegisterTransient("*ginji.UserService", func(logger ILogger) *UserService {
		return NewUserService(logger)
	}); err != nil {
		t.Fatalf("Failed to register user service: %v", err)
	}

	// Resolve user service
	instance, err := container.Resolve("*ginji.UserService", nil)
	if err != nil {
		t.Fatalf("Failed to resolve UserService: %v", err)
	}

	userService := instance.(*UserService)
	result := userService.CreateUser("John")

	if result != "User created: John" {
		t.Errorf("Expected 'User created: John', got '%s'", result)
	}

	// Verify logger was injected
	logger := userService.logger.(*simpleLogger)
	if len(logger.messages) != 1 {
		t.Error("ILogger was not properly injected")
	}
}

func TestRegisterInstance(t *testing.T) {
	container := NewContainer()

	// Create and register an existing instance
	logger := &simpleLogger{messages: []string{"initial"}}
	err := container.RegisterInstance("logger", logger)

	if err != nil {
		t.Fatalf("Failed to register instance: %v", err)
	}

	// Resolve - should get the same instance
	instance, err := container.Resolve("logger", nil)
	if err != nil {
		t.Fatalf("Failed to resolve logger: %v", err)
	}

	resolvedILogger := instance.(*simpleLogger)
	if len(resolvedILogger.messages) != 1 || resolvedILogger.messages[0] != "initial" {
		t.Error("Registered instance was not properly resolved")
	}

	// Should be exact same instance
	resolvedILogger.Log("new")
	if len(logger.messages) != 2 {
		t.Error("Instance should be the same object")
	}
}

func TestGetServiceGeneric(t *testing.T) {
	container := NewContainer()

	// Register a service
	if err := container.RegisterSingleton("logger", func() ILogger {
		return &simpleLogger{messages: make([]string, 0)}
	}); err != nil {
		t.Fatalf("Failed to register logger: %v", err)
	}

	// Use generic GetService
	logger, err := GetService[ILogger](container, "logger", nil)
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	logger.Log("test")
	simpleLogger := logger.(*simpleLogger)
	if len(simpleLogger.messages) != 1 {
		t.Error("Generic GetService did not work correctly")
	}
}

func TestServiceNotFound(t *testing.T) {
	container := NewContainer()

	_, err := container.Resolve("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for nonexistent service")
	}
}

func TestEngineServiceIntegration(t *testing.T) {
	app := New()

	// Register services on the engine
	if err := app.RegisterSingleton("logger", func() ILogger {
		return &simpleLogger{messages: make([]string, 0)}
	}); err != nil {
		t.Fatalf("Failed to register logger: %v", err)
	}

	if err := app.RegisterScoped("*ginji.UserService", func(logger ILogger) *UserService {
		return NewUserService(logger)
	}); err != nil {
		t.Fatalf("Failed to register user service: %v", err)
	}

	// Verify container was initialized
	if app.Container() == nil {
		t.Fatal("Engine container should not be nil")
	}

	// Test resolution
	logger, err := app.Container().Resolve("logger", nil)
	if err != nil {
		t.Fatalf("Failed to resolve logger from engine: %v", err)
	}

	if logger == nil {
		t.Error("ILogger should not be nil")
	}
}

func TestContextServiceIntegration(t *testing.T) {
	app := New()

	// Register services
	if err := app.RegisterSingleton("logger", func() ILogger {
		return &simpleLogger{messages: make([]string, 0)}
	}); err != nil {
		t.Fatalf("Failed to register logger: %v", err)
	}

	// Test route with service injection
	app.Get("/test", func(c *Context) error {
		logger, err := GetServiceTyped[ILogger](c, "logger")
		if err != nil {
			c.AbortWithError(StatusInternalServerError, err)
		}

		logger.Log("Request processed")
		return c.Text(StatusOK, "OK")
	})

	// Perform request
	req, rec := NewTestContextWithRecorder("GET", "/test")
	app.ServeHTTP(rec, req.Req)

	if rec.Code != StatusOK {
		t.Errorf("Expected status %d, got %d", StatusOK, rec.Code)
	}

	// Verify logger was used
	logger, _ := app.Container().Resolve("logger", nil)
	simpleLogger := logger.(*simpleLogger)
	if len(simpleLogger.messages) != 1 {
		t.Error("ILogger was not used in request handler")
	}
}
