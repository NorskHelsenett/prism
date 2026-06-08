package session

import (
	"errors"
	"prism/config"
	"prism/database"
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
	IsCurrent   bool   `gorm:"-"`
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

func (s *SessionStore) IsAdmin(email string) bool {
	var user database.UserData
	result := s.DB.Where("email = ?", email).First(&user)

	// Directly return false if any error (including not found) occurs
	if result.Error != nil {
		return false
	}

	return user.Role == "admin"
}

func (s *SessionStore) IsActive(email string) bool {
	var user database.UserData
	result := s.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return false
	}
	return user.Active == nil || *user.Active
}

func (s *SessionStore) GetRole(email string) string {
	var user database.UserData
	result := s.DB.Where("email = ?", email).First(&user)

	if result.Error != nil || user.Role == "" {
		return "visitor"
	}

	return user.Role
}

func (s *SessionStore) SaveOrUpdateUserData(name string, email string, picture string) error {
	var existingUserData database.UserData

	// First, try to find the existing user data by email
	result := s.DB.Where("email = ?", email).First(&existingUserData)

	now := time.Now()

	// Handle the case where the user data might not exist
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// If not found, create a new record
		newUserData := &database.UserData{
			Name:     name,
			Email:    email,
			Picture:  picture,
			LastSeen: &now,
		}
		return s.DB.Create(newUserData).Error
	} else if result.Error != nil {
		// Handle other potential errors
		return result.Error
	}

	// If found, update the existing record
	existingUserData.Name = name
	if picture != "" {
		existingUserData.Picture = picture
	}
	existingUserData.LastSeen = &now
	return s.DB.Save(&existingUserData).Error
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
			OTPVerified: verifiedValue,
			ExpiresAt:   time.Now().Add(8 * time.Hour),
			SessionID:   sessionID,
		}
		return s.DB.Create(&newSession).Error
	} else if result.Error != nil {
		// Handle any other database error
		return result.Error
	}

	// If a session already exists, you can update it if necessary
	session.OTPVerified = verifiedValue
	session.ExpiresAt = time.Now().Add(8 * time.Hour)
	return s.DB.Model(&session).Where("email = ?", email).Updates(Session{OTPVerified: session.OTPVerified, ExpiresAt: session.ExpiresAt}).Error
}

func (s *SessionStore) GetUserDataByEmail(email string) (*database.UserData, error) {
	var userData database.UserData
	result := s.DB.Where("email = ?", email).First(&userData)

	if result.Error != nil {
		return nil, result.Error
	}

	return &userData, nil
}

func GetAllProfiles(s *SessionStore) (*[]database.UserData, error) {
	var userData []database.UserData
	result := s.DB.Select("Name", "Picture", "Email").Order("role desc").Find(&userData)

	if result.Error != nil {
		return nil, result.Error
	}

	return &userData, nil
}

func GetAllUsers(s *SessionStore) (*[]database.UserData, error) {
	var userData []database.UserData
	result := s.DB.Find(&userData)

	if result.Error != nil {
		return nil, result.Error
	}

	return &userData, nil
}

func (s *SessionStore) UpdateUserRole(user *database.UserData) error {
	return s.DB.Model(&database.UserData{}).Where("email = ?", user.Email).Update("role", user.Role).Error
}

func LoadSessionStore(s *SessionStore) {
	// Load users from the main database
	users, err := database.GetAllUsers()
	if err != nil {
		panic("failed to load users from the main database")
	}

	// Prepare a set of admin emails for quick lookup
	adminEmails := make(map[string]struct{})
	for _, adminEmail := range config.AppConfig.Admins {
		adminEmails[adminEmail] = struct{}{}
	}

	// 2. Check if the database is empty
	if len(*users) == 0 {
		// Database is empty, create new admin users
		for _, adminEmail := range config.AppConfig.Admins {
			newUser := database.UserData{Email: adminEmail, Role: "admin"}
			s.DB.Create(&newUser)
		}
		return // Exit after creating admin users as there are no existing users to update
	}

	// Insert users into the session database
	for _, user := range *users {
		if _, isAdmin := adminEmails[user.Email]; isAdmin || user.Role == "admin" {
			user.Role = "admin"
		}
		s.DB.Create(&user)
	}
}

func isAdmin(email string) bool {

	admins := config.AppConfig.Admins

	isAdmin := false
	for _, adminEmail := range admins {
		if email == adminEmail {
			isAdmin = true
			break
		}
	}
	return isAdmin
}
