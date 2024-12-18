package main

import (
	"flychat/controller"
	"flychat/model"
	"flychat/platform"
	"flychat/service"
	"fmt"
	"os"
	"time"

	_uuid "github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// CORSMiddleware ...
// CORS (Cross-Origin Resource Sharing)
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			fmt.Println("OPTIONS")
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

// RequestIDMiddleware ...
// Generate a unique ID and attach it to each request for future reference or use
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := _uuid.New()
		c.Writer.Header().Set("X-Request-Id", uuid.String())
		c.Set("requestId", uuid.String())
		c.Next()
	}
}

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		requestId := c.GetString("requestId")

		logrus.Infof(
			" [%s] %d | %v | %s | %s | %s | %s ",
			requestId,
			status,
			latency,
			clientIP,
			method,
			path,
			userAgent,
		)

		if len(raw) > 0 {
			path = path + "?" + raw
		}
	}
}

var auth = new(controller.AuthController)

// TokenAuthMiddleware ...
// JWT Authentication middleware attached to each request that needs to be authenitcated to
// validate the access_token in the header
func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth.TokenValid(c)
		c.Next()
	}
}

func main() {
	fmt.Println("Server started...")

	//Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("failed to load the env file")
	}

	platform.InitFile("./log", "gin")

	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(RequestIDMiddleware())
	r.Use(LogMiddleware())

	//init database
	platform.InitDB()
	model.InstallDB()

	platform.InitLLMClient()

	v1 := r.Group("/v1")
	{
		user := new(controller.UserController)
		v1.POST("/user/register", user.Register)
		v1.POST("/user/login", user.Login)
		//v1.GET("/user/logout", user.Logout)

		//Refresh the token
		v1.POST("/token/refresh", auth.Refresh)

		// Summary
		chat := new(controller.ChatController)
		v1.POST("/test", TokenAuthMiddleware(), chat.Test)
		v1.POST("/hsummary", chat.HSummary)
		v1.POST("/summary", TokenAuthMiddleware(), chat.Summary)
	}

	c := cron.New()
	c.AddFunc("39 17 * * *", func() {
		_, _ = service.StartHSummary(5)
		service.SendEMail()
	})
	c.Start()

	port := os.Getenv("PORT")
	r.Run(":" + port)
}
