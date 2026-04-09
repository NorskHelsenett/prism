package webauthn

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"prism/config"
	"prism/database"
	"prism/session"
)

var (
	webAuthn *webauthn.WebAuthn

	// Pending registration sessions (short-lived, keyed by email)
	pendingRegistrations   = make(map[string]*webauthn.SessionData)
	pendingRegistrationsMu sync.RWMutex

	// Pending authentication sessions (short-lived, keyed by email)
	pendingAuthentications   = make(map[string]*webauthn.SessionData)
	pendingAuthenticationsMu sync.RWMutex
)

func Init() {
	origin := config.AppConfig.Cors.Origin
	u, err := url.Parse(origin)
	if err != nil {
		log.Fatalf("Failed to parse CORS origin for WebAuthn: %v", err)
	}

	rpID := u.Hostname()
	rpOrigins := []string{origin}

	wconfig := &webauthn.Config{
		RPDisplayName: "PRISM",
		RPID:          rpID,
		RPOrigins:     rpOrigins,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.CrossPlatform,
			UserVerification:        protocol.VerificationPreferred,
			ResidentKey:             protocol.ResidentKeyRequirementDiscouraged,
		},
		AttestationPreference: protocol.PreferNoAttestation,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    60 * time.Second,
				TimeoutUVD: 60 * time.Second,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    60 * time.Second,
				TimeoutUVD: 60 * time.Second,
			},
		},
	}

	webAuthn, err = webauthn.New(wconfig)
	if err != nil {
		log.Fatalf("Failed to initialize WebAuthn: %v", err)
	}

	// Cleanup expired pending sessions periodically
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleanupPendingSessions()
		}
	}()
}

func cleanupPendingSessions() {
	now := time.Now()

	pendingRegistrationsMu.Lock()
	for email, sess := range pendingRegistrations {
		if now.After(sess.Expires) {
			delete(pendingRegistrations, email)
		}
	}
	pendingRegistrationsMu.Unlock()

	pendingAuthenticationsMu.Lock()
	for email, sess := range pendingAuthentications {
		if now.After(sess.Expires) {
			delete(pendingAuthentications, email)
		}
	}
	pendingAuthenticationsMu.Unlock()
}

// webAuthnUser implements webauthn.User interface
type webAuthnUser struct {
	email       string
	displayName string
	id          []byte
	credentials []webauthn.Credential
}

func (u *webAuthnUser) WebAuthnID() []byte                         { return u.id }
func (u *webAuthnUser) WebAuthnName() string                       { return u.email }
func (u *webAuthnUser) WebAuthnDisplayName() string                { return u.displayName }
func (u *webAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

func loadWebAuthnUser(email string) (*webAuthnUser, error) {
	user, err := database.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	creds, err := database.GetWebAuthnCredentials(email)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	var webauthnCreds []webauthn.Credential
	for _, c := range creds {
		var cred webauthn.Credential
		if err := json.Unmarshal(c.CredentialData, &cred); err != nil {
			log.Printf("Failed to unmarshal credential for %s: %v", email, err)
			continue
		}
		webauthnCreds = append(webauthnCreds, cred)
	}

	// Generate a stable user ID from email
	userID := generateUserID(email)

	return &webAuthnUser{
		email:       email,
		displayName: user.Name,
		id:          userID,
		credentials: webauthnCreds,
	}, nil
}

func generateUserID(email string) []byte {
	// Use a deterministic ID based on email hash
	h := make([]byte, 32)
	copy(h, []byte(email))
	return h
}

// BeginRegistration starts the passkey registration ceremony.
// Requires an authenticated session (user must be logged in via SSO).
func BeginRegistration(c *gin.Context) {
	email, ok := emailFromContext(c)
	if !ok {
		return
	}

	user, err := loadWebAuthnUser(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user data"})
		return
	}

	// Exclude existing credentials to prevent re-registration
	excludeList := make([]protocol.CredentialDescriptor, len(user.credentials))
	for i, cred := range user.credentials {
		excludeList[i] = protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.ID,
			Transport:    cred.Transport,
		}
	}

	options, sessionData, err := webAuthn.BeginRegistration(
		user,
		webauthn.WithExclusions(excludeList),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin registration"})
		log.Printf("WebAuthn BeginRegistration error: %v", err)
		return
	}

	pendingRegistrationsMu.Lock()
	pendingRegistrations[email] = sessionData
	pendingRegistrationsMu.Unlock()

	c.JSON(http.StatusOK, options)
}

