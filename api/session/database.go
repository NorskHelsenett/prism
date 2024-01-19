package session

import (
	"gorm.io/gorm"
	"errors"
	"time"
)

type Session struct {
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	ExpiresAt     time.Time
	Email         string `gorm:"unique" gorm:"primaryKey"`
	OTPVerified   bool
}

type SessionStore struct {
    DB *gorm.DB
}

type ValidationResult struct {
    IsValid      bool
    IsOTPVerified bool
}

func NewSessionStore(db *gorm.DB) *SessionStore {
    return &SessionStore{DB: db}
}

func (s *SessionStore) InvalidateSession(email string) error {
	// Assuming you invalidate by deleting the session
	result := s.DB.Where("email = ?", email).Delete(&Session{})
	return result.Error // This will be nil if no error occurred
}

func (s *SessionStore) ValidateSession(email string) (ValidationResult, error) {
    var session Session
    result := s.DB.Where("email = ?", email).First(&session)

    // Check if session exists
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return ValidationResult{}, errors.New("session not found")
    } else if result.Error != nil {
        // Handle any other database error
        return ValidationResult{}, result.Error
    }

    // Check if the session has expired
    if session.ExpiresAt.Before(time.Now()) {
        return ValidationResult{}, errors.New("session has expired")
    }

    return ValidationResult{
        IsValid:      true,
        IsOTPVerified: session.OTPVerified,
    }, nil
}

func (s *SessionStore) PersistSession(email string, verified ...bool) error {
		verifiedValue := false
    if len(verified) > 0 {
        verifiedValue = verified[0]
    }
    var session Session

    // Check if a session already exists for the given email
    result := s.DB.First(&session, "email = ?", email)

    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        // If not found, create a new session
        newSession := Session{
            Email:       email,
            OTPVerified: verifiedValue, // or true, depending on your logic
            ExpiresAt:   time.Now().Add(24 * time.Hour), // Example: 24 hours from now
        }
        return s.DB.Create(&newSession).Error
    } else if result.Error != nil {
        // Handle any other database error
        return result.Error
    }

    // If a session already exists, you can update it if necessary
    session.OTPVerified = verifiedValue // or true, depending on your logic
    session.ExpiresAt = time.Now().Add(8 * time.Hour) // Example: 8 hours from now
    return s.DB.Model(&session).Where("email = ?", email).Updates(Session{OTPVerified: session.OTPVerified, ExpiresAt: session.ExpiresAt}).Error
}
