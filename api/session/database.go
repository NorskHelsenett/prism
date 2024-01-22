package session

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID          uint      `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	ExpiresAt   time.Time
	Email       string
	OTPVerified bool
	SessionID   string `gorm:"unique"`
	IsCurrent   bool `gorm:"-"`
}

type SessionStore struct {
	DB *gorm.DB
}

type ValidationResult struct {
	IsValid       bool
	IsOTPVerified bool
}

func (s *SessionStore) DeleteUserSessionsFor(email, sessionID string) error {
	result := s.DB.Where("email = ? AND session_id = ?", email, sessionID).Delete(&Session{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *SessionStore) GetUserSessionsFor(email string) (*[]Session, error) {
	var sessions []Session
	currentTime := time.Now()
	result := s.DB.Where("email = ? AND expires_at > ?", email, currentTime).Find(&sessions)
	if result.Error != nil {
		return nil, result.Error
	}
	return &sessions, nil
}

func NewSessionStore(db *gorm.DB) *SessionStore {
	return &SessionStore{DB: db}
}

func (s *SessionStore) InvalidateSession(email string) error {
	// Assuming you invalidate by deleting the session
	result := s.DB.Where("email = ?", email).Delete(&Session{})
	return result.Error // This will be nil if no error occurred
}

func (s *SessionStore) ValidateSession(email, sessionID string) (ValidationResult, error) {
	var session Session
	currentTime := time.Now()
	result := s.DB.Where("email = ? AND session_id = ? AND expires_at > ?", email, sessionID, currentTime).First(&session)

	// Check if session exists
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return ValidationResult{}, errors.New("invalid session")
	} else if result.Error != nil {
		// Handle any other database error
		return ValidationResult{}, result.Error
	}

	return ValidationResult{
		IsValid:       true,
		IsOTPVerified: session.OTPVerified,
	}, nil
}

func (s *SessionStore) PersistSession(email, sessionID string, verified ...bool) error {
	verifiedValue := false
	if len(verified) > 0 {
		verifiedValue = verified[0]
	}
	var session Session

	// Check if a session already exists for the given email
	result := s.DB.First(&session, "session_id = ?", sessionID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// If not found, create a new session
		newSession := Session{
			Email:       email,
			OTPVerified: verifiedValue,                  // or true, depending on your logic
			ExpiresAt:   time.Now().Add(24 * time.Hour), // Example: 24 hours from now
			SessionID:   sessionID,
		}
		return s.DB.Create(&newSession).Error
	} else if result.Error != nil {
		// Handle any other database error
		return result.Error
	}

	// If a session already exists, you can update it if necessary
	session.OTPVerified = verifiedValue               // or true, depending on your logic
	session.ExpiresAt = time.Now().Add(8 * time.Hour) // Example: 8 hours from now
	return s.DB.Model(&session).Where("email = ?", email).Updates(Session{OTPVerified: session.OTPVerified, ExpiresAt: session.ExpiresAt}).Error
}