// FinishRegistration completes the passkey registration ceremony.
func FinishRegistration(c *gin.Context) {
	email, ok := emailFromContext(c)
	if !ok {
		return
	}

	pendingRegistrationsMu.RLock()
	sessionData, exists := pendingRegistrations[email]
	pendingRegistrationsMu.RUnlock()

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No pending registration"})
		return
	}

	user, err := loadWebAuthnUser(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user data"})
		return
	}

	credential, err := webAuthn.FinishRegistration(user, *sessionData, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to verify registration"})
		log.Printf("WebAuthn FinishRegistration error: %v", err)
		return
	}

	// Serialize credential for storage
	credJSON, err := json.Marshal(credential)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store credential"})
		return
	}

	// Generate a friendly name
	credIDStr := base64.RawURLEncoding.EncodeToString(credential.ID)
	name := fmt.Sprintf("Passkey %s", credIDStr[:8])

	if err := database.SaveWebAuthnCredential(email, name, credJSON); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store credential"})
		log.Printf("Failed to save WebAuthn credential: %v", err)
		return
	}

	// Cleanup pending session
	pendingRegistrationsMu.Lock()
	delete(pendingRegistrations, email)
	pendingRegistrationsMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "Passkey registered successfully"})
}

// BeginAuthentication starts the passkey authentication ceremony for 2FA.
// The user must already be authenticated via SSO (cookie session exists).
func BeginAuthentication(c *gin.Context) {
	email, ok := emailFromContext(c)
	if !ok {
		return
	}

	user, err := loadWebAuthnUser(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user data"})
		return
	}

	if len(user.credentials) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No passkeys registered"})
		return
	}

	options, sessionData, err := webAuthn.BeginLogin(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin authentication"})
		log.Printf("WebAuthn BeginLogin error: %v", err)
		return
	}

	pendingAuthenticationsMu.Lock()
	pendingAuthentications[email] = sessionData
	pendingAuthenticationsMu.Unlock()

	c.JSON(http.StatusOK, options)
}

// FinishAuthentication completes the passkey 2FA verification and marks the session as OTP verified.
func FinishAuthentication(c *gin.Context, store *session.SessionStore) {
	email, ok := emailFromContext(c)
	if !ok {
		return
	}

	sessionIDInterface, exists := c.Get("sessionID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sessionID, ok := sessionIDInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
		return
	}

	pendingAuthenticationsMu.RLock()
	sessionData, exists := pendingAuthentications[email]
	pendingAuthenticationsMu.RUnlock()

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No pending authentication"})
		return
	}

	user, err := loadWebAuthnUser(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user data"})
		return
	}

	_, err = webAuthn.FinishLogin(user, *sessionData, c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Passkey verification failed"})
		log.Printf("WebAuthn FinishLogin error: %v", err)
		return
	}

	// Cleanup pending session
	pendingAuthenticationsMu.Lock()
	delete(pendingAuthentications, email)
	pendingAuthenticationsMu.Unlock()

	// Mark the session as OTP/2FA verified (same flag as TOTP)
	if err := store.PersistSession(email, sessionID, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Passkey verified successfully"})
}

// GetCredentials returns the user's registered passkeys.
func GetCredentials(c *gin.Context) {
	email, ok := emailFromContext(c)
	if !ok {
		return
	}

	creds, err := database.GetWebAuthnCredentials(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load credentials"})
		return
	}

	type credentialInfo struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"createdAt"`
	}

	result := make([]credentialInfo, len(creds))
	for i, c := range creds {
		result[i] = credentialInfo{
			ID:        c.ID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, result)
}

// DeleteCredential removes a passkey registration.
func DeleteCredential(c *gin.Context) {
	email, ok := emailFromContext(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing credential ID"})
		return
	}

	if err := database.DeleteWebAuthnCredential(email, idStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete credential"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Passkey deleted"})
}

// HasPasskeys returns whether a user has any registered passkeys.
func HasPasskeys(c *gin.Context) {
	email, ok := emailFromContext(c)
	if !ok {
		return
	}

	has, err := database.HasWebAuthnCredentials(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hasPasskeys": has})
}

// ResetPasskeysForUser allows admin to reset a user's passkeys.
func ResetPasskeysForUser(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing email"})
		return
	}

	if err := database.DeleteAllWebAuthnCredentials(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset passkeys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Passkeys reset successfully"})
}

// GetMFAStatusForUser returns the MFA status for a specific user (admin only).
func GetMFAStatusForUser(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing email"})
		return
	}

	hasOTP, err := database.CheckForOtpEnabled(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check OTP status"})
		return
	}

	passkeyCount, err := database.CountWebAuthnCredentials(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count passkeys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasOTP":       hasOTP,
		"passkeyCount": passkeyCount,
	})
}

func emailFromContext(c *gin.Context) (string, bool) {
	emailInterface, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", false
	}
	email, ok := emailInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return "", false
	}
	return email, true
}
