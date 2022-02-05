package models

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/lekan/gophermart/internal/utils"
	_ "github.com/lib/pq"
	"github.com/omeid/pgerror"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

type Token struct {
	UserID int
	jwt.Claims
}

type Account struct {
	Login    string `json:"login" db:"user_login"`
	Password string `json:"password" db:"user_password"`
	Token    string `json:"token" db:"-"`
}

var db *sqlx.DB

var schema = `
CREATE TABLE IF NOT EXISTS users(
    id SERIAL,
	user_login VARCHAR NOT NULL,
	user_password VARCHAR NOT NULL,
	PRIMARY KEY (id),
    UNIQUE (user_login)
)`

func InitDB(databaseURI string) error {
	db = sqlx.MustConnect("postgres", databaseURI)

	db.MustExec(schema)

	log.Println("create db is done...")
	return db.Ping()
}

func (account *Account) Validate(ctx context.Context) (map[string]interface{}, bool) {
	temp := &Account{}

	//проверка на наличие ошибок и дубликатов
	err := db.QueryRowxContext(ctx, `SELECT user_login, user_password FROM users WHERE user_login = $1`, account.Login).Scan(&temp.Login, &temp.Password)
	if err != nil {
		if e := pgerror.CaseNotFound(err); e != nil {
			return utils.Message(false, 200, e.Error()), true
		}
		return utils.Message(false, 500, err.Error()), true
	}

	if temp.Login != "" {
		return utils.Message(false, 409, "duplicate"), false
	}

	log.Println("In Validate: DB is initiate")
	return utils.Message(true, 409, "Requirement passed"), false
}

func (account *Account) CreateUser(ctx context.Context) map[string]interface{} {
	if resp, ok := account.Validate(ctx); !ok {
		return resp
	}

	log.Println("In CreateUser: DB is initiate")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.Message(false, 500, err.Error())
	}
	account.Password = string(hashedPassword)

	var id int
	err = db.QueryRowxContext(ctx, `INSERT INTO users (user_login, user_password) VALUES ($1, $2) RETURNING id`, account.Login, account.Password).Scan(&id)
	if err != nil {
		return utils.Message(false, 500, "wrong INSERT...")
	}

	tk := &Token{UserID: id}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("token_password")))
	if err != nil {
		return utils.Message(false, 500, err.Error())
	}
	account.Token = tokenString
	account.Password = ""
	response := utils.Message(true, 200, "Account has been created")
	response["account"] = account
	return response
}

func GetUser(ctx context.Context, login uint) (*Account, error) {
	acc := &Account{}

	err := db.QueryRowxContext(ctx, `SELECT user_login, user_password FROM users WHERE user_login = $1`, login).Scan(acc.Login, acc.Password)
	if err != nil {
		return nil, err
	}
	if acc.Login == "" {
		return nil, errors.New("Find nothing")
	}
	return acc, nil
}
