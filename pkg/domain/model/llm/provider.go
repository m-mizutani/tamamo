package llm

// Provider represents an LLM provider configuration
type Provider struct {
	ID          string  `yaml:"-" json:"id"`
	DisplayName string  `yaml:"display_name" json:"display_name"`
	Models      []Model `yaml:"models" json:"models"`
}

// Model represents an LLM model configuration
type Model struct {
	ID          string `yaml:"id" json:"id"`
	DisplayName string `yaml:"display_name" json:"display_name"`
	Description string `yaml:"description" json:"description"`
}

// ProvidersConfig represents the complete LLM providers configuration
type ProvidersConfig struct {
	Providers map[string]Provider `yaml:"providers"`
	Defaults  DefaultConfig       `yaml:"defaults"`
	Fallback  FallbackConfig      `yaml:"fallback"`
}

// DefaultConfig represents default provider and model settings
type DefaultConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
}

// FallbackConfig represents fallback settings when primary provider fails
type FallbackConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
}

// ValidateProviderModel checks if a provider and model combination is valid
func (c *ProvidersConfig) ValidateProviderModel(provider, model string) bool {
	if provider == "" || model == "" {
		return false
	}

	p, exists := c.Providers[provider]
	if !exists {
		return false
	}

	for _, m := range p.Models {
		if m.ID == model {
			return true
		}
	}

	return false
}

// GetProvider returns a provider by ID
func (c *ProvidersConfig) GetProvider(id string) (*Provider, bool) {
	p, exists := c.Providers[id]
	if !exists {
		return nil, false
	}
	p.ID = id
	return &p, true
}

// GetModel returns a model by provider and model ID
func (c *ProvidersConfig) GetModel(provider, modelID string) (*Model, bool) {
	p, exists := c.Providers[provider]
	if !exists {
		return nil, false
	}

	for _, m := range p.Models {
		if m.ID == modelID {
			return &m, true
		}
	}

	return nil, false
}
