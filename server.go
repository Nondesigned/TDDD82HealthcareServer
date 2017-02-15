package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	auth := r.Group("/")
	// Handlers
	r.GET("/", DefaultHandler)

	auth.Use(AuthReq())
	{
		//Format: auth.GET("/", Handler)
	}

	r.RunTLS(":8080", "cert.pem", "key.unencrypted.pem")
}

//AuthReq : Middleware for authentication
func AuthReq() gin.HandlerFunc {
	return func(c *gin.Context) {
		//Make authentication
		//If fails call: c.AbortWithStatus(401)
		c.Next()
	}
}

/*------------ Functions ------------*/

//ValidateUser : Validates user
func ValidateUser() {

}

/*------------ Handlers -------------*/

//DefaultHandler : Handler for root
func DefaultHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "running"})
}

//LoginHandler  : Handler for the login
func LoginHandler(c *gin.Context) {

}

//CreateUserHandler : Handler for user creation
func CreateUserHandler(c *gin.Context) {

}
