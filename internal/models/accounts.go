package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
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

	log.Println("create db is done...")
	return db.Ping()
}

func Validate(ctx context.Context, creds *Credentials) bool {
	var username string
	if err := db.GetContext(ctx, &username, `SELECT username FROM users WHERE username = $1`, creds.Login); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true
		}
		return false
	}

	if username != "" {
		return false
	}
	return true
}

func Signup(ctx context.Context, creds *Credentials) error {
	if !Validate(ctx, creds) {
		return fmt.Errorf("409 %w", errors.New("Login in use"))
	}

	err := db.QueryRowxContext(ctx, `INSERT INTO users(username, password) VALUES ($1, $2)`, creds.Login, creds.Password)
	if err != nil {

	}

	return nil
}

func Signin(ctx context.Context, creds *Credentials) error {
	temp := &Credentials{}
	if err := db.GetContext(ctx, temp, `SELECT username, password FROM users WHERE username = $1 RETURNING`, creds.Login); err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(temp.Password), []byte(creds.Password)); err != nil {
		return err
	}

	return nil
}
