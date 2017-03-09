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
	"strconv"
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
	r.POST("/create", CreateUserHandler)

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
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	stmtOut, err := db.Prepare("SELECT password, salt FROM user WHERE NFC_id = ?")
	checkErr(err)

	defer stmtOut.Close()

	var salt string
	var password string

	//Retrieves password and salt from the DB for chosen user
	err = stmtOut.QueryRow(user.Card).Scan(&password, &salt)
	if err != nil {
		return false
	}

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
func CreateToken(user Login) string {
	//Retrieves phonenr and converts it to string
	phonenr := GetPhoneNumber(user.Card)
	phonestring := strconv.Itoa(phonenr)
	//Create the Claims for the JWT
	claims := jws.Claims{}
	claims.SetExpiration(time.Now().AddDate(1, 0, 0))
	claims.SetIssuer("Sjukvårdsgruppen")
	claims.SetSubject(phonestring)
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
	checkErr(err)

	if err = j.Validate(publicKey, crypto.SigningMethodRS256); err == nil {
		return true
	}
	return false

}

//CheckContactList : Checks which contacts user should access
func CheckContactList() string {

	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	rows, err := db.Query("SELECT name, phonenumber FROM user")
	checkErr(err)
	defer rows.Close()

	//Retrieves the tuples (Name, Phonenumber) and creates a JSON object with those N number of tuples where N is the number of rows in the mysql user table
	columns, err := rows.Columns()
	checkErr(err)

	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePointers := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePointers[i] = &values[i]
		}
		rows.Scan(valuePointers...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}
	jsonData, err := json.Marshal(tableData)
	checkErr(err)

	return string(jsonData)

}

//CreateUser creates a user from the input JSON object
func CreateUser(user Create) bool {

	salt := Salt()
	hashedpw := Hash(user.Password, salt)

	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	stmtOut, err := db.Prepare("INSERT INTO user (name, NFC_id, password, salt, phonenumber) VALUES (?, ? ,? ,? ,?)")
	checkErr(err)
	defer stmtOut.Close()

	_, err = stmtOut.Exec(user.Name, user.Card, hashedpw, salt, user.Phonenumber)
	if err != nil {
		return false
	}
	return true
}

//Salt generates a salt and hashes it with the password
func Salt() string {
	//Generates a random salt of length SaltSize
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, SaltSize)
	for i := 0; i < SaltSize; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

//Hash password and salt together and returns the result as a string
func Hash(secret string, salt string) string {
	//Hashes password + salt and converts to string
	hashpw := hex.EncodeToString(SHA3(secret + salt)) // converts hex to string

	return hashpw
}

//InsertFCMToken insert unique fcmtoken for each client into the mysql database
func InsertFCMToken(user Login) bool {

	phonenr := GetPhoneNumber(user.Card)

	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	stmtOut, err := db.Prepare("REPLACE INTO token (owner_number, data) VALUES (?, ?)")
	checkErr(err)
	defer stmtOut.Close()

	_, err = stmtOut.Exec(phonenr, user.FCMToken)
	if err != nil {
		return false
	}
	return true

}

//GetPhoneNumber retrieves phonenumber for input NFC id
func GetPhoneNumber(card int) int {
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	stmtOut, err := db.Prepare("SELECT phonenumber FROM user WHERE NFC_id = ?")
	checkErr(err)
	defer stmtOut.Close()

	var phonenr int

	err = stmtOut.QueryRow(card).Scan(&phonenr)

	return phonenr
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

/*------------ Handlers -------------*/

//DefaultHandler : Handler for root
func DefaultHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "running"})
}

//GetContactsHandler : Return contacts for the logged-in user
func GetContactsHandler(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"status": "ok", "contact list": CheckContactList()})

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
		if InsertFCMToken(user) == true {
			c.JSON(http.StatusOK, gin.H{"status": "accepted", "token": CreateToken(user)})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Insert fcmtoken failed"})
		}

	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Wrong cridentials"})
	}
}

//CreateUserHandler : Handler for user creation
func CreateUserHandler(c *gin.Context) {
	var user Create
	err := c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	if CreateUser(user) {
		c.JSON(http.StatusOK, gin.H{"status": "accepted"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Duplicate entry of unique ID"})
	}

}
