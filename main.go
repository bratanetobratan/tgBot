package main

import "fmt"

func main() {
	var artist, title string
	fmt.Scan(&artist)
	fmt.Scan(&title)
	url, err := SearchYouTube(artist, title)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	path, _ := DownloadAudio("~/Downloads", url)
	fmt.Printf("Скачано в:", path)
}
