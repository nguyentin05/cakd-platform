package llm

// LLM defines the interface for AI integrations that provide diagnosis capabilities.
// Implementations of this interface interact with Large Language Models (LLMs) like Gemini.
type LLM interface {
	// Analyze sends system logs and metrics context to the LLM and returns a troubleshooting diagnosis.
	Analyze(context string) (diagnosis string, err error)
}
