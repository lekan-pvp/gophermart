package models

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/lekan/gophermart/internal/logger"
	_ "github.com/lib/pq"
)

type Token struct {
	UserID int
	jwt.Claims
}

type Credentials struct {
	Login    string `json:"login" db:"username"`
	Password string `json:"password" db:"password"`
}

var db *sqlx.DB
var log = logger.GetLogger()

var schema = `
CREATE TABLE IF NOT EXISTS users(
    id SERIAL,
	username VARCHAR NOT NULL,
	password VARCHAR NOT NULL,
	PRIMARY KEY (id),
    UNIQUE (username)
)`

func InitDB(databaseURI string) error {
	db = sqlx.MustConnect("postgres", databaseURI)

	db.MustExec(schema)

	log.Info().Msg("create db is done...")
	return db.Ping()
}

func Signup(ctx context.Context, creds *Credentials) error {
	_, err := db.ExecContext(ctx, `INSERT INTO users(username, password) VALUES ($1, $2)`, creds.Login, creds.Password)
	if err != nil {
		return fmt.Errorf("409 %w", err)
	}

	return nil
}

func Signin(ctx context.Context, creds *Credentials) error {
	temp := &Credentials{}
	if err := db.GetContext(ctx, temp, `SELECT username, password FROM users WHERE username = $1`, creds.Login); err != nil {
		return err
	}

	return nil
}
