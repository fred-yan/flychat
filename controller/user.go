package controller

import (
	"flychat/service"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// UserController ...
type UserController struct{}

var userService = service.UserService{}

func (ctrl UserController) Register(c *gin.Context) {
	// 1. 添加日志记录
	logger := logrus.New()
	logger.Info("Handling user registration request")

	// 2. 获取请求参数
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Error("Invalid input: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 3. 处理用户注册逻辑
	user := &service.User{
		Username: input.Username,
		Password: input.Password,
		Email:    input.Email,
	}
	if err := userService.Register(user); err != nil {
		logger.Error("Failed to register user: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// 4. 返回成功响应
	logger.Info("User registered successfully")
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func (ctrl UserController) Login(c *gin.Context) {
	// 1. 获取请求参数
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

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
