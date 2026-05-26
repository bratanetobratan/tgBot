package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ID int64 `json:"id"`
}

var tgClient = &http.Client{
	Timeout: 35 * time.Second,
}

const tgAPIBase = "https://api.telegram.org/bot"

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
