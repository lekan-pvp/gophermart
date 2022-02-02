package models

import (
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

func (account *Account) Validate() (map[string]interface{}, bool) {
	if len(account.Password) < 6 {
		return utils.Message(false, "Password is required"), false
	}

	temp := &Account{}
	//проверка на наличие ошибок и дубликатов
	err := GetDB().Table("users").Where("login = ?", account.Login).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return utils.Message(false, "Connection error. Please retry"), false
	}

	if temp.Login != "" {
		return utils.Message(false, "Login already in use by another user."), false
	}

	return utils.Message(false, "Requirement passed"), true
}

func (account *Account) Create() map[string]interface{} {
	if resp, ok := account.Validate(); !ok {
		return resp
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)

	GetDB().Create(account)
	if account.ID <= 0 {
		return utils.Message(false, "Failed to create account, connection error")
	}

	// создаем новый токен JWT для новой учетной записи
	tk := &Token{UserID: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString
	resp := utils.Message(true, "Logged in")
	resp["account"] = account
	return resp
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
