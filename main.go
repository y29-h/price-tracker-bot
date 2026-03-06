package main

import (
	"net/http"
	"strings"

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

	// Визначаємо магазин по URL і беремо відповідний селектор
	var selector string
	switch {
	case strings.Contains(url, "rozetka.com.ua"):
		selector = ".product-price__big"
	case strings.Contains(url, "epicentrk.ua"):
		selector = ".product-price"
	case strings.Contains(url, "foxtrot.com.ua"):
		selector = ".cost"
	case strings.Contains(url, "prom.ua"):
		selector = ".x-money-amount"
	default:
		return "магазин не підтримується", nil
	}

	price := doc.Find(selector).First().Text()
	price = strings.TrimSpace(price)

	if price == "" {
		return "ціна не знайдена", nil
	}
	return price, nil
}
