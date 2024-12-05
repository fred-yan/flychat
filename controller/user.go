package controller

import (
	"flychat/platform"
	"flychat/service"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UserController ...
type UserController struct{}

var userService = service.UserService{}

var logger = platform.Logger

func (ctrl UserController) Register(c *gin.Context) {
	logger.Infof("[%s] Handling user registration request", c.GetString("requestId"))

	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Warnf("[%s] Invalid input, %s", c.GetString("requestId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user := &service.User{
		Username: input.Username,
		Password: input.Password,
		Email:    input.Email,
	}
	if err := userService.Register(user); err != nil {
		logger.Warnf("[%s] Failed to register user %s: %s", c.GetString("requestId"), user.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	logger.Infof("[%s] User %s registered successfully", c.GetString("requestId"), user.Username)
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func (ctrl UserController) Login(c *gin.Context) {
	logger.Infof("[%s] Handling user login request", c.GetString("requestId"))

	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	token, err := userService.Login(&service.User{
		Username: loginRequest.Username,
		Password: loginRequest.Password,
	})
	if err != nil {
		logger.Warnf("[%s] User %s failed to login: %s", c.GetString("requestId"), loginRequest.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	logger.Infof("[%s] User %s login successfully", c.GetString("requestId"), loginRequest.Username)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// isValidRequest 是一个示例函数，用于验证请求是否合法
func isValidRequest(c *gin.Context) bool {
	// 这里可以添加具体的验证逻辑，例如验证 API 密钥或 JWT 令牌
	// 示例：验证 JWT 令牌
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return false
	}
	// 解析和验证 JWT 令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("my_secret_key"), nil
	})
	if err != nil || !token.Valid {
		return false
	}
	return true
}
