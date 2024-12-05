package service

import (
	"errors"
	"flychat/model"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type UserService struct {
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (service *UserService) Register(user *User) error {

	// 唯一性检查
	if model.UserExists(user.Username, user.Email) {
		return errors.New("user already exists")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("internal server error")
	}

	// 存储用户信息
	newUser := &model.User{
		Username: user.Username,
		Email:    user.Email,
		Password: string(hashedPassword),
	}
	if err := model.CreateUser(newUser); err != nil {
		return errors.New("internal server error")
	}
	return nil
}

func (service *UserService) Login(user *User) (string, error) {
	// 验证用户名和密码
	registeredUser, err := model.GetUserByUsername(user.Username)
	if err != nil {
		return "", errors.New("failed to get user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(registeredUser.Password), []byte(user.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// 生成会话令牌
	ts := &TokenService{}
	token, err := ts.CreateToken(registeredUser.ID, registeredUser.Username)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return "", errors.New("failed to generate token")
	}

	return token.AccessToken, nil
}
