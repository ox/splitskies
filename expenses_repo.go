package main

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type ExpensesRepository struct {
	db *sqlx.DB
}

var _ Repo = &ExpensesRepository{}

type Expense struct {
	ID             string                `db:"id"`
	Name           string                `db:"name"`
	OwnerUsername  string                `db:"owner_username"`
	OwnerID        string                `db:"owner_id"`
	OwnerCostCents int                   `db:"owner_cost_cents"`
	TripID         string                `db:"trip_id"`
	CreatedAt      time.Time             `db:"created_at"`
	Participants   []*ExpenseParticipant `db:"-"`
}

type ExpenseParticipant struct {
	UserID    string `db:"user_id"`
	ExpenseID string `db:"expense_id"`
	CostCents int    `db:"cost_cents"`

	UserName string `db:"username"`
}

func (er *ExpensesRepository) CreateTable() error {
	var createTableQuery = `
	create table if not exists expenses(
		id text primary key,
		name text not null,
		owner_id text not null,
		owner_cost_cents integer not null,
		trip_id text not null,
		created_at DATETIME NOT NULL DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
		foreign key(owner_id) references users(id),
		foreign key(trip_id) references trips(id)
	);

	create table if not exists expense_participants(
		user_id text not null,
		expense_id text not null,
		cost_cents integer not null,
		foreign key(user_id) references users(id),
		foreign key(expense_id) references expenses(id)
	);
`
	_, err := er.db.Exec(createTableQuery)
	return err
}

var getExpensesForTripQuery = `
	select expenses.*, users.username as owner_username from expenses, users where trip_id = $1 and users.id = expenses.owner_id;
`

var getExpenseParticipantsQuery = `
	select expense_participants.*, users.username
	from expense_participants left join users on expense_participants.user_id = users.id
	where expense_id = $1;
`

func (er *ExpensesRepository) GetTripExpenses(tripID string) ([]*Expense, error) {
	expenses := make([]*Expense, 0)
	err := er.db.Select(&expenses, getExpensesForTripQuery, tripID)
	if err != nil {
		return nil, fmt.Errorf("could not get expenses for trip %s: %w", tripID, err)
	}

	for _, expense := range expenses {
		participants := make([]*ExpenseParticipant, 0)
		err := er.db.Select(&participants, getExpenseParticipantsQuery, expense.ID)
		if err != nil {
			return nil, fmt.Errorf("could not get participants for expense %s of trip %s: %w", expense.ID, tripID, err)
		}

		expense.Participants = participants
	}

	return expenses, nil
}
