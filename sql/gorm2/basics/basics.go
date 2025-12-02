// Basic setup and CRUD operations using GORM version 2.
// https://gorm.io/docs/
package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	if err != nil {
		panic("failed to connect to database")
	}

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

	// Update multiple fields #1
	db.Model(&product).Updates(Product{Price: 499, Code: "F43"})

	// Update multiple fields #2
	db.Model(&product).Updates(map[string]any{"Price": 200, "Code": "F42"})

	// Delete
	db.Delete(&product, 1)
	fmt.Printf("[%T]%+[1]v\n", product)
}
