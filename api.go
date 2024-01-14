package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Valgard/godotenv"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddress string
	store         Storage
}

func NewApiServer(listenAddress string, store Storage) *APIServer {
	return &APIServer{
		listenAddress: listenAddress,
		store:         store,
	}
}

func (server *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHttpHandleFunc(server.handleAccount))

	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandleFunc(server.handleAccountById), server.store))

	router.HandleFunc("/account/{id}/transfer", makeHttpHandleFunc(server.handleTransfer))

	log.Println("JSON API server running on port: ", server.listenAddress)

	http.ListenAndServe(server.listenAddress, router)
}

func (server *APIServer) handleAccount(writter http.ResponseWriter, request *http.Request) error {
	switch request.Method {
	case "GET":
		return server.handleGetAccount(writter, request)
	case "POST":
		return server.handleCreateAccount(writter, request)
	}
	return fmt.Errorf("method not allowed %s", request.Method)
}

func (server *APIServer) handleAccountById(writter http.ResponseWriter, request *http.Request) error {
	switch request.Method {
	case "GET":
		return server.handleGetAccountById(writter, request)
	case "DELETE":
		return server.handleDeleteAccount(writter, request)
	}

	return fmt.Errorf("method not allowed %s", request.Method)
}

func (server *APIServer) handleTransfer(writter http.ResponseWriter, request *http.Request) error {
	switch request.Method {
	case "POST":
		return server.handleTransferAccount(writter, request)
	}

	return fmt.Errorf("method not allowed %s", request.Method)
}

func (server *APIServer) handleGetAccount(writter http.ResponseWriter, request *http.Request) error {
	account, err := server.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(writter, http.StatusOK, &account)
}

func (server *APIServer) handleGetAccountById(writter http.ResponseWriter, request *http.Request) error {
	id, err := getId(request)
	if err != nil {
		return err
	}

	account, err := server.store.GetAccountById(id)

	if err != nil {
		return err
	}
	return WriteJSON(writter, http.StatusOK, &account)
}

func (server *APIServer) handleCreateAccount(writter http.ResponseWriter, request *http.Request) error {

	createAccountRequest := new(CreateAccountRequest)
	if err := json.NewDecoder(request.Body).Decode(&createAccountRequest); err != nil {
		return err
	}

	account := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName)

	if err := server.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	fmt.Print(tokenString)

	return WriteJSON(writter, http.StatusOK, account)
}

func (server *APIServer) handleDeleteAccount(writter http.ResponseWriter, request *http.Request) error {
	id, err := getId(request)
	if err != nil {
		return err
	}

	if err := server.store.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJSON(writter, http.StatusOK, map[string]int{"deleted": id})
}

func (server *APIServer) handleTransferAccount(writter http.ResponseWriter, request *http.Request) error {
	transferRequest := new(TransferRequest)

	if err := json.NewDecoder(request.Body).Decode(transferRequest); err != nil {
		return nil
	}
	defer request.Body.Close()
	return WriteJSON(writter, http.StatusOK, transferRequest)
}

func WriteJSON(writter http.ResponseWriter, status int, data any) error {
	writter.Header().Set("Content-type", "application/json")
	writter.WriteHeader(status)
	return json.NewEncoder(writter).Encode(data)
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"exp":           time.Now().Add(time.Hour * 1).Unix(),
		"accountNumber": account.Number,
	}

	err := loadEnv()

	if err != nil {
		return "", fmt.Errorf("env file not found")
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func permissionDenied(writter http.ResponseWriter) {
	WriteJSON(writter, http.StatusForbidden, ApiError{Error: "permission denied"})
}

func withJWTAuth(handlerFunction http.HandlerFunc, store Storage) http.HandlerFunc {
	return func(writter http.ResponseWriter, request *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := request.Header.Get("x-jwt-token")

		token, err := validateJWT(tokenString)

		if err != nil {
			permissionDenied(writter)
			return
		}

		if !token.Valid {
			permissionDenied(writter)
			return
		}

		userId, err := getId(request)
		if err != nil {
			permissionDenied(writter)
			return
		}
		account, err := store.GetAccountById(userId)
		if err != nil {
			permissionDenied(writter)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64(claims["accountNumber"].(float64)) {
			permissionDenied(writter)
			return
		}

		fmt.Print(claims)

		handlerFunction(writter, request)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	err := loadEnv()

	if err != nil {
		return nil, fmt.Errorf("env file not found")
	}

	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHttpHandleFunc(function apiFunc) http.HandlerFunc {
	return func(writter http.ResponseWriter, resquest *http.Request) {
		if err := function(writter, resquest); err != nil {
			WriteJSON(writter, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getId(request *http.Request) (int, error) {
	idString := mux.Vars(request)["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		return id, fmt.Errorf("Invalid id given %s", idString)
	}
	return id, nil
}

func loadEnv() error {
	dotenv := godotenv.New()
	if err := dotenv.Load(".env"); err != nil {
		return fmt.Errorf("env file not found")
	}
	return nil
}
