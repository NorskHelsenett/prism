package session

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"

	"prism/database"
)

type otp struct {
	secret string
	url    string
}

type pendingSecret struct {
	otp       *otp
	createdAt time.Time
}

var (
	pendingSecrets   = make(map[string]pendingSecret)
	pendingSecretsMu sync.RWMutex
	pendingSecretTTL = 10 * time.Minute
)

func getPendingSecret(email string) (*otp, bool) {
	pendingSecretsMu.RLock()
	entry, ok := pendingSecrets[email]
	pendingSecretsMu.RUnlock()
	if !ok {
		return nil, false
	}

	if time.Since(entry.createdAt) > pendingSecretTTL {
		pendingSecretsMu.Lock()
		delete(pendingSecrets, email)
		pendingSecretsMu.Unlock()
		return nil, false
	}

	return entry.otp, true
}

func setPendingSecret(email string, secret *otp) {
	pendingSecretsMu.Lock()
	pendingSecrets[email] = pendingSecret{otp: secret, createdAt: time.Now()}
	pendingSecretsMu.Unlock()
}

func clearPendingSecret(email string) {
	pendingSecretsMu.Lock()
	delete(pendingSecrets, email)
	pendingSecretsMu.Unlock()
}

func DeleteUserSession(c *gin.Context, session *SessionStore) {
	sessionID := c.Param("uuid")
	email, shouldReturn := emailFromContext(c)
	if shouldReturn {
		return
	}

	err := session.DeleteUserSessionsFor(email, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Session Deleted successfully"})
}

func GetUserSessions(c *gin.Context, session *SessionStore) {
	email, shouldReturn := emailFromContext(c)
	if shouldReturn {
		return
	}

	sessions, err := session.GetUserSessionsFor(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	sessionID, _ := c.Get("sessionID")

	sessionsArray := *sessions

	for i := range sessionsArray {
		if sessionsArray[i].SessionID == sessionID {
			sessionsArray[i].IsCurrent = true
		}
	}

	c.JSON(http.StatusOK, sessionsArray)
}

func HandleOTPResetForUser(c *gin.Context, session *SessionStore) {
	email := c.Param("email")

	// Persist the session with OTP verified status
	if err := session.InvalidateSession(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	err := database.DeleteOTPCode(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve OTP data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP reset successfully"})
}

func HandleOTPReset(c *gin.Context, session *SessionStore) {
	email, shouldReturn := emailFromContext(c)
	if shouldReturn {
		return
	}

	// Persist the session with OTP verified status
	if err := session.InvalidateSession(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	err := database.DeleteOTPCode(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve OTP data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP reset successfully"})
}

func emailFromContext(c *gin.Context) (string, bool) {
	emailInterface, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", true
	}

	email, ok := emailInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return "", true
	}
	return email, false
}

func HandleOTPValidate(c *gin.Context, session *SessionStore) {
	emailInterface, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionIdInterface, exists := c.Get("sessionID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	email, ok := emailInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	sessionID, ok := sessionIdInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var otpInput struct {
		Code string `json:"otp_code"`
	}

	// Parse the incoming JSON to otpInput
	if err := c.BindJSON(&otpInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	// Retrieve the OTP secret from the database for the user
	otpSecret, err := database.GetOTPCode(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve OTP data"})
		return
	}

	// Validate the OTP code
	if !totp.Validate(otpInput.Code, otpSecret) {
		c.JSON(406, gin.H{"error": "Invalid OTP code"})
		return
	}

	// Persist the session with OTP verified status
	if err := session.PersistSession(email, sessionID, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{"message": "OTP validated successfully"})
}

func HandleOTPGenerate(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userEmail := email.(string)

	exists, err := database.CheckForOtpEnabled(userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate OTP"})
		return
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{"status": "OTP already activated", "otp_activated": true})
		return
	}

	if pending, ok := getPendingSecret(userEmail); ok {
		c.JSON(http.StatusOK, gin.H{
			"secret":        pending.secret,
			"url":           pending.url,
			"otp_activated": false,
			"pending":       true,
		})
		return
	}

	secret, err := generateOTP(userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate OTP"})
		return
	}

	// Check if secret is not nil before dereferencing
	if secret == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OTP generation resulted in nil"})
		return
	}

	setPendingSecret(userEmail, secret)

	// Dereference secret to access its fields
	c.JSON(http.StatusOK, gin.H{
		"secret":        secret.secret,
		"url":           secret.url,
		"otp_activated": false,
		"pending":       true,
	})
}

func HandleOTPGenerateConfirm(c *gin.Context) {
	emailInterface, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	email, ok := emailInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var payload struct {
		Code string `json:"otp_code"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	if payload.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP code is required"})
		return
	}

	pending, found := getPendingSecret(email)
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No pending OTP setup"})
		return
	}

	if !totp.Validate(payload.Code, pending.secret) {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "Invalid OTP code"})
		return
	}

	if err := database.PersistOTPSecret(email, pending.secret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist OTP"})
		return
	}

	clearPendingSecret(email)

	c.JSON(http.StatusOK, gin.H{"message": "OTP activated successfully"})
}

func generateOTP(email string) (*otp, error) {
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "PRISM",
		AccountName: email,
	})

	if err != nil {
		return nil, err
	}

	otp := &otp{
		secret: secret.Secret(),
		url:    secret.URL(),
	}

	return otp, nil
}
