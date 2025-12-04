package ginji

import (
	"fmt"
	"reflect"
	"sync"
)

// ServiceLifetime defines the lifetime of a service.
type ServiceLifetime int

const (
	// Singleton creates a single instance shared across all requests.
	Singleton ServiceLifetime = iota
	// Transient creates a new instance for each resolution.
	Transient
	// Scoped creates one instance per request scope.
	Scoped
)

// ServiceDescriptor describes a registered service.
type ServiceDescriptor struct {
	Name     string
	Lifetime ServiceLifetime
	Factory  any // Factory function
	Instance any // For singleton instances
	Type     reflect.Type
}

// Container is the main DI container.
type Container struct {
	services map[string]*ServiceDescriptor
	mu       sync.RWMutex
}

// NewContainer creates a new DI container.
func NewContainer() *Container {
	return &Container{
		services: make(map[string]*ServiceDescriptor),
	}
}

// Register registers a service with the container.
// The factory must be a function that returns the service or (service, error).
func (c *Container) Register(name string, factory any, lifetime ServiceLifetime) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate factory is a function
	factoryVal := reflect.ValueOf(factory)
	if factoryVal.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function")
	}

	factoryType := factoryVal.Type()

	// Factory should return 1 or 2 values (service) or (service, error)
	if factoryType.NumOut() == 0 || factoryType.NumOut() > 2 {
		return fmt.Errorf("factory must return (service) or (service, error)")
	}

	// If 2 return values, second must be error
	if factoryType.NumOut() == 2 {
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if !factoryType.Out(1).Implements(errorType) {
			return fmt.Errorf("second return value must be error")
		}
	}

	serviceType := factoryType.Out(0)

	c.services[name] = &ServiceDescriptor{
		Name:     name,
		Lifetime: lifetime,
		Factory:  factory,
		Type:     serviceType,
	}

	return nil
}

// RegisterSingleton registers a singleton service.
func (c *Container) RegisterSingleton(name string, factory any) error {
	return c.Register(name, factory, Singleton)
}

// RegisterTransient registers a transient service.
func (c *Container) RegisterTransient(name string, factory any) error {
	return c.Register(name, factory, Transient)
}

// RegisterScoped registers a scoped service.
func (c *Container) RegisterScoped(name string, factory any) error {
	return c.Register(name, factory, Scoped)
}

// RegisterInstance registers a pre-created instance as a singleton.
func (c *Container) RegisterInstance(name string, instance any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[name] = &ServiceDescriptor{
		Name:     name,
		Lifetime: Singleton,
		Instance: instance,
		Type:     reflect.TypeOf(instance),
	}

	return nil
}

// Resolve resolves a service by name.
func (c *Container) Resolve(name string, scope *ServiceScope) (any, error) {
	c.mu.RLock()
	descriptor, ok := c.services[name]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("service '%s' not found", name)
	}

	return c.resolveDescriptor(descriptor, scope)
}

// MustResolve resolves a service or panics if not found.
func (c *Container) MustResolve(name string, scope *ServiceScope) any {
	instance, err := c.Resolve(name, scope)
	if err != nil {
		panic(err)
	}
	return instance
}

// resolveDescriptor resolves a service based on its descriptor.
func (c *Container) resolveDescriptor(descriptor *ServiceDescriptor, scope *ServiceScope) (any, error) {
	switch descriptor.Lifetime {
	case Singleton:
		return c.resolveSingleton(descriptor)

	case Scoped:
		if scope == nil {
			return nil, fmt.Errorf("scoped service '%s' requires a scope", descriptor.Name)
		}
		return scope.Resolve(descriptor.Name, descriptor)

	case Transient:
		return c.createInstance(descriptor, scope)

	default:
		return nil, fmt.Errorf("unknown service lifetime: %d", descriptor.Lifetime)
	}
}

// resolveSingleton resolves or creates a singleton instance.
func (c *Container) resolveSingleton(descriptor *ServiceDescriptor) (any, error) {
	// Check if instance already exists
	if descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	// Create instance
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring lock
	if descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	instance, err := c.createInstance(descriptor, nil)
	if err != nil {
		return nil, err
	}

	descriptor.Instance = instance
	return instance, nil
}

