package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func SearchYouTube(query string) (string, error) {
	query = strings.TrimSpace(query)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--dump-json",
		"--no-download",
		"--no-warnings",
		"ytsearch1:"+query,
	)

	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("поиск занял слишком много времени")
		}
		return "", fmt.Errorf("ошибка выполнения yt-dlp: %w", err)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return "", fmt.Errorf("не удалось распарсить ответ yt-dlp: %w", err)
	}

	if result.ID == "" {
		return "", fmt.Errorf("видео не найдено по запросу: %q", query)
	}

	return "https://www.youtube.com/watch?v=" + result.ID, nil
}

func DownloadAudio(tmpDir, url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	outPath := filepath.Join(tmpDir, "%(title)s.%(ext)s")

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"-o", outPath,
		"--no-playlist",
		"--no-warnings",
		"--match-filter", "duration < 600",
		"--print", "after_move:filepath",
		url,
	)

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("скачивание заняло слишком много времени")
		}
		return "", fmt.Errorf("ошибка скачивания: %w", err)
	}

	finalPath := strings.TrimSpace(string(output))
	if finalPath == "" {
		return "", fmt.Errorf("yt-dlp не вернул путь к файлу")
	}

	return finalPath, nil
}
