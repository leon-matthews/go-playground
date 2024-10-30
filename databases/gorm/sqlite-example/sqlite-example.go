package main

import "fmt"

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	fmt.Println(db)
	fmt.Println(err)
}

type Product struct {
	gorm.Model
	Code  string
	Price uint
}
