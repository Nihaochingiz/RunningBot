package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lpernett/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}
	token := os.Getenv("TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	currentDate := time.Now().Format("2006-01-02")
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			processMessage(bot, update, currentDate)
		}
	}
}

func processMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, currentDate string) {
	message := update.Message.Text

	if isValidFormat(message) {
		km, minutes := parseResponse(message)
		fmt.Printf("Километраж: %s км, Время: %s\n", km, minutes)

		insert(currentDate, km, minutes)

		responseMessage := fmt.Sprintf("Введено: %s км - %s", km, minutes)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)
		_, err := bot.Send(msg)
		if err != nil {
			panic(err)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат. Пожалуйста, введите данные в формате 'км-мин'")
		_, err := bot.Send(msg)
		if err != nil {
			panic(err)
		}
	}
}


func isValidFormat(input string) bool {
	r := regexp.MustCompile(`^\d+км-\d+мин$`)
	return r.MatchString(input)
}

func parseResponse(input string) (string, string) {
	values := strings.SplitN(input, "км-", 2)
	return values[0], values[1]
}

func insert(currentDate, km, minutes string) {
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUsername, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	formattedMinutes := strings.ReplaceAll(minutes, "мин", " мин")

	result, err := db.Exec("INSERT INTO running_statistics (date, distance, time) VALUES (?, ?, ?)", currentDate, km+" км", formattedMinutes)
	if err != nil {
		panic(err)
	}

	lastInsertId, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	fmt.Printf("ID добавленного объекта: %d\n", lastInsertId)
	fmt.Printf("Количество затронутых строк: %d\n", rowsAffected)
}