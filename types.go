package main

import (
	"math/rand"
	"time"

	"github.com/lib/pq"
)

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Account struct {
	Id        int         `json:"id"`
	FirstName string      `json:"firstName"`
	LastName  string      `json:"lastName"`
	Number    int64       `json:"number"`
	Balance   int64       `json:"balance"`
	CreatedAt time.Time   `json:"created_at"`
	UpdateAt  time.Time   `json:"updated_at"`
	DeleteAt  pq.NullTime `json:"deleted_at"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Number:    int64(rand.Intn(1000000000)),
		CreatedAt: time.Now().UTC(),
		UpdateAt:  time.Now().UTC(),
	}
}
