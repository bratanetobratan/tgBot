package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var userState = map[int64]string{}

const helpMessage = `🎵 Привет! Вот что я умею:

/track Исполнитель - Название
Скачать один трек
Пример: /track ACDC - Highway to Hell

/list
Скачать несколько треков сразу
После команды отправь список, каждый трек с новой строки:
ACDC - Highway to Hell
Nirvana - Smells Like Teen Spirit
Jay-Z - 99 Problems

/file
Скачать треки из текстового файла
После команды отправь .txt файл, каждый трек с новой строки

/cancel
Отменить текущее действие`

func HandleMessage(token string, chatID int64, text string) {
	text = strings.TrimSpace(text)

	switch {
	case text == "/start":
		SendMessage(token, chatID, "Здарова пират!\nЯ помогу тебе слушать ту музыку, в которую ты влюбился изначально, а не запиканное дерьмо\nНапиши трек в формте: Исполнитель - Название")
		return
	case text == "/cancel":
		delete(userState, chatID)
		SendMessage(token, chatID, "Отменил, жду команду")
		return
	case userState[chatID] == "waiting_list":
		urls := parseQueries(text)
		if len(urls) == 0 {
			SendMessage(token, chatID, "Список пустой, кидай каждый трек с новой строки")
			return
		}
		delete(userState, chatID)
		SendMessage(token, chatID, "Погнал скачивать треки")
		go processDownloadBatch(token, chatID, urls)
	case userState[chatID] == "waiting_file":
		SendMessage(token, chatID, "Бро, пришли файл с исполнителями и названиями треков")
	case strings.HasPrefix(text, "/track"):
		query := strings.TrimPrefix(text, "/track")
		if len(query) == 0 {
			SendMessage(token, chatID, "Бро, после команды должен быть трек и его исполнитель")
			return
		}
		SendMessage(token, chatID, "Нормалды, я погнал искать - это займет некоторое время")
		go processDownload(token, chatID, query)
		return
	case text == "/list":
		userState[chatID] = "waiting_list"
		return
	case text == "/file":
		userState[chatID] = "waiting_file"
		return
	case text == "/help":
		SendMessage(token, chatID, helpMessage)
	default:
		SendMessage(token, chatID, "Не понял, используй /track, /list или /file")
	}
}

func HandleDocument(token string, chatID int64, fileID string) error {
	if userState[chatID] != "waiting_file" {
		SendMessage(token, chatID, "Я не жду файл, используй /file сначала")
		return nil
	}

	filePath, err := DownloadTelegramFile(token, fileID)
	if err != nil {
		SendMessage(token, chatID, "Не могу скачать файл, попробуй ещё раз")
		return nil
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		SendMessage(token, chatID, "Не могу скачать файл, попробуй ещё раз")
		return nil
	}

	os.Remove(filePath)

	fileText := string(fileData)
	listOfTracks := parseQueries(fileText)
	if len(listOfTracks) == 0 {
		SendMessage(token, chatID, "Не могу скачать файл, он пустой")
		return nil
	}

	delete(userState, chatID)

	go processDownloadBatch(token, chatID, listOfTracks)

	return nil
}

func processDownload(token string, chatID int64, query string) {
	tmpDir, err := os.MkdirTemp("", "ytmp3_*")
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

	fileName := strings.TrimSuffix(filepath.Base(filePath), ".mp3")
	err = SendDocument(token, chatID, filePath, fileName)
	if err != nil {
		SendMessage(token, chatID, fmt.Sprintf("Ошибка отправки: %v", err))
		return
	}

	log.Printf("[Chat: %d] Успешно!", chatID)
}

func processDownloadBatch(token string, chatID int64, listOfTracks []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, track := range listOfTracks {
		wg.Add(1)
		go func(track string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			processDownload(token, chatID, track)
		}(track)
	}

	wg.Wait()
	SendMessage(token, chatID, "Все треки отправлены родничок!")
}

func parseQueries(text string) []string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 5 {
			continue
		}
		result = append(result, line)
	}
	return result
}
