package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/lekan/gophermart/internal/cfg"
	"github.com/lekan/gophermart/internal/logger"
	"github.com/lekan/gophermart/internal/luhn"
	_ "github.com/lib/pq"
	"github.com/omeid/pgerror"
	"golang.org/x/sync/errgroup"
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
	balance NUMERIC DEFAULT 0,
	withdrawn NUMERIC DEFAULT 0, 
	PRIMARY KEY (username)
);`

var orders = `
CREATE TABLE IF NOT EXISTS orders(
	order_id VARCHAR UNIQUE NOT NULL,
	username VARCHAR NOT NULL, 
	status VARCHAR DEFAULT '',
	accrual NUMERIC DEFAULT 0,
	uploaded_at TIMESTAMP,
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
	PRIMARY KEY (operation_id);
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

type Order struct {
	OrderId string  `json:"order" db:"order_id"`
	Status  string  `json:"status" db:"status"`
	Accrual float32 `json:"accrual,omitempty" db:"accrual"`
}

func worker(url string, orderCh chan Order) error {
	var order Order
	for i := 0; i < 5; i++ {
		res, err := http.Get(url)
		if err != nil {
			log.Err(err).Msg("goroutine get error")
			return err
		}

		log.Info().Msgf("in worker: %s", res.Status)
		if res.StatusCode == http.StatusNoContent {
			order = Order{}
			break
		}

		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			if err = json.NewDecoder(res.Body).Decode(&order); err != nil {
				log.Err(err).Msg("in goroutine json error")
				return err
			}
		}

		if order.Status == "PROCESSED" || order.Status == "INVALID" {
			break
		}
	}
	orderCh <- order
	return nil
}

func PostOrder(ctx context.Context, login string, orderId []byte) (int, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	number, err := strconv.Atoi(string(orderId))
	if err != nil {
		log.Err(err).Msg("order must be a number")
		return http.StatusInternalServerError, err
	}

	ok := luhn.Valid(number)
	if !ok {
		log.Info().Msg("wrong order number format")
		return http.StatusUnprocessableEntity, nil
	}

	var other string

	if err := db.GetContext(ctx, &other, `SELECT username FROM orders WHERE order_id=$1`, string(orderId)); err != nil {
		if errors.Is(err, pgerror.NoDataFound(err)) {
			log.Err(err).Msg("its ok")
		}
	}

	log.Info().Msgf("other: %s  login: %s", other, login)

	if other != "" && other != login {
		return http.StatusConflict, nil
	}

	if other != "" && other == login {
		return http.StatusOK, nil
	}

	_, err = db.ExecContext(ctx, `INSERT INTO orders
    (order_id, username, status, uploaded_at) 
    VALUES ($1, $2, $3, $4);`,
		string(orderId), login, "NEW", time.Now().Format(time.RFC3339))
	if err != nil {
		if errors.Is(err, pgerror.UniqueViolation(err)) {
			log.Err(err).Msg("Unique Violation")
			return http.StatusOK, err
		}
		log.Err(err).Msg("add order error")
		return http.StatusInternalServerError, err
	}

	errGr, _ := errgroup.WithContext(ctx)
	url := cfg.GetAccuralSystemAddress() + "/api/orders/" + string(orderId)
	orderCh := make(chan Order, 1)

	errGr.Go(func() error {
		return worker(url, orderCh)
	})

	err = errGr.Wait()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	order := Order{}
	order = <-orderCh

	log.Info().Msgf("%+v", order)
	//if order.OrderId == "" {
	//	log.Info().Msg("StatusNoContent")
	//	return http.StatusNoContent, errors.New("no content")
	//}

	if order.Status == "PROCESSED" {
		_, err = db.ExecContext(ctx, `UPDATE orders SET status=$1, accrual=$2, uploaded_at=$3 WHERE order_id=$4 AND username=$5;`,
			order.Status, order.Accrual, time.Now().Format(time.RFC3339), order.OrderId, login)
		if err != nil {
			log.Err(err).Msg("database update error")
			return http.StatusInternalServerError, err
		}

		_, err = db.ExecContext(ctx, `UPDATE users SET balance=balance+$1 WHERE username=$2`, order.Accrual, login)
		if err != nil {
			log.Err(err).Msg("user balance update error")
			return http.StatusInternalServerError, err
		}

	} else {
		_, err = db.ExecContext(ctx, `UPDATE orders SET status=$1, uploaded_at=$2 WHERE order_id=$3 AND username=$4`, order.Status, time.Now().Format(time.RFC3339), order.OrderId, login)
		if err != nil {
			log.Err(err).Msg("database update error")
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusAccepted, nil
}

type Balance struct {
	Current   float32 `json:"current" db:"balance"`
	Withdrawn float32 `json:"withdrawn" db:"withdrawn"`
}

func GetBalance(ctx context.Context, login string) (Balance, error) {
	res := Balance{}
	log.Info().Msgf("balance login: %s", login)
	if err := db.GetContext(ctx, &res, `SELECT balance, withdrawn FROM users WHERE username = $1`, login); err != nil {
		return Balance{}, err
	}
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
	Status     string    `json:"status,omitempty" db:"status"`
	Accrual    float32   `json:"accrual,omitempty" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

func GetOrders(ctx context.Context, login string) ([]Orders, error) {
	orders := []Orders{}

	rows, err := db.QueryxContext(ctx, `SELECT order_id, status, accrual, uploaded_at FROM orders WHERE username = $1`, login)
	if err != nil {
		log.Err(err).Msg("in GetOrder query error")
		return nil, err
	}

	for rows.Next() {
		var v Orders
		err = rows.Scan(&v.Number, &v.Status, &v.Accrual, &v.UploadedAt)
		if err != nil {
			log.Err(err).Msg("in GetOrders scan error")
			return nil, err
		}
		orders = append(orders, v)
	}

	err = rows.Err()
	if err != nil {
		log.Err(err).Msg("in GetOrders rows.Err()")
		return nil, err
	}

	if len(orders) == 0 {
		return nil, errors.New("204 StatusNoContent")
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

	number, err := strconv.Atoi(order)
	if err != nil {
		log.Err(err).Msg("order must to be a number")
		return http.StatusInternalServerError, err
	}

	ok := luhn.Valid(number)

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

	_, errUser := db.ExecContext(ctx, `
UPDATE users 
SET balance = $1, withdrawn = $2 
WHERE username = $3;`, balance.Current, balance.Withdrawn, login)
	if errUser != nil {
		log.Err(errUser).Msg("user balance update error")
		return http.StatusInternalServerError, errUser
	}

	_, errWdwl := db.ExecContext(ctx, `
INSERT INTO withdrawals(username, order_id, withdraw_sum, processed_at)
VALUES ($1, $2, $3, $4);`, login, order, withdraw, time.Now().Format(time.RFC3339))
	if errWdwl != nil {
		log.Err(errWdwl).Msg("withdrawals error")
		return http.StatusInternalServerError, errWdwl
	}

	return http.StatusOK, nil
}