// createInstance creates a new instance using the factory.
func (c *Container) createInstance(descriptor *ServiceDescriptor, scope *ServiceScope) (any, error) {
	factoryVal := reflect.ValueOf(descriptor.Factory)
	factoryType := factoryVal.Type()

	// Prepare arguments for factory function
	args := make([]reflect.Value, factoryType.NumIn())
	for i := 0; i < factoryType.NumIn(); i++ {
		argType := factoryType.In(i)

		// Check if argument is *Container
		if argType == reflect.TypeOf((*Container)(nil)) {
			args[i] = reflect.ValueOf(c)
			continue
		}

		// Check if argument is *ServiceScope
		if argType == reflect.TypeOf((*ServiceScope)(nil)) {
			args[i] = reflect.ValueOf(scope)
			continue
		}

		// Check if argument is *Context
		if argType == reflect.TypeOf((*Context)(nil)) {
			if scope != nil && scope.Context != nil {
				args[i] = reflect.ValueOf(scope.Context)
				continue
			}
		}

		// Try to resolve dependency by type name
		typeName := argType.String()
		instance, err := c.Resolve(typeName, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency '%s' for service '%s': %w", typeName, descriptor.Name, err)
		}
		args[i] = reflect.ValueOf(instance)
	}

	// Call factory
	results := factoryVal.Call(args)

	// Check for error
	if len(results) == 2 {
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
	}

	return results[0].Interface(), nil
}

// GetService is a generic method to resolve a service with type safety.
func GetService[T any](c *Container, name string, scope *ServiceScope) (T, error) {
	var zero T
	instance, err := c.Resolve(name, scope)
	if err != nil {
		return zero, err
	}

	service, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("service '%s' is not of type %T", name, zero)
	}

	return service, nil
}

// MustGetService is like GetService but panics on error.
func MustGetService[T any](c *Container, name string, scope *ServiceScope) T {
	service, err := GetService[T](c, name, scope)
	if err != nil {
		panic(err)
	}
	return service
}

// RegisterTyped registers a service with automatic type inference.
// This provides compile-time type safety and eliminates string-based registration errors.
//
// Example:
//
//	engine.RegisterTyped(func() (*UserService, error) {
//	    return &UserService{}, nil
//	}, Singleton)
func RegisterTyped[T any](c *Container, factory any, lifetime ServiceLifetime) error {
	// Use the full type name as the service key
	var zero T
	typeName := reflect.TypeOf(&zero).Elem().String()
	return c.Register(typeName, factory, lifetime)
}

// RegisterSingletonTyped registers a singleton service with type inference.
func RegisterSingletonTyped[T any](c *Container, factory any) error {
	return RegisterTyped[T](c, factory, Singleton)
}

// RegisterTransientTyped registers a transient service with type inference.
func RegisterTransientTyped[T any](c *Container, factory any) error {
	return RegisterTyped[T](c, factory, Transient)
}

// RegisterScopedTyped registers a scoped service with type inference.
func RegisterScopedTyped[T any](c *Container, factory any) error {
	return RegisterTyped[T](c, factory, Scoped)
}

// RegisterInstanceTyped registers a pre-created instance with type inference.
func RegisterInstanceTyped[T any](c *Container, instance T) error {
	var zero T
	typeName := reflect.TypeOf(&zero).Elem().String()
	return c.RegisterInstance(typeName, instance)
}

// ResolveTyped resolves a service by type instead of by name.
// This provides compile-time type safety.
//
// Example:
//
//	service, err := ResolveTyped[*UserService](container, scope)
func ResolveTyped[T any](c *Container, scope *ServiceScope) (T, error) {
	var zero T
	typeName := reflect.TypeOf(&zero).Elem().String()
	return GetService[T](c, typeName, scope)
}

// MustResolveTyped is like ResolveTyped but panics on error.
func MustResolveTyped[T any](c *Container, scope *ServiceScope) T {
	service, err := ResolveTyped[T](c, scope)
	if err != nil {
		panic(err)
	}
	return service
}

// ServiceScope represents a request-scoped container.
type ServiceScope struct {
	container *Container
	instances map[string]any
	mu        sync.RWMutex
	Context   *Context // Reference to the request context
}

// NewServiceScope creates a new service scope.
func NewServiceScope(container *Container, ctx *Context) *ServiceScope {
	return &ServiceScope{
		container: container,
		instances: make(map[string]any),
		Context:   ctx,
	}
}

// Resolve resolves a scoped service.
func (s *ServiceScope) Resolve(name string, descriptor *ServiceDescriptor) (any, error) {
	s.mu.RLock()
	instance, ok := s.instances[name]
	s.mu.RUnlock()

	if ok {
		return instance, nil
	}

	// Create new instance
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring lock
	if instance, ok := s.instances[name]; ok {
		return instance, nil
	}

	instance, err := s.container.createInstance(descriptor, s)
	if err != nil {
		return nil, err
	}

	s.instances[name] = instance
	return instance, nil
}

// Dispose cleans up scoped services that implement io.Closer.
func (s *ServiceScope) Dispose() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, instance := range s.instances {
		// Check if service implements Disposable interface
		if disposable, ok := instance.(interface{ Dispose() error }); ok {
			_ = disposable.Dispose()
		}
	}

	s.instances = make(map[string]any)
}
