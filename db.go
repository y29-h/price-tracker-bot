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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER,
		url TEXT,
		price TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func saveProduct(chatID int64, url string, price string) {
	db.Exec("INSERT INTO products (chat_id, url, price) VALUES (?, ?, ?)", chatID, url, price)
}

func updatePrice(chatID int64, url string, price string) {
	db.Exec("UPDATE products SET price = ? WHERE chat_id = ? AND url = ?", price, chatID, url)
}

type Product struct {
	URL   string
	Price string
}

func getProducts(chatID int64) []Product {
	rows, err := db.Query("SELECT url, price FROM products WHERE chat_id = ?", chatID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		rows.Scan(&p.URL, &p.Price)
		products = append(products, p)
	}
	return products
}

func getAllProducts() []Product {
	rows, err := db.Query("SELECT url, price FROM products")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		rows.Scan(&p.URL, &p.Price)
		products = append(products, p)
	}
	return products
}

func getChatIDByURL(url string) int64 {
	var chatID int64
	db.QueryRow("SELECT chat_id FROM products WHERE url = ?", url).Scan(&chatID)
	return chatID
}
