package main

//Login user structure
type Login struct {
	Card     int    `json:"card" binding:"required"`
	Password string `json:"password" binding:"required"`
}
