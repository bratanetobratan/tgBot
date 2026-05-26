package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Бот запускается")

	offset := 0

	for {
		updates, err := GetUpdates(GetBotToken(), offset, 5, 30)

		if err != nil {
			log.Printf("Ошибка получения обновлений: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, u := range updates {
			if u.Message != nil && u.Message.Text != "" {
				go HandleMessage(GetBotToken(), u.Message.Chat.ID, u.Message.Text)
			}

			offset = u.UpdateID + 1
		}

		time.Sleep(100 * time.Millisecond)
	}
}
