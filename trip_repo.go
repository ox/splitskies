package main

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type TripRepository struct {
	db *sqlx.DB
}

var _ Repo = &TripRepository{}

func (t *TripRepository) CreateTable() error {
	var createTableQuery = `
		create table if not exists trips(
			id text primary key,
			name text not null,
			created_at DATETIME NOT NULL DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
			finished_at DATETIME
		);

		create table if not exists trip_participants(
			user_id text not null,
			trip_id text not null,
			foreign key(user_id) references users(id),
			foreign key(trip_id) references trips(id)
		);
	`
	_, err := t.db.Exec(createTableQuery)
	return err
}

type Trip struct {
	ID           string     `db:"id"`
	Name         string     `db:"name"`
	Participants []User     `db:"-"`
	CreatedAt    time.Time  `db:"created_at"`
	FinishedAt   *time.Time `db:"finished_at"`
}

type TripParticipant struct {
	UserID string `db:"user_id"`
	TripID string `db:"trip_id"`
}

var addTripQuery = `insert into trips (id, name) values (:id, :name);`

func (t *TripRepository) CreateTrip(name string) (*Trip, error) {
	trip := &Trip{
		ID:   generateID(8),
		Name: name,
	}

	_, err := t.db.Exec(addTripQuery, trip)
	return trip, err
}

var addParticipantQuery = `insert into trip_participants (user_id, trip_id) values (:user_id, :trip_id);`

func (t *TripRepository) AddUserToTrip(userID, tripID string) (*TripParticipant, error) {
	participant := &TripParticipant{
		UserID: userID,
		TripID: tripID,
	}
	_, err := t.db.Exec(addParticipantQuery, participant)
	return participant, err
}

var getTripsForUserQuery = `
	with usertrips as (
		select trip_id from trip_participants where user_id = $1
	) select * from trips where trips.id in usertrips;
`

func (t *TripRepository) GetTripsForUser(userID string) ([]*Trip, error) {
	trips := make([]*Trip, 0)
	err := t.db.Select(&trips, getTripsForUserQuery, userID)
	return trips, err
}

var getTripForUserQuery = `
	with usertrips as (
		select trip_id from trip_participants where user_id = $1
	) select * from trips where trips.id = $2 limit 1;
`

func (t *TripRepository) GetTripForUser(tripID, userID string) (*Trip, error) {
	trip := &Trip{}
	err := t.db.Get(trip, getTripForUserQuery, userID, tripID)
	return trip, err
}
