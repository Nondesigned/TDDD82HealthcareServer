package main

type User struct {
	Name   string `json:"name"`
	Card   string `json:"card"`
	Number string `json:"number"`
}

//Login user structure
type Login struct {
	Card     int    `json:"card" binding:"required"`
	Password string `json:"password" binding:"required"`
	FCMToken string `json:"fcmtoken" binding:"required"`
}

//Settings struct
type Settings struct {
	DBUser string `json:"DBUser"`
	DBPass string `json:"DBPass"`
	DBName string `json:"DBName"`
}

//Token struct
type Token struct {
	Token string `json:"token" binding:"required"`
}

//Create struct
type Create struct {
	Card        int    `json:"card" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Phonenumber int    `json:"phonenumber" binding:"required"`
}

type Pin struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Long string `json:"long"`
	Lat  string `json:"lat"`
}
type EditPin struct {
	Id      string `json:"id"`
	GroupId string `json:"groupid"`
}

type NewPin struct {
	Type    string `json:"type"`
	Long    string `json:"long"`
	Lat     string `json:"lat"`
	GroupId string `json:"groupid"`
}

type Contacts struct {
	Name        string `json:"name"`
	Phonenumber string `json:"phonenumber"`
}
type Group struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
