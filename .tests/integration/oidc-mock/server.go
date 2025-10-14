package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"gopkg.in/yaml.v2"
)

type User struct {
	Email       string `yaml:"email"`
	Role        string `yaml:"role"`
	Password    string `yaml:"password"`
	SkipOTP     bool   `yaml:"skip_otp"`
	DisplayName string `yaml:"display_name"`
}

type UsersConfig struct {
	Users []User `yaml:"users"`
}

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	users      map[string]User
	issuerURL  string
	// simple in-memory map from auth code to user email for tests
	codeToUser map[string]string
)

func init() {
	// Generate RSA key pair
	var err error
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}
	publicKey = &privateKey.PublicKey

	// Load users from config
	users = make(map[string]User)
	codeToUser = make(map[string]string)
	configPath := os.Getenv("USERS_CONFIG_PATH")
	if configPath == "" {
		configPath = "/config/users.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: Could not read users config: %v", err)
		// Create default user
		users["test@example.com"] = User{
			Email:       "test@example.com",
			Role:        "admin",
			Password:    "password",
			SkipOTP:     true,
			DisplayName: "Test User",
		}
	} else {
		var config UsersConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Fatalf("Failed to parse users config: %v", err)
		}
		for _, u := range config.Users {
			users[u.Email] = u
		}
	}

	issuerURL = os.Getenv("ISSUER_URL")
	if issuerURL == "" {
		issuerURL = "http://localhost:9999"
	}

	log.Printf("Mock OIDC server initialized with %d users", len(users))
}

// OIDC Discovery endpoint
func discoveryHandler(w http.ResponseWriter, r *http.Request) {
	discovery := map[string]interface{}{
		"issuer":                 issuerURL,
		"authorization_endpoint": issuerURL + "/authorize",
		"token_endpoint":         issuerURL + "/token",
		"userinfo_endpoint":      issuerURL + "/userinfo",
		"jwks_uri":               issuerURL + "/.well-known/jwks.json",
		"response_types_supported": []string{
			"code",
			"token",
			"id_token",
		},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported": []string{
			"openid",
			"profile",
			"email",
		},
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_post",
			"client_secret_basic",
		},
		"claims_supported": []string{
			"sub",
			"email",
			"name",
			"preferred_username",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discovery)
}

// JWKS endpoint
func jwksHandler(w http.ResponseWriter, r *http.Request) {
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	// proper big-endian exponent bytes for 65537
	eBytes := []byte{0x01, 0x00, 0x01}
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": "mock-key-1",
				"alg": "RS256",
				"n":   n,
				"e":   e,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jwks)
}

// Authorization endpoint
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	// allow selecting a user via `login` query param for tests
	login := r.URL.Query().Get("login")

	// Generate a simple authorization code
	code := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("code-%d", time.Now().UnixNano())))

	// remember which user this code belongs to (if provided)
	if login != "" {
		codeToUser[code] = login
	}

	// Redirect back with code
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// Token endpoint
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	code := r.FormValue("code")

	if grantType != "authorization_code" || code == "" {
		http.Error(w, "Invalid grant", http.StatusBadRequest)
		return
	}

	// Resolve which user this code maps to. If mapping exists, use it,
	// otherwise fall back to the first user in the users map.
	var user User
	if email, ok := codeToUser[code]; ok {
		if u, exists := users[email]; exists {
			user = u
		}
	}
	if user.Email == "" {
		for _, u := range users {
			user = u
			break
		}
	}

	// Generate tokens
	idToken, err := generateIDToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	accessToken, err := generateAccessToken(user)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"id_token":      idToken,
		"refresh_token": "mock-refresh-token",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Userinfo endpoint
func userinfoHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	user, exists := users[email]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	userinfo := map[string]interface{}{
		"sub":                email,
		"email":              user.Email,
		"name":               user.DisplayName,
		"preferred_username": user.Email,
		"email_verified":     true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userinfo)
}

// Direct login endpoint for testing (bypasses OAuth flow)
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, exists := users[req.Email]
	if !exists || user.Password != req.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	idToken, err := generateIDToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	accessToken, err := generateAccessToken(user)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token": accessToken,
		"id_token":     idToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateIDToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"iss":                issuerURL,
		"sub":                user.Email,
		"aud":                "prism-test-client",
		"exp":                time.Now().Add(time.Hour).Unix(),
		"iat":                time.Now().Unix(),
		"email":              user.Email,
		"email_verified":     true,
		"name":               user.DisplayName,
		"preferred_username": user.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "mock-key-1"

	return token.SignedString(privateKey)
}

func generateAccessToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"iss":   issuerURL,
		"sub":   user.Email,
		"aud":   "prism-test-client",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"email": user.Email,
		"scope": "openid profile email",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"users":  fmt.Sprintf("%d", len(users)),
	})
}

func main() {
	http.HandleFunc("/.well-known/openid-configuration", discoveryHandler)
	http.HandleFunc("/.well-known/jwks.json", jwksHandler)
	http.HandleFunc("/authorize", authorizeHandler)
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/userinfo", userinfoHandler)
	http.HandleFunc("/login", loginHandler) // Direct login for tests
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}

	log.Printf("Mock OIDC server listening on :%s", port)
	log.Printf("Issuer URL: %s", issuerURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
