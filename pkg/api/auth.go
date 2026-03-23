package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

const jwtSecret = "todo-app-secret"

func signinHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		sendError(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}

	var data map[string]string
	if err = json.Unmarshal(buf.Bytes(), &data); err != nil {
		log.Println("ошибка десериализации:", err)
		sendError(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	password := data["password"]
	envPassword := os.Getenv("TODO_PASSWORD")

	if password != envPassword {
		sendError(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	// создаём JWT токен, кладём в claims хеш пароля
	hash := sha256.Sum256([]byte(password))
	claims := jwt.MapClaims{
		"password_hash": fmt.Sprintf("%x", hash),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString([]byte(jwtSecret))
	if err != nil {
		sendError(w, "Ошибка создания токена", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]string{"token": signedToken})
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(resp)
}

// auth оборачивает обработчик проверкой аутентификации
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var tokenStr string
			cookie, err := r.Cookie("token")
			if err == nil {
				tokenStr = cookie.Value
			}

			if tokenStr == "" {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			hash := sha256.Sum256([]byte(pass))
			expectedHash := fmt.Sprintf("%x", hash)
			if claims["password_hash"] != expectedHash {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}
