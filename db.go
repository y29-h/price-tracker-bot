package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./products.db")
	if err != nil {
		log.Fatal(err)
	}

	// Створюємо таблицю якщо не існує
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER,
		url TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func saveProduct(chatID int64, url string) {
	db.Exec("INSERT INTO products (chat_id, url) VALUES (?, ?)", chatID, url)
}

func getProducts(chatID int64) []string {
	rows, err := db.Query("SELECT url FROM products WHERE chat_id = ?", chatID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var products []string
	for rows.Next() {
		var url string
		rows.Scan(&url)
		products = append(products, url)
	}
	return products
}
