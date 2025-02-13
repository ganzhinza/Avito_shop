package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"

	"avito_shop/pkg/structs"
)

type API struct {
	router *mux.Router
	tokens map[string]jwt.Token
}

func New() API {
	api := API{
		router: mux.NewRouter(),
		tokens: make(map[string]jwt.Token),
	}

	api.endpoints()
	return api
}

func (api *API) endpoints() {
	api.router.Use(api.authCheck)

	api.router.HandleFunc("/api/info", api.info).Methods(http.MethodGet)
	api.router.HandleFunc("/api/sendCoin", api.sendCoin).Methods(http.MethodPost)
	api.router.HandleFunc("/api/buy/{item}", api.buyItem).Methods(http.MethodGet)
	api.router.HandleFunc("/api/auth", api.auth).Methods(http.MethodPost)
}

func (api *API) info(w http.ResponseWriter, r *http.Request) {

}

func (api *API) sendCoin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err) //TODO
	}

	CoinsMessage := structs.CoinsMessage{}
	err = json.Unmarshal(body, &CoinsMessage)
	if err != nil {
		log.Fatal(err) //TODO
	}

	db.

}

func (api *API) buyItem(w http.ResponseWriter, r *http.Request) {

}

func (api *API) auth(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	name := params.Get("name")
	password := params.Get("password")

	User := user{name: name, password: password}

	if checkUser(User) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"usr": User.name,
		})

		tokenString, err := token.SignedString(password) //Make smth better
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(tokenString))
	}
}

// TODO
func checkUser(User structs.User) bool {
	return true
}
