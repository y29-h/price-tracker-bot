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

func formatPrice(n int) string {
	s := strconv.Itoa(n)
	result := ""
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += " "
		}
		result += string(r)
	}
	return result
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
					msg := fmt.Sprintf("🔥 Знижка!\n%s\nБула: %s ₴ → Стала: %s ₴", p.URL, formatPrice(oldVal), formatPrice(newVal))
					botInstance.Send(tgbotapi.NewMessage(chatID, msg))
				}
			}
		}
	}()
}

func mainMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Мої товари"),
			tgbotapi.NewKeyboardButton("🗑 Видалити товар"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ℹ️ Як користуватись"),
		),
	)
}

func sendWelcome(chatID int64, name string) {
	text := fmt.Sprintf(
		"👋 Привіт, %s!\n\n"+
			"🛍 Я відстежую ціни на Rozetka, Епіцентр, Фокстрот та Prom.\n\n"+
			"📌 Просто кинь посилання на товар — я збережу його і слідкуватиму за ціною.\n\n"+
			"⚡️ Перевірка цін кожну годину автоматично.",
		name,
	)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = mainMenu()
	botInstance.Send(msg)
}

func sendList(chatID int64) {
	products := getProducts(chatID)
	if len(products) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 Список порожній\n\nКинь посилання на товар щоб додати його!")
		msg.ReplyMarkup = mainMenu()
		botInstance.Send(msg)
		return
	}

	total := 0
	var lines []string
	for i, p := range products {
		val, err := strconv.Atoi(cleanPrice(p.Price))
		priceStr := p.Price
		if err == nil {
			priceStr = formatPrice(val) + " ₴"
			total += val
		}
		lines = append(lines, fmt.Sprintf("%d. %s\n💵 %s", i+1, p.URL, priceStr))
	}
	lines = append(lines, fmt.Sprintf("\n💰 Загальна вартість: %s ₴", formatPrice(total)))

	msg := tgbotapi.NewMessage(chatID, strings.Join(lines, "\n\n"))
	msg.ReplyMarkup = mainMenu()
	botInstance.Send(msg)
}

func sendDeleteMenu(chatID int64) {
	products := getProducts(chatID)
	if len(products) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 Немає товарів для видалення")
		msg.ReplyMarkup = mainMenu()
		botInstance.Send(msg)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i, p := range products {
		parts := strings.Split(strings.TrimSuffix(p.URL, "/"), "/")
		name := parts[len(parts)-1]
		if len(name) > 25 {
			name = name[:25] + "..."
		}
		// Передаємо ID товару замість URL — бо Telegram ліміт 64 байти
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🗑 %d. %s", i+1, name),
			fmt.Sprintf("del:%d", p.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msg := tgbotapi.NewMessage(chatID, "Виберіть товар для видалення:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	botInstance.Send(msg)
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
		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			chatID := callback.Message.Chat.ID

			if strings.HasPrefix(callback.Data, "del:") {
				idStr := strings.TrimPrefix(callback.Data, "del:")
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err == nil {
					deleteProductByID(id)
				}

				botInstance.Request(tgbotapi.NewCallback(callback.ID, "✅ Видалено!"))
				botInstance.Request(tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID))

				msg := tgbotapi.NewMessage(chatID, "✅ Товар видалено!")
				msg.ReplyMarkup = mainMenu()
				botInstance.Send(msg)
			}
			continue
		}

		if update.Message == nil {
			continue
		}

		msg := update.Message.Text
		chatID := update.Message.Chat.ID
		firstName := update.Message.From.FirstName

		if msg == "/start" {
			sendWelcome(chatID, firstName)
			continue
		}

		if msg == "📋 Мої товари" {
			sendList(chatID)
			continue
		}

		if msg == "🗑 Видалити товар" {
			sendDeleteMenu(chatID)
			continue
		}

		if msg == "ℹ️ Як користуватись" {
			text := "📖 Інструкція:\n\n" +
				"1️⃣ Скопіюй посилання на товар з магазину\n" +
				"2️⃣ Відправ його мені\n" +
				"3️⃣ Я збережу товар і буду стежити за ціною\n" +
				"4️⃣ Як тільки ціна впаде — отримаєш сповіщення 🔔\n\n" +
				"⏱ Перевірка цін: кожну годину"
			m := tgbotapi.NewMessage(chatID, text)
			m.ReplyMarkup = mainMenu()
			botInstance.Send(m)
			continue
		}

		if isSupported(msg) {
			m := tgbotapi.NewMessage(chatID, "⏳ Отримую ціну...")
			botInstance.Send(m)

			price, err := fetchPrice(msg)
			if err != nil {
				botInstance.Send(tgbotapi.NewMessage(chatID, "❌ Помилка при отриманні ціни"))
				continue
			}

			saveProduct(chatID, msg, price)

			val, err := strconv.Atoi(cleanPrice(price))
			priceStr := price
			if err == nil {
				priceStr = formatPrice(val) + " ₴"
			}

			resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Збережено!\n💵 Поточна ціна: %s", priceStr))
			resp.ReplyMarkup = mainMenu()
			botInstance.Send(resp)
			continue
		}

		m := tgbotapi.NewMessage(chatID, "🔗 Підтримувані магазини:\n• rozetka.com.ua\n• epicentrk.ua\n• comfy.ua\n• prom.ua")
		m.ReplyMarkup = mainMenu()
		botInstance.Send(m)
	}
}
func isSupported(url string) bool {
	supported := []string{
		"rozetka.com.ua",
		"epicentrk.ua",
		"comfy.ua",
		"prom.ua",
	}
	for _, s := range supported {
		if strings.Contains(url, s) {
			return true
		}
	}
	return false
}
