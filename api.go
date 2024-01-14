package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	router.HandleFunc("/account/{id}", makeHttpHandleFunc(server.handleAccountById))

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
