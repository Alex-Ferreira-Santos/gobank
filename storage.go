package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=admin dbname=gobank password=123 sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (store *PostgresStore) Init() error {
	return store.createAccountTable()
}

func (store *PostgresStore) createAccountTable() error {
	query := `create table if not exists account(
		Id serial primary key, 
		created_at TIMESTAMP WITH TIME ZONE DEFAULT Now() not null,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT Now() not null,
		deleted_at TIMESTAMP WITH TIME ZONE,
		first_name varchar(50) not null, 
		last_name varchar(50) not null, 
		number serial,
		balance serial
	);`

	_, err := store.db.Exec(query)
	return err
}

func (store *PostgresStore) CreateAccount(account *Account) error {
	query := `insert into account(first_name, last_name, number, balance) values ($1, $2, $3, $4)`

	response, err := store.db.Exec(query, account.FirstName, account.LastName, account.Number, account.Balance)

	if err != nil {
		return err
	}

	fmt.Printf("%v\n", response)

	return nil
}

func (store *PostgresStore) DeleteAccount(id int) error {
	return nil
}

func (store *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (store *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := store.db.Query("Select id, first_name, last_name, number, balance, created_at, updated_at, deleted_at from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account := new(Account)
		err := rows.Scan(
			&account.Id,
			&account.FirstName,
			&account.LastName,
			&account.Number,
			&account.Balance,
			&account.CreatedAt,
			&account.UpdateAt,
			&account.DeleteAt,
		)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (store *PostgresStore) GetAccountById(id int) (*Account, error) {
	return &Account{}, nil
}
