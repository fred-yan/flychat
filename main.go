package main

import (
	"flychat/controller"
	"flychat/model"
	"flychat/platform"
	"fmt"
	"log"
	"os"
	"time"

	_uuid "github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
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
		c.Next()
	}
}

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		logrus.Infof(
			"| %3d | %13v | %15s | %4s | %33s | %s |",
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

func main() {
	fmt.Println("Server started...")

	//Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error: failed to load the env file")
	}

	//Start the default gin server
	r := gin.Default()

	r.Use(CORSMiddleware())
	r.Use(RequestIDMiddleware())
	r.Use(LogMiddleware())

	//Start database
	platform.InitDB()
	model.InstallDB()

	v1 := r.Group("/v1")
	{
		/*** START USER ***/
		user := new(controller.UserController)

		v1.POST("/user/login", user.Login)
		v1.POST("/user/register", user.Register)
		//v1.GET("/user/logout", user.Logout)

		/*** START AUTH ***/
		auth := new(controller.AuthController)

		//Refresh the token when needed to generate new access_token and refresh_token for the user
		v1.POST("/token/refresh", auth.Refresh)
	}

	port := os.Getenv("PORT")
	r.Run(":" + port)
}
