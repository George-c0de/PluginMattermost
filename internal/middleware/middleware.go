// Package middleware Получение User_id
package middleware

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
)

type contextKey string

const userIDKey = contextKey("user_id")

// UserIDMiddleware проверяет наличие cookie "user_id" и, если её нет
func UserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")
		var userID string
		if err != nil || cookie.Value == "" {
			userID = strconv.FormatUint(rand.Uint64(), 10)
			http.SetCookie(w, &http.Cookie{
				Name:  "user_id",
				Value: userID,
				Path:  "/",
			})
		} else {
			userID = cookie.Value
		}

		// Добавляем user_id в контекст запроса
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
