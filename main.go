package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func fetchPrice(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	price := doc.Find(".product-price__big").First().Text()

	if price == "" {
		return "цена не найдена", nil
	}

	return price, nil
}

func main() {
	url := "https://rozetka.com.ua/ua/universalnye-mobilnye-batarei-qinetiq-152025260/p549129762/"
	price, err := fetchPrice(url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Цена:", price)
}
