package llm

// LLM provides AI analysis capabilities.
type LLM interface {
	Analyze(context string) (diagnosis string, err error)
}
