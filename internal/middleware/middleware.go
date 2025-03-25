package middleware

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
)

// UserIDMiddleware проверяет наличие cookie "user_id" и, если её нет
func UserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем наличие cookie "user_id"
		cookie, err := r.Cookie("user_id")
		var userID string
		if err != nil || cookie.Value == "" {
			// Если cookie отсутствует, генерируем новый user_id
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
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
