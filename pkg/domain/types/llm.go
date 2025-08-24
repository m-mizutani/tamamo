package types

// LLMProvider represents the type of LLM provider
type LLMProvider string

const (
	// LLMProviderOpenAI represents OpenAI provider
	LLMProviderOpenAI LLMProvider = "openai"
	// LLMProviderClaude represents Claude/Anthropic provider
	LLMProviderClaude LLMProvider = "claude"
	// LLMProviderGemini represents Google Gemini provider
	LLMProviderGemini LLMProvider = "gemini"
)

// String returns the string representation of the provider
func (p LLMProvider) String() string {
	return string(p)
}

// IsValid checks if the provider is valid
func (p LLMProvider) IsValid() bool {
	switch p {
	case LLMProviderOpenAI, LLMProviderClaude, LLMProviderGemini:
		return true
	default:
		return false
	}
}

// ToUpperCase returns the uppercase version of the provider
func (p LLMProvider) ToUpperCase() string {
	switch p {
	case LLMProviderOpenAI:
		return "OPENAI"
	case LLMProviderClaude:
		return "CLAUDE"
	case LLMProviderGemini:
		return "GEMINI"
	default:
		return string(p)
	}
}

// LLMProviderFromString converts a string to LLMProvider
func LLMProviderFromString(s string) LLMProvider {
	switch s {
	case "openai", "OPENAI", "OpenAI":
		return LLMProviderOpenAI
	case "claude", "CLAUDE", "Claude":
		return LLMProviderClaude
	case "gemini", "GEMINI", "Gemini":
		return LLMProviderGemini
	default:
		return LLMProvider(s)
	}
}
