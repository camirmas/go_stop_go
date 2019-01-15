package models

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"os"
)

type DB struct {*sql.DB}

type Database interface {
	CheckPw(username, password string) (*User, error)
	GetUser(identifier interface{}) (*User, error)
	CreateUser(username, email, password, passwordConfirm string) (*User, error)
	GetGame(gameId int) (*Game, error)
	GetGames(userId int) ([]*Game, error)
	CreateGame(userId, opponentId int) (*Game, error)
	Pass(userId, gameId int) (*Game, error)
}

func getSecret(fileName string) string {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ""
	}

	return string(file)
}

// ConnectDB creates a connection with the postgres database, using credentials
// pulled from environment variables:
//
//	POSTGRES_DB - reads value
//	POSTGRES_USER - reads a file
//	POSTGRES_PASSWORD - reads a file
// 	POSTGRES_HOST - reads value
func ConnectDB() (*DB, error) {
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "go_stop_go"
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	} else {
		user = getSecret(user)
	}

	pw := os.Getenv("POSTGRES_PASSWORD")
	if pw == "" {
		pw = "postgres"
	} else {
		pw = getSecret(pw)
	}

	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, pw, host, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// CheckPw compares the given password against the encrypted password for the given User.
func (db *DB) CheckPw(username, password string) (*User, error) {
	user, err := db.GetUser(username)

	if err != nil {
		return user, err
	}

	inputPassword := []byte(password)
	userPassword := []byte(user.encryptedPassword)
	err = bcrypt.CompareHashAndPassword(userPassword, inputPassword)
	return user, HandleError(err)
}

func parsePlayerRows(rows *sql.Rows) ([]Player, error) {
	defer rows.Close()

	players := make([]Player, 0)
	for rows.Next() {
		var player Player
		rows.Scan(
			&player.Id,
			&player.userId,
			&player.gameId,
			&player.Status,
			&player.Color,
			&player.Stats,
			&player.HasPassed,
			&player.InsertedAt,
			&player.UpdatedAt,
		)
		players = append(players, player)
	}

	return players, rows.Err()
}
