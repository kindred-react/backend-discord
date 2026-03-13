package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$10$6OUylC8XEKrzvPDuoK7e.O51PrIwlkFbL54f/Lh8WeXdNbXTMO8f6"
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("password123"))
	if err == nil {
		fmt.Println("Password matches")
	} else {
		fmt.Println("Password does not match:", err)
	}
}
