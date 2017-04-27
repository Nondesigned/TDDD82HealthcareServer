package main

import (
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
	"golang.org/x/crypto/sha3"
)

//SaltSize sets length of salt
const SaltSize = 16

func main() {
	r := gin.Default()
	auth := r.Group("/")
	admin := r.Group("/")
	// Handlers
	r.GET("/", DefaultHandler)
	r.POST("/login", LoginHandler)
	r.POST("/create", CreateUserHandler)

	//Handlers that requires authentication
	auth.Use(AuthReq())
	{
		auth.GET("/contacts", GetContactsHandler)
		auth.GET("/pins", GetPinsHandler)
		auth.POST("/pins", CreatePinHandler)
		auth.POST("/deletepin", DeletePinHandler)
		auth.GET("/groups", GetGroupsHandler)
	}

	admin.Use(AdminReq())
	{
		admin.GET("/users", GetUsersHandler)
		admin.GET("/contacts/:number", GetNumberContactsHandler)
		admin.GET("/groups/:number", GetNumberGroupsHandler)
		admin.DELETE("/groups/:source/:destination", DeleteFromGroupHandler)
		admin.PUT("/groups/:source/:destination", PutUserInGroupHandler)
		admin.GET("/ngroups/:number", GetNonMemberGroupsHandler)
		admin.GET("/pins/:number", GetAdminPinsHandler)
	}
	r.Static("/site", "site/")
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
func AdminReq() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, _ := c.Request.Cookie("AdminToken")
		token := cookie.String()
		if token == "AdminToken=adminkaffekaka" {
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

//GetNumber : Returns phonenr from token
func GetNumber(token string) int {
	j, err := jws.ParseJWT([]byte(token))
	if err != nil {
		return 0
	}

	checkErr(err)

	number, err := strconv.Atoi((j.Claims().Get("sub").(string)))
	if err != nil {
		return 0
	}
	return number
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
	phonenr := GetPhoneNumberForToken(user.Card)
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

func HasGroup(number int, groupid string) bool {
	for _, group := range GetGroups(number) {
		if group.Id == groupid {
			return true
		}
	}
	return false
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

//InsertFCMToken insert unique fcmtoken for each client into the mysql database on login
func InsertFCMToken(user Login) bool {
	phonenr := GetPhoneNumberForToken(user.Card)

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

func GetGroups(number int) []*Group {
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	checkErr(err)

	rows, err := db.Query("SELECT DISTINCT(usergroup.id), usergroup.name FROM usergroup, groupmember WHERE groupmember.user_number=? AND groupmember.group_id = usergroup.id;", number)
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		p := new(Group)
		if err := rows.Scan(&p.Id, &p.Name); err != nil {
			return nil
		}
		groups = append(groups, p)
	}
	if err := rows.Err(); err != nil {
		return nil
	}

	return groups
}

func GetNonMemberGroups(number int) []*Group {
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	checkErr(err)

	rows, err := db.Query("SELECT distinct(usergroup.id), usergroup.name FROM healthcare.usergroup WHERE usergroup.id NOT IN (SELECT usergroup.id FROM healthcare.usergroup, healthcare.groupmember WHERE usergroup.id = groupmember.group_id AND groupmember.user_number=?);", number)
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		p := new(Group)
		if err := rows.Scan(&p.Id, &p.Name); err != nil {
			return nil
		}
		groups = append(groups, p)
	}
	if err := rows.Err(); err != nil {
		return nil
	}

	return groups
}

//GetPhoneNumber retrieves phonenumber for input NFC id
func GetPhoneNumberForToken(card int) int {
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
	buf := []byte(str)
	h := make([]byte, 256)
	sha3.ShakeSum256(h, buf)
	return h
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
	token := c.Request.Header.Get("Token")

	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	stmtOut, err := db.Prepare("SELECT DISTINCT name, phonenumber FROM user INNER JOIN groupmember ON phonenumber = user_number WHERE group_id IN (SELECT group_id FROM groupmember WHERE user_number = ?) AND NOT phonenumber = ?")
	checkErr(err)
	defer stmtOut.Close()

	phonenr := GetNumber(token)

	rows, err := stmtOut.Query(phonenr, phonenr)
	checkErr(err)

	var contacts []*Contacts
	for rows.Next() {
		p := new(Contacts)
		if err := rows.Scan(&p.Name, &p.Phonenumber); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		contacts = append(contacts, p)
	}
	if err := rows.Err(); err != nil {
		//c.AbortWithStatus(http.StatusInternalServerError)
		var a []int
		c.JSON(http.StatusAccepted, a)
	}
	c.JSON(http.StatusAccepted, contacts)

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

//GetPinsHandler : Returns pins to the user
func GetPinsHandler(c *gin.Context) {
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	checkErr(err)
	token := c.Request.Header.Get("Token")

	number := GetNumber(token)
	rows, err := db.Query("SELECT healthcare.marking.id,healthcare.marking.type, healthcare.marking.longitude, healthcare.marking.latitude FROM healthcare.marking, healthcare.groupmember, healthcare.user where marking.group_id = groupmember.group_id and groupmember.user_number = user.phonenumber and user.phonenumber = ?;", number)
	defer rows.Close()

	var pin []*Pin
	for rows.Next() {
		p := new(Pin)
		if err := rows.Scan(&p.Id, &p.Type, &p.Long, &p.Lat); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		pin = append(pin, p)
	}
	if err := rows.Err(); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.JSON(http.StatusAccepted, pin)
}

//CreatePinHandler : Inset pin and returns id
func CreatePinHandler(c *gin.Context) {
	var newPin NewPin
	err := c.BindJSON(&newPin)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val
	token := c.Request.Header.Get("Token")
	number := GetNumber(token)
	if HasGroup(number, newPin.GroupId) == false {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	rows, err := db.Query("INSERT INTO `healthcare`.`marking` (`group_id`, `type`, `creation_time`, `latitude`, `longitude`) VALUES (?, ?, NOW(), ?, ?);", newPin.GroupId, newPin.Type, newPin.Lat, newPin.Long)
	defer rows.Close()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "ok"})
}

//DeletePinHandler : Delete pin
func DeletePinHandler(c *gin.Context) {
	var pin EditPin
	err := c.BindJSON(&pin)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val
	token := c.Request.Header.Get("Token")
	number := GetNumber(token)
	stat, err := db.Prepare("DELETE healthcare.marking.* FROM healthcare.marking,healthcare.groupmember WHERE marking.group_id = groupmember.group_id AND marking.id = ? AND groupmember.user_number = ?;")
	defer stat.Close()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	res, err := stat.Exec(pin.Id, number)
	affected, _ := res.RowsAffected()
	if err != nil || (affected < int64(1)) {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "ok"})
}

//GetGroupsHandler : Returns the groups available to the user
func GetGroupsHandler(c *gin.Context) {
	token := c.Request.Header.Get("Token")
	number := GetNumber(token)

	if groups := GetGroups(number); groups != nil {
		c.JSON(http.StatusAccepted, groups)
	} else {
		c.JSON(http.StatusAccepted, gin.H{"status": "No groups found"})
	}
}
