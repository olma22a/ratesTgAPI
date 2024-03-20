package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type ExchangeRate struct {
	Koken string             `json:"static_koken"`
	Time1 time.Time          `json:"time"`
	Base  string             `json:"base_code"`
	Rates map[string]float64 `json:"conversion_rates"`
}

var correctPass bool = false

func main() {
	bot, err := tgbotapi.NewBotAPI("yourTGBOTTOKEN")
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()
	collection := client.Database("exchange").Collection("rub")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if correctPass {
			if isValidDateFormat(update.Message.Text) {

				messageTime, err := time.Parse("2006-01-02T15:04:05.000Z07:00", update.Message.Text)
				if err != nil {
					log.Fatal(err)
				}
				messageTime = messageTime.UTC()
				filter := bson.M{
					"time1": messageTime,
				}
				var result ExchangeRate
				if err := collection.FindOne(context.Background(), filter).Decode(&result); err != nil {
					log.Fatal(err)
				}

				responseText := fmt.Sprintf("курсы валют RUB:\n")
				for currency, rates := range result.Rates {
					responseText += fmt.Sprintf("%s: %f\n", currency, rates)
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "введите дату в формате:\n 2024-03-19T22:36:27.628+00:00")
				bot.Send(msg)
			}
		} else {
			var result ExchangeRate
			filter := bson.M{"koken": update.Message.Text}
			err := collection.FindOne(context.Background(), filter).Decode(&result)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "введите действующий токен")
				bot.Send(msg)
			} else {
				correctPass = true
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "токен ахуени\nвведите дату в формате: 2024-03-19T22:36:27.628+00:00\nдля получения котировок по всем валютам относительно рубля")
				bot.Send(msg)
			}
		}
	}
}

func isValidDateFormat(date string) bool {
	layout := "2006-01-02T15:04:05.000Z07:00"
	_, err := time.Parse(layout, date)
	return err == nil
}
