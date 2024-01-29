package main

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

var _ Repo = &UserRepository{}

type User struct {
	ID        string     `db:"id"`
	Username  string     `db:"username"`
	Email     string     `db:"email"`
	Phone     string     `db:"phone"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (u *UserRepository) CreateTable() error {
	var createTableQuery = `
		create table if not exists users(
			id text primary key,
			username text not null,
			email text not null default '',
			phone text not null,
			created_at DATETIME NOT NULL DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
			deleted_at DATETIME
		);
	`
	_, err := u.db.Exec(createTableQuery)
	return err
}

var getUserByIDQuery = `select * from users where id = $1;`

func (u *UserRepository) GetUserByID(id string) (*User, bool, error) {
	user := &User{}
	err := u.db.Get(user, getUserByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return user, true, nil
}

var getUserByPhoneQuery = `select * from users where phone = $1;`

func (u *UserRepository) GetUserByPhone(phone string) (*User, bool, error) {
	user := &User{}
	err := u.db.Get(user, getUserByPhoneQuery, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return user, true, nil
}

var addUserQuery = `insert into users (id, username, phone) values (:id, :username, :phone);`

func (u *UserRepository) AddUser(username, phone string) (*User, error) {
	user := &User{
		ID:       generateID(8),
		Username: username,
		Phone:    phone,
	}
	_, err := u.db.NamedExec(addUserQuery, user)
	return user, err
}
