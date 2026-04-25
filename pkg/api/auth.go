package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// hashString computes the SHA‑256 hash of a string and returns it as a hex‑encoded string.
func hashString(s string) string {
	result := sha256.Sum256([]byte(s))
	hashString := hex.EncodeToString(result[:])
	return hashString
}

// generateJWT creates a JWT token containing a claim with the SHA‑256 hash of the password.
// The token is signed using the password itself as the secret (HS256).
// Returns the signed token string or an error if signing fails.
func generateJWT(password string) (string, error) {
	claims := jwt.MapClaims{
		"password_hash": hashString(password),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString([]byte(password))
	if err != nil {
		fmt.Printf("failed to sign jwt: %s\n", err)
		return "", err
	}

	return signedToken, nil
}

// SigninHandler handles POST /api/signin requests.
// It expects a JSON body with a "password" field.
// Authentication logic:
//   - If TODO_PASSWORD environment variable is not set (development mode), any password is accepted
//     and a JWT token is generated using the submitted password as the secret.
//   - If TODO_PASSWORD is set, the submitted password must match it exactly.
//     On success, a JWT token is generated using the environment password as the secret.
//
// The handler returns a JSON object with a "token" field on success, or an appropriate HTTP error.
func SigninHandler(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Password string `json:"password"`
	}
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	envPassword := os.Getenv("TODO_PASSWORD")

	if envPassword == "" {
		// If no password is set, allow any password (for development)
		token, err := generateJWT(req.Password)
		if err != nil {
			http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
			return
		}
		resp, err := json.Marshal(map[string]string{"token": token})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}
	// Compare passwords
	if req.Password != envPassword {
		http.Error(w, `{"error": "Неверный пароль"}`, http.StatusUnauthorized)
		return
	}
	// Generate JWT token
	token, err := generateJWT(envPassword)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
