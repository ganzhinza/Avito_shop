package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"

	"avito_shop/pkg/db"
	"avito_shop/pkg/structs"
)

const DEFAULT_BALANCE = 1000

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type API struct {
	router *mux.Router
	db     db.Interface
	jwtKey []byte
}

func New(db db.Interface, jwtKey []byte) *API {
	api := API{
		router: mux.NewRouter(),
		db:     db,
		jwtKey: jwtKey,
	}

	api.endpoints()
	return &api
}

func (api *API) createToken(name string) (string, error) {
	claims := Claims{
		Username: name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fullToken, err := token.SignedString(api.jwtKey)
	if err != nil {
		return "", fmt.Errorf("token create: %v", err)
	}

	return fullToken, nil
}

func (api *API) verifyToken(authHeader string) (Claims, error) {
	if authHeader == "" {
		return Claims{}, fmt.Errorf("missing token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return api.jwtKey, nil
	})
	if err != nil || !token.Valid {
		return Claims{}, fmt.Errorf("invalid token %v", err)
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return Claims{}, fmt.Errorf("token has expired")
	}

	return *claims, nil
}

func (api *API) auth(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var authInfo structs.AuthInfo
	err := decoder.Decode(&authInfo)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	exists, storedUser, err := api.db.GetUser(authInfo.Name)
	if _, ok := err.(*structs.DBerror); ok {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		storedUser.Name = authInfo.Name
		storedUser.Password = authInfo.Password
		storedUser.Balance = DEFAULT_BALANCE
		err = api.db.AddUser(storedUser)
		if err != nil {
			http.Error(w, "Could not add user", http.StatusInternalServerError)
			return
		}

	}

	if authInfo.Password != storedUser.Password {
		http.Error(w, "Wrong password", http.StatusUnauthorized)
		return
	}

	tokenString, err := api.createToken(authInfo.Name)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Bearer " + tokenString))
}

func (api *API) info(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	claims, err := api.verifyToken(authHeader)
	if err != nil {
		http.Error(w, "verification error", http.StatusUnauthorized)
		return
	}

	name := claims.Username

	user, err := api.db.GetUserWithHistory(name)
	if err != nil {
		strInfo, code := dbOrBadRequest(err)
		http.Error(w, strInfo, code)
	}

	userEncoded, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
		return
	}
	w.Write(userEncoded)
}

func (api *API) sendCoin(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	claims, err := api.verifyToken(authHeader)
	if err != nil {
		http.Error(w, "verification error", http.StatusUnauthorized)
		return
	}

	name := claims.Username

	var transferInfo structs.CoinsSend

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&transferInfo)
	if err != nil {
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
		return
	}

	err = api.db.SendCoins(name, transferInfo)
	if err != nil {
		strInfo, code := dbOrBadRequest(err)
		http.Error(w, strInfo, code)
	}
}

func (api *API) buyItem(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	claims, err := api.verifyToken(authHeader)
	if err != nil {
		http.Error(w, "verification error", http.StatusUnauthorized)
		return
	}

	name := claims.Username

	vars := mux.Vars(r)
	item := vars["item"]
	err = api.db.BuyItem(name, item)
	if err != nil {
		strInfo, code := dbOrBadRequest(err)
		http.Error(w, strInfo, code)
	}
}

func dbOrBadRequest(err error) (string, int) {
	if _, ok := err.(*structs.DBerror); ok {
		return "Database error", http.StatusInternalServerError
	} else {
		return "Bad request", http.StatusBadRequest
	}
}

func (api *API) Router() *mux.Router {
	return api.router
}

func (api *API) endpoints() {

	api.router.HandleFunc("/api/auth", api.auth).Methods(http.MethodPost)

	api.router.HandleFunc("/api/info", http.HandlerFunc(api.info)).Methods(http.MethodGet)
	api.router.HandleFunc("/api/sendCoin", http.HandlerFunc(api.sendCoin)).Methods(http.MethodPost)
	api.router.HandleFunc("/api/buy/{item}", http.HandlerFunc(api.buyItem)).Methods(http.MethodGet)

}
