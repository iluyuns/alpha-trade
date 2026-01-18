package query

// WebauthnCredentialsCustom is the custom extension for WebauthnCredentials.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type WebauthnCredentialsCustom struct {
	*webauthnCredentialsDo
}

// NewWebauthnCredentials creates a new WebauthnCredentials data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewWebauthnCredentials(db Executor) *WebauthnCredentialsCustom {
	return &WebauthnCredentialsCustom{
		webauthnCredentialsDo: webauthnCredentials.WithDB(db).(*webauthnCredentialsDo),
	}
}
