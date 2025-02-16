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

func New(db db.Interface) *API {
	api := API{
		router: mux.NewRouter(),
		db:     db,
		jwtKey: []byte("my-secret-key"), //TODO: get from env
	}

	api.endpoints()
	return &api
}

func (api *API) endpoints() {

	api.router.HandleFunc("/api/auth", api.auth).Methods(http.MethodPost)

	api.router.HandleFunc("/api/info", http.HandlerFunc(api.info)).Methods(http.MethodGet)
	api.router.HandleFunc("/api/sendCoin", http.HandlerFunc(api.sendCoin)).Methods(http.MethodPost)
	api.router.HandleFunc("/api/buy/{item}", http.HandlerFunc(api.buyItem)).Methods(http.MethodGet)

}

func (api *API) verifyToken(authHeader string) (Claims, error) {
	if authHeader == "" {
		return Claims{}, fmt.Errorf("Missing token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return api.jwtKey, nil
	})
	if err != nil || !token.Valid {
		return Claims{}, fmt.Errorf("Invalid token")
	}

	claims, ok := token.Claims.(Claims)
	if !ok || claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return Claims{}, fmt.Errorf("Token has expired")
	}

	return claims, nil
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

func (api *API) auth(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	var authInfo structs.AuthInfo
	authInfo.Name = params.Get("name")
	authInfo.Password = params.Get("password")

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
	w.Write([]byte(tokenString))
}

func (api *API) createToken(name string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": name,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(api.jwtKey)
}
