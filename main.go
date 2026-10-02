package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	var (
		folderPath, logPath string
		folderGames         []GamesList
	)

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Aviso: no se encontró archivo .env (%v)", err)
	}

	userDirectory, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error al obtener el directorio de usuario:", err)
		return
	}

	dropboxDir := filepath.Join(userDirectory, "Dropbox")

	if info, err := os.Stat(dropboxDir); err == nil && info.IsDir() {
		folderPath = filepath.Join(dropboxDir, "Partidas Juegos")
	} else {
		docsDir := filepath.Join(userDirectory, "Documents")
		folderPath = filepath.Join(docsDir, "Partidas Juegos")
	}

	if err := os.MkdirAll(folderPath, 0755); err != nil {
		log.Fatalf("Error al crear carpeta destino: %v", err)
	}

	logPath = filepath.Join(folderPath, "backup_log.txt")

	jsonData, err := os.ReadFile("./data.json")
	if err != nil {
		log.Fatalf("Error al leer data.json: %v", err)
	}

	if err := json.Unmarshal(jsonData, &folderGames); err != nil {
		log.Fatalf("Error al deserializar data.json: %v", err)
	}

	for i := range folderGames {
		clean := filepath.Clean(folderGames[i].GamePath)
		if filepath.VolumeName(clean) == "" {
			folderGames[i].GamePath = filepath.Join(userDirectory, clean)
		} else {
			folderGames[i].GamePath = clean
		}
	}

	var excludeFolder = map[string]bool{
		"trophy": true,
	}

	err = copyToFolder(folderPath, logPath, folderGames, excludeFolder)
	if err != nil {
		fmt.Println("Error al copiar a la carpeta de respaldo:", err)
	} else {
		fmt.Println("✅ Archivos copiados correctamente.")
	}

	err = copyToGithub(folderPath, logPath)
	if err != nil {
		fmt.Println("Error al subir a GitHub:", err)
	} else {
		fmt.Println("✅ Archivos subidos correctamente a GitHub.")
	}
}
