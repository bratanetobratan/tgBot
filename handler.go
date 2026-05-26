package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func HandleMessage(token string, chatID int64, text string) {
	text = strings.TrimSpace(text)

	if text == "/start" {
		SendMessage(token, chatID, "Здарова пират!\nЯ помогу тебе слушать ту музыку, в которую ты влюбился изначально, а не запиканное дерьмо\nНапиши трек в формте: Исполнитель - Название")
		return
	}

	if len(text) < 5 {
		SendMessage(token, chatID, "Бро, напиши исполнителя и название трека")
		return
	}

	SendMessage(token, chatID, "Нормалды, я погнал искать - это займет некоторое время")

	go processDownload(token, chatID, text)
}

func processDownload(token string, chatID int64, query string) {
	tmpDir, err := os.MkdirTemp("/home/snow/Downloads", "ytmp3_*")
	if err != nil {
		log.Printf("ошибка создания директории: %v", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	log.Printf("[Chat: %d] Ищу: %s", chatID, query)
	videoURL, err := SearchYouTube(query)
	if err != nil {
		SendMessage(token, chatID, fmt.Sprintf("Не нашел трек: %v, анлаки(", err))
		return
	}

	log.Printf("[Chat: %d] Скачиваю: %s", chatID, videoURL)
	filePath, err := DownloadAudio(tmpDir, videoURL)
	if err != nil {
		SendMessage(token, chatID, fmt.Sprintf("Не могу скачать трек: %v, анлаки(", err))
		return
	}

	log.Printf("[Chat: %d] Отправляю файл", chatID)
	err = SendDocument(token, chatID, filePath, query)
	if err != nil {
		SendMessage(token, chatID, fmt.Sprintf("Ошибка отправки: %v", err))
		return
	}

	log.Printf("[Chat: %d] Успешно!", chatID)
}
