package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Adelphopoet/dnd-bot-app/telegram"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

type Secrets struct {
	DatabaseHost     string `yaml:"database_host"`
	DatabaseLogin    string `yaml:"database_login"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseName     string `yaml:"database_name"`
	TelegramToken    string `yaml:"telegram_token"`
}

func main() {
	// Чтение содержимого YAML файла
	content, err := ioutil.ReadFile("secrets.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Распаковка содержимого YAML файла в структуру Secrets
	var secrets Secrets
	err = yaml.Unmarshal(content, &secrets)
	if err != nil {
		log.Fatal(err)
	}

	// Подключение к базе данных PostgreSQL
	connectionString := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		secrets.DatabaseHost, secrets.DatabaseLogin, secrets.DatabasePassword, secrets.DatabaseName)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверка соединения с базой данных
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := telegram.NewBot(secrets.TelegramToken, db)
	if err != nil {
		log.Fatal("failed to create bot:", err)
	}

	err = bot.Start()
	if err != nil {
		log.Fatal("failed to start bot:", err)
	}
}
