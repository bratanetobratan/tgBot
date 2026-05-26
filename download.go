package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

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
