package repo

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/lekan/gophermart/internal/config"
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

var log = logger.New()

//go:embed users_req.txt
var schema string

//go:embed orders.txt
var orders string

//go:embed withdrawals.txt
var withdrawals string

// New
func New(databaseURI string) error {
	db = sqlx.MustConnect("postgres", databaseURI)

	db.MustExec(schema)
	db.MustExec(orders)
	db.MustExec(withdrawals)

	log.Info().Msg("create db is done...")
	return db.Ping()
}

// Signup
func Signup(ctx context.Context, creds *Credentials) error {

	_, err := db.ExecContext(ctx, `INSERT INTO users(username, password) VALUES ($1, $2)`, creds.Login, creds.Password)
	if err != nil {
		return fmt.Errorf("409 %w", err)
	}

	return nil
}

// Signin
func Signin(ctx context.Context, creds *Credentials) error {
	temp := &Credentials{}
	if err := db.GetContext(ctx, temp, `SELECT username, password FROM users WHERE username = $1`, creds.Login); err != nil {
		return err
	}

	return nil
}

// Order
type Order struct {
	OrderID string  `json:"order" db:"order_id"`
	Status  string  `json:"status" db:"status"`
	Accrual float32 `json:"accrual,omitempty" db:"accrual"`
}

// worker
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

// PostOrder
func PostOrder(ctx context.Context, login string, orderID []byte) (int, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	number, err := strconv.Atoi(string(orderID))
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

	if err := db.GetContext(ctx, &other, `SELECT username FROM orders WHERE order_id=$1`, string(orderID)); err != nil {
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
		string(orderID), login, "NEW", time.Now().Format(time.RFC3339))
	if err != nil {
		if errors.Is(err, pgerror.UniqueViolation(err)) {
			log.Err(err).Msg("Unique Violation")
			return http.StatusOK, err
		}
		log.Err(err).Msg("add order error")
		return http.StatusInternalServerError, err
	}

	errGr, _ := errgroup.WithContext(ctx)

	url := config.GetAccrualSystemAddress() + "/api/orders/" + string(orderID)
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

	if order.Status == "PROCESSED" {
		_, err = db.ExecContext(ctx, `UPDATE orders SET status=$1, accrual=$2, uploaded_at=$3 WHERE order_id=$4 AND username=$5;`,
			order.Status, order.Accrual, time.Now().Format(time.RFC3339), order.OrderID, login)
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
		_, err = db.ExecContext(ctx, `UPDATE orders SET status=$1, uploaded_at=$2 WHERE order_id=$3 AND username=$4`, order.Status, time.Now().Format(time.RFC3339), order.OrderID, login)
		if err != nil {
			log.Err(err).Msg("database update error")
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusAccepted, nil
}

// Balance
type Balance struct {
	Current   float32 `json:"current" db:"balance"`
	Withdrawn float32 `json:"withdrawn" db:"withdrawn"`
}

// GetBalance
func GetBalance(ctx context.Context, login string) (Balance, error) {
	res := Balance{}
	log.Info().Msgf("balance login: %s", login)
	if err := db.GetContext(ctx, &res, `SELECT balance, withdrawn FROM users WHERE username = $1`, login); err != nil {
		return Balance{}, err
	}
	return res, nil
}

// Withdrawals
type Withdrawals struct {
	Order       string    `json:"order" db:"order_id"`
	Sum         float32   `json:"sum" db:"withdraw_sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

// GetWithdrawals
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

// Orders
type Orders struct {
	Number     string    `json:"number" db:"order_id"`
	Status     string    `json:"status,omitempty" db:"status"`
	Accrual    float32   `json:"accrual,omitempty" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// GetOrders
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

// Wdraw
type Wdraw struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

// Withdraw
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
