package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
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

//GetNumberContactsHandler : Returns contacts based on phone number
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

//GetNumberGroupsHandler : Returns groups based on phone number
func GetNumberGroupsHandler(c *gin.Context) {
	phonenr := c.Param("number")
	number, _ :=  strconv.Atoi(phonenr); 
	var groups = GetGroups(number);

	c.JSON(http.StatusAccepted, groups)
}

//GetNonMemberGroupsHandler : Returns groups based on phone number
func GetNonMemberGroupsHandler(c *gin.Context) {
	phonenr := c.Param("number")
	number, _ :=  strconv.Atoi(phonenr)
	var groups = GetNonMemberGroups(number)
	c.JSON(http.StatusAccepted, groups)
}

//DeleteUserFromGroup : Delete user from group
func DeleteUserFromGroup(src string, dst string) bool{
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val
	stat, err := db.Prepare("DELETE healthcare.groupmember.* FROM healthcare.groupmember WHERE groupmember.group_id = ? AND groupmember.user_number = ?;")
	defer stat.Close()
	if err != nil {
		return false;		
	}
	res, err := stat.Exec(dst, src)
	affected, _ := res.RowsAffected()
	if err != nil || (affected < int64(1)) {
		return false;
	}
	return true;
}

//DeleteFromGroupHandler : Delete user from group
func DeleteFromGroupHandler(c *gin.Context) {
	src := c.Param("source")
	dst := c.Param("destination")
	if(DeleteUserFromGroup(src, dst)){
		c.AbortWithStatus(200);
	}else{
		c.AbortWithStatus(http.StatusInternalServerError);
	}
}
//PutUserInGroup : Put User in group
func PutUserInGroup(src string, dst string) bool{

	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val
	stat, err := db.Prepare("INSERT INTO healthcare.groupmember (groupmember.user_number, groupmember.group_id) VALUES (?,?);")

	defer stat.Close()
	if err != nil {
		println("syntax error")
		return false;		
	}
	res, err := stat.Exec(src, dst)
	affected, _ := res.RowsAffected()
	if err != nil || (affected < int64(1)) {
		println("Not affected")
		return false;
	}
	return true;
}

//PutFromInndler : Put User in group
func PutUserInGroupHandler(c *gin.Context) {
	src := c.Param("source")
	dst := c.Param("destination")
	if(PutUserInGroup(src, dst)){
		c.AbortWithStatus(200);
	}else{
		c.AbortWithStatus(http.StatusInternalServerError);
	}
}

//GetAdminPinsHandler : Really just a copy of the GetPinsHandler. Lift out the sql later
func GetAdminPinsHandler(c *gin.Context){
	phonenr := c.Param("number")
	number, _ :=  strconv.Atoi(phonenr)
	DBUser, DBPass, DBName := GetSettings()
	db, err := sql.Open("mysql", DBUser+":"+DBPass+DBName)
	checkErr(err)
	defer db.Close() //Close DB after function has returned a val

	checkErr(err)

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