package models

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/lekan/gophermart/internal/logger"
	_ "github.com/lib/pq"
	"strconv"
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
	username VARCHAR UNIQUE NOT NULL,
	password VARCHAR NOT NULL,
	balance NUMERIC,
	withdrawn NUMERIC, 
	PRIMARY KEY (username)
);`

var orders = `
CREATE TABLE IF NOT EXISTS orders(
	order_id INT UNIQUE NOT NULL,
	username VARCHAR NOT NULL, 
	status VARCHAR NOT NULL,
	uploaded_at TIMESTAMP NOT NULL,
	PRIMARY KEY (order_id, username),
    FOREIGN KEY (username)
    	REFERENCES users (username)
	);`

func InitDB(databaseURI string) error {
	db = sqlx.MustConnect("postgres", databaseURI)

	db.MustExec(schema)
	db.MustExec(orders)

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

func Luna(num []byte) (bool, error) {
	var err error
	var number int
	l := len(num)
	number = 0
	checkDigit := 0
	for i := 0; i < l; i++ {
		number, err = strconv.Atoi(string(num[i]))
		if err != nil {
			return false, err
		}
		if i%2 != 0 {
			number *= 2
			if number > 9 {
				number -= 9
			}
		}
		checkDigit += number
	}
	lastNumber, err := strconv.Atoi(string(num[l-1]))
	return (checkDigit*9)%10 == lastNumber, nil
}

type Balance struct {
	Current   float64 `json:"current,omitempty" db:"balance"`
	Withdrawn float64 `json:"withdrawn,omitempty" db:"withdrawn"`
}

func GetBalance(ctx context.Context, login string) (*Balance, error) {
	res := &Balance{}
	if err := db.GetContext(ctx, res, `SELECT balance, withdrawn FROM users WHERE username = $1`, login); err != nil {
		return nil, err
	}
	fmt.Println(res.Current, res.Withdrawn)
	return res, nil
}
