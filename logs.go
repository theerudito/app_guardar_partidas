package main

import (
	"fmt"
	"os"
	"time"
)

func logsManager(path string, logType string, message string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("❌ No se pudo abrir el log: %v\n", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, logType, message)
	f.WriteString(entry)
	fmt.Printf("📝 [%s] %s\n", logType, message)
}
