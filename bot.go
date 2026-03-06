package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var botInstance *tgbotapi.BotAPI

func cleanPrice(price string) string {
	var result strings.Builder
	for _, r := range price {
		if unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func startPriceChecker() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			products := getAllProducts()
			for _, p := range products {
				newPriceRaw, err := fetchPrice(p.URL)
				if err != nil {
					continue
				}

				newPrice := cleanPrice(newPriceRaw)
				oldPrice := cleanPrice(p.Price)

				newVal, err1 := strconv.Atoi(newPrice)
				oldVal, err2 := strconv.Atoi(oldPrice)

				if err1 != nil || err2 != nil {
					continue
				}

				if newVal < oldVal {
					chatID := getChatIDByURL(p.URL)
					updatePrice(chatID, p.URL, newPriceRaw)
					msg := fmt.Sprintf("🔥 Знижка!\n%s\nБула: %s₴ → Стала: %s₴", p.URL, oldPrice, newPrice)
					botInstance.Send(tgbotapi.NewMessage(chatID, msg))
				}
			}
		}
	}()
}

func main() {
	godotenv.Load()
	token := os.Getenv("BOT_TOKEN")

	initDB()

	var err error
	botInstance, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Бот запущен: %s", botInstance.Self.UserName)

	startPriceChecker()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := botInstance.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message.Text
		chatID := update.Message.Chat.ID

		if msg == "/start" {
			botInstance.Send(tgbotapi.NewMessage(chatID, "Привіт! Кидай посилання на товар з Rozetka"))
			continue
		}

		if msg == "/list" {
			products := getProducts(chatID)
			if len(products) == 0 {
				botInstance.Send(tgbotapi.NewMessage(chatID, "Список порожній"))
				continue
			}

			total := 0
			var lines []string
			for _, p := range products {
				lines = append(lines, fmt.Sprintf("%s\nЦіна: %s", p.URL, p.Price))
				val, err := strconv.Atoi(cleanPrice(p.Price))
				if err == nil {
					total += val
				}
			}
			lines = append(lines, fmt.Sprintf("\n💰 Загальна вартість: %d₴", total))
			botInstance.Send(tgbotapi.NewMessage(chatID, strings.Join(lines, "\n\n")))
			continue
		}

		if strings.HasPrefix(msg, "https://rozetka.com.ua") {
			price, err := fetchPrice(msg)
			if err != nil {
				botInstance.Send(tgbotapi.NewMessage(chatID, "Помилка"))
				continue
			}

			saveProduct(chatID, msg, price)
			botInstance.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Збережено! Ціна: %s", price)))
			continue
		}

		botInstance.Send(tgbotapi.NewMessage(chatID, "Надішли посилання з rozetka.com.ua"))
	}
}
