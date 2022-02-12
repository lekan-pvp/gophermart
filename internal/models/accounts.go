package models

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/lekan/gophermart/internal/cfg"
	"github.com/lekan/gophermart/internal/logger"
	_ "github.com/lib/pq"
	"net/http"
	"sort"
	"strconv"
	"time"
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
	order_id VARCHAR UNIQUE NOT NULL,
	username VARCHAR NOT NULL, 
	status VARCHAR NOT NULL,
	accrual NUMERIC,
	uploaded_at TIMESTAMP NOT NULL,
	PRIMARY KEY (order_id, username),
    FOREIGN KEY (username)
    	REFERENCES users (username)
	);`

var withdrawals = `
CREATE TABLE IF NOT EXISTS withdrawals(
    operation_id SERIAL,
	username VARCHAR NOT NULL,
	order_id VARCHAR NOT NULL,
	withdraw_sum NUMERIC,
	processed_at TIMESTAMP NOT NULL,
	PRIMARY KEY (operation_id),
    FOREIGN KEY (username)
    	REFERENCES users (username),
    FOREIGN KEY (order_id)
    	REFERENCES orders (order_id));
`

func InitDB(databaseURI string) error {
	db = sqlx.MustConnect("postgres", databaseURI)

	db.MustExec(schema)
	db.MustExec(orders)
	db.MustExec(withdrawals)

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

type Order struct {
	OrderId string  `json:"order_id" db:"order_id"`
	Status  string  `json:"status" db:"status"`
	Accrual float32 `json:"accrual" db:"accrual"`
}

func PostOrder(ctx context.Context, login string, orderId []byte) (int, error) {
	ok, err := Luna(orderId)
	if err != nil {
		log.Err(err).Msg("convert number error")
		return http.StatusInternalServerError, err
	}
	if !ok {
		log.Info().Msg("wrong order number format")
		return http.StatusUnprocessableEntity, nil
	}

	order := Order{}
	address := cfg.GetAccuralSystemAddress()
	response, err := http.Get(address + "/api/orders/" + string(orderId))
	if err != nil {
		return response.StatusCode, err
	}

	if err = json.NewDecoder(response.Body).Decode(&order); err != nil {
		return 500, err
	}

	_, err = db.ExecContext(ctx, `INSERT INTO orders(order_id, username, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5);`, order.OrderId, login, order.Status, order.Accrual, time.Now().Format(time.RFC3339))
	if err != nil {
		return 500, err
	}
	return response.StatusCode, nil
}

type Balance struct {
	Current   float32 `json:"current" db:"balance"`
	Withdrawn float32 `json:"withdrawn" db:"withdrawn"`
}

func GetBalance(ctx context.Context, login string) (Balance, error) {
	res := Balance{}
	if err := db.GetContext(ctx, res, `SELECT balance, withdrawn FROM users WHERE username = $1`, &login); err != nil {
		return Balance{}, err
	}
	log.Info().Msg("check")
	return res, nil
}

type Withdrawals struct {
	Order       string    `json:"order" db:"order_id"`
	Sum         float32   `json:"sum" db:"withdraw_sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

func GetWithdrawals(ctx context.Context, login string) ([]Withdrawals, error) {
	withdrawals := []Withdrawals{}
	rows, err := db.QueryxContext(ctx, `SELECT order_id, withdraw_sum, processed_at FROM withdrawals WHERE username = $1`, login)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var v Withdrawals
		err = rows.Scan(&v.Order, &v.Sum, &v.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, v)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].ProcessedAt.Before(withdrawals[j].ProcessedAt)
	})

	return withdrawals, nil
}

type Orders struct {
	Number     string    `json:"number" db:"order_id"`
	Status     string    `json:"status" db:"status"`
	Accrual    float32   `json:"accrual" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

func GetOrders(ctx context.Context, login string) ([]Orders, error) {
	orders := []Orders{}
	rows, err := db.QueryxContext(ctx, `SELECT order_id, status, accrual, uploaded_at FROM orders WHERE username = $1`, login)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var v Orders
		err = rows.Scan(&v.Number, &v.Status, &v.Accrual, &v.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, v)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})
	return orders, nil
}

type Wdraw struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func Withdraw(ctx context.Context, login string, wdraw *Wdraw) (int, error) {
	order := wdraw.Order
	withdraw := wdraw.Sum

	ok, err := Luna([]byte(order))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if !ok {
		return http.StatusUnprocessableEntity, nil
	}

	balance, err := GetBalance(ctx, login)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if balance.Current < withdraw {
		return http.StatusPaymentRequired, nil
	}

	balance.Current = balance.Current - withdraw
	balance.Withdrawn = balance.Withdrawn + withdraw

	tx, err := db.Beginx()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	_, errExec := tx.ExecContext(ctx, `UPDATE users SET balance = $1, withdrawn = $2 WHERE username = $3`, balance.Current, balance.Withdrawn, login)
	if errExec != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Err(rollbackErr).Msg("rollback error")
			return http.StatusInternalServerError, err
		}
		log.Err(errExec).Msg("update error")
		return http.StatusInternalServerError, err
	}
	if err := tx.Commit(); err != nil {
		log.Err(err).Msg("commit error")
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
