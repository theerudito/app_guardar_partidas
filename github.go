package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type UploadBody struct {
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha,omitempty"`
}

type GitFileResponse struct {
	SHA string `json:"sha"`
}

func copyToGithub(pathDropbox string, logPath string) error {

	var user = os.Getenv("theerudito")
	var token = os.Getenv("token")
	var repo = "saved_games"

	files, err := os.ReadDir(pathDropbox)
	if err != nil {
		return fmt.Errorf("no se pudo leer la carpeta de Dropbox: %v", err)
	}

	client := &http.Client{}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".zip" {
			continue
		}

		filePath := filepath.Join(pathDropbox, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error al leer %s: %v\n", file.Name(), err)
			continue
		}

		encoded := base64.StdEncoding.EncodeToString(content)
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", user, repo, file.Name())

		var sha string
		reqCheck, _ := http.NewRequest("GET", url, nil)
		reqCheck.Header.Set("Authorization", "Bearer "+token)
		reqCheck.Header.Set("Accept", "application/vnd.github+json")

		respCheck, err := client.Do(reqCheck)
		if err == nil && respCheck.StatusCode == 200 {
			defer respCheck.Body.Close()
			var existing GitFileResponse
			json.NewDecoder(respCheck.Body).Decode(&existing)
			sha = existing.SHA
			fmt.Printf("🔄 Archivo existente detectado: %s (sha=%s)\n", file.Name(), sha)
		}

		body := UploadBody{
			Message: fmt.Sprintf("Backup actualizado"),
			Content: encoded,
			SHA:     sha,
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := client.Do(req)
		if err != nil {
			logsManager(logPath, "ERROR", fmt.Sprintf("Error al subir %s: %v", file.Name(), err))
			continue
		}

		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("✅ Subido: %s\n", file.Name())
			logsManager(logPath, "INFO", fmt.Sprintf("Subido correctamente a GitHub: %s", file.Name()))
		} else {
			fmt.Printf("❌ Error al subir %s: %s\n", file.Name(), string(bodyBytes))
			logsManager(logPath, "ERROR", fmt.Sprintf("Error al subir %s: %s", file.Name(), string(bodyBytes)))
		}
	}

	return nil
}
