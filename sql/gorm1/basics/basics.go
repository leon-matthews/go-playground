// Basic setup and CRUD operations using GORM version 1.
// https://v1.gorm.io/docs/
package main

import "fmt"

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Product contains fields for the database table
//
// This model includes a `gorm.DeletedAt` field (from `gorm.Model`), so will get
// 'soft delete' behaviour:  the record WON’T be removed from the database, but
// GORM will set the DeletedAt‘s value to the current time, and the data is not
// findable with normal Query methods anymore.
//
// You can find and delete matched records permanently with `db.Unscoped()`
type Product struct {
	gorm.Model // Embed for ID, CreatedAt, UpdatedAt, and DeletedAt
	Name       string
	Code       string
	Price      uint
}

func main() {
	db, err := gorm.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		panic("failed to connect to database")
	}
	defer db.Close()

	// Migrate schema
	db.AutoMigrate(&Product{})

	// Create record
	db.Create(&Product{
		Name:  "Shoes",
		Code:  "D42",
		Price: 112,
	})

	// Fetch by PK
	var product Product
	db.First(&product, 1)

	// Fetch by WHERE
	db.First(&product, "code=?", "D42")

	// Update single field
	db.Model(&product).Update("Price", 200)

	// Delete
	db.Delete(&product)
	fmt.Printf("[%T]%+[1]v\n", product)
}
