package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

//SaltSize sets length of salt
const SaltSize = 16

func main() {
	r := gin.Default()
	auth := r.Group("/")
	// Handlers
	r.GET("/", DefaultHandler)
	r.POST("/login", LoginHandler)
	r.GET("/create", CreateUserHandler)

	//Handlers that requires authentication
	auth.Use(AuthReq())
	{
		auth.GET("/contacts", GetContactsHandler)
	}

	r.RunTLS(":8080", "cert.pem", "key.unencrypted.pem")
}

//AuthReq : Middleware for authentication
func AuthReq() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Token")
		if ValidateToken(token) {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
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

	stmtOut, err := db.Prepare("SELECT password, salt FROM user WHERE NFC_id = ?")
	if err != nil {
		panic(err.Error())
	}
	defer stmtOut.Close()
	var salt string
	var password string

	//Retrieves password and salt from the DB for chosen user
	err = stmtOut.QueryRow(user.Card).Scan(&password, &salt)
	checkErr(err)

	//Hashes the login password with the users salt and converts to string
	hashedPW := hex.EncodeToString(SHA3(user.Password + salt))

	return hashedPW == password

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

//CreateToken : Creates Token
func CreateToken() string {
	//Create the Claims for the JWT
	claims := jws.Claims{}
	claims.SetExpiration(time.Now().AddDate(1, 0, 0))
	claims.SetIssuer("Sjukvårdsgruppen")
	claims.SetSubject("TDDD82Login")
	claims.SetAudience("mobile")
	claims.SetIssuedAt(time.Now())

	//Sign it with the privatekey and return it
	bytes, _ := ioutil.ReadFile("key.unencrypted.pem")
	rsaPrivate, _ := crypto.ParseRSAPrivateKeyFromPEM(bytes)
	jwt := jws.NewJWT(claims, crypto.SigningMethodRS256)
	b, _ := jwt.Serialize(rsaPrivate)
	return string(b)

}

//ValidateToken : Validates Token
func ValidateToken(token string) bool {
	j, err := jws.ParseJWT([]byte(token))
	if err != nil {
		return false
	}
	bytes, _ := ioutil.ReadFile("cert.pem")
	publicKey, err := crypto.ParseRSAPublicKeyFromPEM(bytes)
	if err != nil {
		return false
	}
	if err = j.Validate(publicKey, crypto.SigningMethodRS256); err == nil {
		return true
	}
	return false

}

/*------------ Handlers -------------*/

//DefaultHandler : Handler for root
func DefaultHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "running"})
}

//GetContactsHandler : Return contacts for the logged-in user
func GetContactsHandler(c *gin.Context) {
	var user Login
	err := c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

//LoginHandler  : Handler for the login
func LoginHandler(c *gin.Context) {
	//Bind JSON and check if cridentials matches
	var user Login
	err := c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	if ValidateUser(user) == true {
		c.JSON(http.StatusOK, gin.H{"status": "accepted", "token": CreateToken()})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Wrong cridentials"})
	}
}

//CreateUserHandler : Handler for user creation
func CreateUserHandler(c *gin.Context) {

	pass := saltedHash("kaffekaka")

	c.JSON(http.StatusOK, gin.H{"pass": pass})
}

//Generates a salt and hashes it with the password
func saltedHash(secret string) string {

	var nfcid int
	var usrname string
	var hashpw string

	//Temporary variables
	nfcid = 123
	usrname = "Markus Johansson"

	phonenr := rand.Intn(99999999)

	//Generates a random salt of length SaltSize
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, SaltSize)
	for i := 0; i < SaltSize; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	salt := string(result)

	//Hashes password + salt and converts to string
	hashpw = hex.EncodeToString(SHA3(secret + salt)) // converts hex to string

	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+"@/"+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	stmtOut, err := db.Prepare("INSERT INTO user (name, NFC_id, password, salt, phonenumber) VALUES (?, ? ,? ,? ,?)")
	checkErr(err)
	defer stmtOut.Close()

	_, err = stmtOut.Exec(usrname, nfcid, hashpw, salt, phonenr)
	checkErr(err)

	return hashpw
}

//SHA3 Converts input to SHA3 hash
func SHA3(str string) []byte {

	bytes := []byte(str)

	h := sha256.New()  // new sha256 object
	h.Write(bytes)     // data is now converted to hex
	code := h.Sum(nil) // code is now the hex sum

	return code
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
