package models

import (
	"context"
	"github.com/golang-jwt/jwt"
	"github.com/lekan/gophermart/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"os"
)

type Token struct {
	UserID uint
	jwt.Claims
}

type Account struct {
	gorm.Model
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token" sql:"-"`
}

func (account *Account) Validate(ctx context.Context) (map[string]interface{}, bool) {
	db := ctx.Value("DB").(*gorm.DB)
	temp := &Account{}
	//проверка на наличие ошибок и дубликатов
	err := db.Table("users").Where("login = ?", account.Login).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return utils.Message(false, 500, err.Error()), false
	}

	if temp.Login != "" {
		return utils.Message(false, 409, err.Error()), false
	}

	return utils.Message(false, 200, "Requirement passed"), true
}

func (account *Account) CreateUser(ctx context.Context) map[string]interface{} {
	if resp, ok := account.Validate(ctx); !ok {
		return resp
	}

	db := ctx.Value("DB").(*gorm.DB)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.Message(false, 500, err.Error())
	}
	account.Password = string(hashedPassword)

	db.Create(account)
	if account.ID <= 0 {
		return utils.Message(false, 500, "Failed to create user, connection error.")
	}

	tk := &Token{UserID: account.ID}
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

func GetUser(u uint) *Account {
	acc := &Account{}
	GetDB().Table("users").Where("id = ?", u).First(acc)
	if acc.Login == "" {
		return nil
	}
	acc.Password = ""
	return acc
}
