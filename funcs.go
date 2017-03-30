package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func GetUsersHandler(c *gin.Context) {
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	checkErr(err)

	rows, err := db.Query("SELECT healthcare.user.NFC_id, healthcare.user.name,healthcare.user.phonenumber FROM  healthcare.user ;")
	defer rows.Close()

	var users []*User
	for rows.Next() {
		p := new(User)
		if err := rows.Scan(&p.Card, &p.Name, &p.Number); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		users = append(users, p)
	}
	if err := rows.Err(); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.JSON(http.StatusAccepted, users)
}

func GetNumberContactsHandler(c *gin.Context) {
	phonenr := c.Param("number")

	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	stmtOut, err := db.Prepare("SELECT DISTINCT name, phonenumber FROM user INNER JOIN groupmember ON phonenumber = user_number WHERE group_id IN (SELECT group_id FROM groupmember WHERE user_number = ?) AND NOT phonenumber = ?")
	checkErr(err)
	defer stmtOut.Close()

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
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.JSON(http.StatusAccepted, contacts)

}
