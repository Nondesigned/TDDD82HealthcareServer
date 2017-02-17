package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	r := gin.Default()
	auth := r.Group("/")
	// Handlers
	r.GET("/", DefaultHandler)
	r.POST("/login", LoginHandler)
	r.GET("/create", CreateUserHandler)

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
func ValidateUser(user Login) bool {
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+"@/"+DBName)
	if err != nil {
		return false
	}
	defer db.Close() //Close DB after function has returned a val

	stmtOut, err := db.Prepare("SELECT password FROM user WHERE NFC_id = ?")
	if err != nil {
		panic(err.Error())
	}
	defer stmtOut.Close()

	var password string
	err = stmtOut.QueryRow(user.Card).Scan(&password)
	if err != nil {
		panic(err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(user.Password))
	if err == nil {
		return true
	}
	return false

}

//GetSettings : Returns the settings for the DB
func GetSettings() (string, string, string) {
	var settings = new(Settings)
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	err = json.Unmarshal(raw, &settings)
	if err != nil {
		return "", "", ""
	}
	return settings.DBUser, settings.DBPass, settings.DBName
}

//ValidateToken : Validates Token
func ValidateToken(token Login) bool {
	return false

}

/*------------ Handlers -------------*/

//DefaultHandler : Handler for root
func DefaultHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "running"})
}

//LoginHandler  : Handler for the login
func LoginHandler(c *gin.Context) {
	var user Login
	err := c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	if ValidateUser(user) == true {
		c.JSON(http.StatusOK, gin.H{"status": "accepted"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Wrong cridentials"})
	}
}

//CreateUserHandler : Handler for user creation
func CreateUserHandler(c *gin.Context) {
	pass, err := bcrypt.GenerateFromPassword([]byte("kaffekaka"), 10)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, gin.H{"pass": string(pass)})
}
