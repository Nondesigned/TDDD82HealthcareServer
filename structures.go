package main

//Login user structure
type Login struct {
	Card     int    `json:"card" binding:"required"`
	Password string `json:"password" binding:"required"`
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
