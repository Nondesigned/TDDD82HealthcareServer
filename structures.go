package main

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
	Type string `json:"type"`
	Long string `json:"long"`
	Lat  string `json:"lat"`
}
