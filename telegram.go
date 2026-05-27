package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type APIResponse struct {
	OK          bool     `json:"ok"`
	Result      []Update `json:"result"`
	Description string   `json:"description"`
}

type APIOkResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

type APIFileResponse struct {
	OK          bool         `json:"ok"`
	Result      TelegramFile `json:"result"`
	Description string       `json:"description"`
}

type TelegramFile struct {
	FilePath string `json:"file_path"`
}

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	Chat     Chat      `json:"chat"`
	Text     string    `json:"text"`
	Document *Document `json:"document"`
}

type Document struct {
	FileID string `json:"file_id"`
}

type Chat struct {
	ID int64 `json:"id"`
}

var tgClient = &http.Client{
	Timeout: 35 * time.Second,
}

const tgAPIBase = "https://api.telegram.org/bot"

func StartPolling(token string) {
	if err := SetCommands(token); err != nil {
		log.Printf("ошибка установки команд: %v", err)
	}

	log.Println("Бот запускается")

	offset := 0

	for {
		updates, err := GetUpdates(token, offset, 5, 30)

		if err != nil {
			log.Printf("Ошибка получения обновлений: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, u := range updates {
			if u.Message.Document != nil {
				go HandleDocument(token, u.Message.Chat.ID, u.Message.Document.FileID)
			} else if u.Message != nil && u.Message.Text != "" {
				go HandleMessage(token, u.Message.Chat.ID, u.Message.Text)
			}

			offset = u.UpdateID + 1
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func GetUpdates(token string, offset, limit, timeout int) ([]Update, error) {
	url := fmt.Sprintf("%s%s/getUpdates?offset=%d&limit=%d&timeout=%d", tgAPIBase, token, offset, limit, timeout)

	resp, err := tgClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка сети: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("API ошибка: %s", apiResp.Description)
	}

	return apiResp.Result, nil
}

func SendMessage(token string, chatID int64, text string) error {
	url := fmt.Sprintf("%s%s/sendMessage", tgAPIBase, token)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка конвертации в json: %w", err)
	}

	resp, err := tgClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка отправки: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIOkResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram error: %s", apiResp.Description)
	}

	return nil
}

func SendDocument(token string, chatID int64, filePath, caption string) error {
	url := fmt.Sprintf("%s%s/sendDocument", tgAPIBase, token)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("chat_id", fmt.Sprintf("%d", chatID))

	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("ошибка создания form-file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("ошибка копирования файла: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия файла: %w", err)
	}

	resp, err := tgClient.Post(url, writer.FormDataContentType(), body)
	if err != nil {
		return fmt.Errorf("ошибка HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIOkResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram error: %s", apiResp.Description)
	}

	return nil
}

func SetCommands(token string) error {
	url := fmt.Sprintf("%s%s/setMyCommands", tgAPIBase, token)
	payload := map[string]interface{}{
		"commands": []map[string]string{
			{"command": "start", "description": "Запустить бота"},
			{"command": "track", "description": "Скачать один трек"},
			{"command": "list", "description": "Скачать список треков"},
			{"command": "file", "description": "Скачать треки из файла"},
			{"command": "cancel", "description": "Отменить текущее действие"},
			{"command": "help", "description": "Посмотреть примеры комманд"},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка конвертации в json: %w", err)
	}

	resp, err := tgClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка запроса: %w", err)
	}

	defer resp.Body.Close()

	var apiResp APIOkResponse

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram error: %s", apiResp.Description)
	}

	return nil
}

func DownloadTelegramFile(token string, fileID string) (string, error) {
	url := fmt.Sprintf("%s%s/getFile?file_id=%s", tgAPIBase, token, fileID)

	resp, err := tgClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса пути файла: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("ошибка парсинга в JSON при скачивании файла: %w", err)
	}

	if !apiResp.OK {
		return "", fmt.Errorf("telegram error: %s", apiResp.Description)
	}

	url = fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, apiResp.Result.FilePath)

	resp2, err := tgClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса файла: %w", err)
	}
	defer resp2.Body.Close()

	tmpFile, err := os.CreateTemp("", "tgfile-*.txt")
	if err != nil {
		return "", fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp2.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка копирования файла: %w", err)
	}

	finalPath := tmpFile.Name()

	return finalPath, nil
}
