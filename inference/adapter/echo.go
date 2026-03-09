package adapter

// EchoAdapter returns the original query and is intended for tests.
type EchoAdapter struct{}

// NewEchoAdapter creates an echo adapter.
func NewEchoAdapter() *EchoAdapter {
	return &EchoAdapter{}
}

// Infer returns the original query without modification.
func (a *EchoAdapter) Infer(query string) (string, error) {
	return query, nil
}

// ModelID returns the adapter identifier.
func (a *EchoAdapter) ModelID() string {
	return KindEcho
}

// IsReady always reports ready for the echo adapter.
func (a *EchoAdapter) IsReady() bool {
	return true
}
