package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func WriteJSON(writter http.ResponseWriter, status int, data any) error {
	writter.WriteHeader(status)
	writter.Header().Set("Content-type", "application/json")
	return json.NewEncoder(writter).Encode(data)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string
}

func makeHttpHandleFunc(function apiFunc) http.HandlerFunc {
	return func(writter http.ResponseWriter, resquest *http.Request) {
		if err := function(writter, resquest); err != nil {
			WriteJSON(writter, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAddress string
}

func NewApiServer(listenAddress string) *APIServer {
	return &APIServer{
		listenAddress: listenAddress,
	}
}

func (server *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHttpHandleFunc(server.handleAccount))

	router.HandleFunc("/account/{id}", makeHttpHandleFunc(server.handleGetAccount))

	log.Println("JSON API server running on port: ", server.listenAddress)

	http.ListenAndServe(server.listenAddress, router)
}

func (server *APIServer) handleAccount(writter http.ResponseWriter, request *http.Request) error {
	switch request.Method {
	case "GET":
		return server.handleGetAccount(writter, request)
	case "POST":
		return server.handleCreateAccount(writter, request)
	case "DELETE":
		return server.handleDeleteAccount(writter, request)
	}

	return fmt.Errorf("method not allowed %s", request.Method)
}

func (server *APIServer) handleGetAccount(writter http.ResponseWriter, request *http.Request) error {
	id := mux.Vars(request)["id"]
	fmt.Println(id)
	account := NewAccount("Alex", "Santos")
	return WriteJSON(writter, http.StatusOK, &account)
}

func (server *APIServer) handleCreateAccount(writter http.ResponseWriter, request *http.Request) error {
	return nil
}

func (server *APIServer) handleDeleteAccount(writter http.ResponseWriter, request *http.Request) error {
	return nil
}

func (server *APIServer) handleTransferAccount(writter http.ResponseWriter, request *http.Request) error {
	return nil
}
