package ginji

import "fmt"

// Plugin defines the interface that all plugins must implement.
type Plugin interface {
	// Name returns the unique name of the plugin.
	Name() string

	// Version returns the plugin version.
	Version() string

	// Install is called when the plugin is registered with the engine.
	// This is where plugins can register routes, middleware, hooks, etc.
	Install(engine *Engine) error

	// Start is called when the engine starts (before Listen).
	// This is where plugins can start background tasks, connect to databases, etc.
	Start() error

	// Stop is called when the engine shuts down.
	// This is where plugins should clean up resources.
	Stop() error
}

// PluginConfig holds configuration for a plugin.
type PluginConfig struct {
	Enabled  bool
	Priority int // Lower numbers run first
	Config   map[string]any
}

// PluginMetadata holds metadata about a registered plugin.
type PluginMetadata struct {
	Plugin    Plugin
	Config    *PluginConfig
	Installed bool
	Started   bool
}

// PluginRegistry manages registered plugins.
type PluginRegistry struct {
	plugins map[string]*PluginMetadata
	order   []string // Plugin names in execution order
}

// newPluginRegistry creates a new plugin registry.
func newPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]*PluginMetadata),
		order:   make([]string, 0),
	}
}

// Register registers a plugin with the engine.
func (pr *PluginRegistry) Register(plugin Plugin, config *PluginConfig) error {
	name := plugin.Name()

	if _, exists := pr.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	if config == nil {
		config = &PluginConfig{
			Enabled:  true,
			Priority: 100,
			Config:   make(map[string]any),
		}
	}

	pr.plugins[name] = &PluginMetadata{
		Plugin:    plugin,
		Config:    config,
		Installed: false,
		Started:   false,
	}

	// Insert in priority order
	pr.insertInOrder(name, config.Priority)

	return nil
}

// insertInOrder maintains plugins in priority order.
func (pr *PluginRegistry) insertInOrder(name string, priority int) {
	// Find insertion point
	pos := len(pr.order)
	for i, pName := range pr.order {
		if pr.plugins[pName].Config.Priority > priority {
			pos = i
			break
		}
	}

	// Insert at position
	pr.order = append(pr.order[:pos], append([]string{name}, pr.order[pos:]...)...)
}

// Install installs all registered plugins.
func (pr *PluginRegistry) Install(engine *Engine) error {
	for _, name := range pr.order {
		meta := pr.plugins[name]

		if !meta.Config.Enabled {
			continue
		}

		if err := meta.Plugin.Install(engine); err != nil {
			return fmt.Errorf("failed to install plugin '%s': %w", name, err)
		}

		meta.Installed = true
	}

	return nil
}

// Start starts all installed plugins.
func (pr *PluginRegistry) Start() error {
	for _, name := range pr.order {
		meta := pr.plugins[name]

		if !meta.Config.Enabled || !meta.Installed {
			continue
		}

		if err := meta.Plugin.Start(); err != nil {
			return fmt.Errorf("failed to start plugin '%s': %w", name, err)
		}

		meta.Started = true
	}

	return nil
}

// Stop stops all running plugins.
func (pr *PluginRegistry) Stop() error {
	// Stop in reverse order
	for i := len(pr.order) - 1; i >= 0; i-- {
		name := pr.order[i]
		meta := pr.plugins[name]

		if !meta.Started {
			continue
		}

		if err := meta.Plugin.Stop(); err != nil {
			return fmt.Errorf("failed to stop plugin '%s': %w", name, err)
		}

		meta.Started = false
	}

	return nil
}

// Get retrieves a plugin by name.
func (pr *PluginRegistry) Get(name string) (Plugin, bool) {
	if meta, exists := pr.plugins[name]; exists {
		return meta.Plugin, true
	}
	return nil, false
}

// List returns all registered plugin names.
func (pr *PluginRegistry) List() []string {
	return append([]string{}, pr.order...)
}

// UsePlugin registers a plugin with the engine.
func (e *Engine) UsePlugin(plugin Plugin, config ...*PluginConfig) *Engine {
	var cfg *PluginConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if err := e.plugins.Register(plugin, cfg); err != nil {
		// In production, you might want to handle this differently
		panic(err)
	}

	return e
}

// InstallPlugins installs all registered plugins.
func (e *Engine) InstallPlugins() error {
	return e.plugins.Install(e)
}

// StartPlugins starts all installed plugins.
func (e *Engine) StartPlugins() error {
	return e.plugins.Start()
}

// StopPlugins stops all running plugins.
func (e *Engine) StopPlugins() error {
	return e.plugins.Stop()
}
