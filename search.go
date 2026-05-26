package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func SearchYouTube(Uquery string) (string, error) {
	query := strings.TrimSpace(Uquery)

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
