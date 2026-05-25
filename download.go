package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func DownloadAudio(tmpDir, url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	outPath := filepath.Join(tmpDir, "audio.%(ext)s")

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"-o", outPath,
		"--no-playlist",
		"--no-warnings",
		url,
	)

	if _, err := cmd.Output(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("скачивание заняло слишком много времени")
		}
		return "", fmt.Errorf("ошибка скачивания: %w", err)
	}

	finalPath := "~/Downloads/audio.mp3"
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		return "", fmt.Errorf("файл audio.mp3 не был создан после скачивания")
	}

	return finalPath, nil
}
