package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func (api *API) authCheck(next http.Handler) http.Handler { /* проверка токена */
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		if tokenHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 { //проверка токена на валидность
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenString := splitted[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret-password"), nil
		})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			fmt.Printf("Данные токена JWT: %+v\n", claims)
		}

		next.ServeHTTP(w, r)
	})
}
