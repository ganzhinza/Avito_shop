package api

import (
	"encoding/json"
	"fmt"
	"log"
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
		log.Fatal(err)
	}
	api.verifyToken(fullToken)
	return token.SignedString(api.jwtKey)
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
		log.Fatal(err)
	}
	exists, storedUser, err := api.db.GetUser(authInfo.Name)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		storedUser.Name = authInfo.Name
		storedUser.Password = authInfo.Password
		storedUser.Balance = DEFAULT_BALANCE
		err = api.db.AddUser(storedUser)
		if err != nil {
			log.Fatal(err)
		}

	}

	if authInfo.Password != storedUser.Password {
		log.Fatal(err)
	}

	tokenString, err := api.createToken(authInfo.Name)
	if err != nil {
		log.Fatal(err)
	}
	w.Write([]byte("Bearer " + tokenString))
}

func (api *API) info(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	claims, err := api.verifyToken(authHeader)
	if err != nil {
		log.Fatal(err)
	}

	name := claims.Username

	user, err := api.db.GetUserWithHistory(name)
	if err != nil {
		log.Fatal(err)
	}

	userEncoded, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(userEncoded)
	if err != nil {
		log.Fatal(err)
	}
}

func (api *API) sendCoin(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	claims, err := api.verifyToken(authHeader)
	if err != nil {
		log.Fatal(err)
	}

	name := claims.Username

	var transferInfo structs.CoinsSend

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&transferInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = api.db.SendCoins(name, transferInfo)
	if err != nil {
		log.Fatal(err)
	}
}

func (api *API) buyItem(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	claims, err := api.verifyToken(authHeader)
	if err != nil {
		log.Fatal(err)
	}

	name := claims.Username

	vars := mux.Vars(r)
	item := vars["item"]
	err = api.db.BuyItem(name, item)
	if err != nil {
		log.Fatal(err)
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
