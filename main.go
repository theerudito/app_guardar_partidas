package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	userDirectory, err := os.UserHomeDir()

	if err != nil {
		fmt.Println("Error al obtener el directorio de usuario:", err)
		return
	}

	folderGames := []string{
		filepath.Join(userDirectory, "Saved Games"),
		filepath.Join(userDirectory, "Downloads", "God Of War", "PS3", "dev_hdd0", "home"),
		filepath.Join(userDirectory, "AppData", "LocalLow", "Team Cherry"),
	}

	dropboxPath := filepath.Join(userDirectory, "Dropbox", "Partidas Juegos")
	logPath := filepath.Join(dropboxPath, "backup_log.txt")

	//======================= DROPBOX =================================
	var excludeFolder = map[string]bool{"trophy": true}
	err = copyToDropbox(userDirectory, dropboxPath, logPath, folderGames, excludeFolder)
	if err != nil {
		fmt.Println("error copiar a carpeta de dropbox:", err)
	} else {
		fmt.Println("✅ Archivos copiado correctamente a Dropbox.")
	}
	//======================= DROPBOX =================================

	//======================= GITHUB ==================================
	err = copyToGithub(dropboxPath, logPath)
	if err != nil {
		fmt.Println("Error subir a GitHub:", err)
	} else {
		fmt.Println("✅ Archivos subidos correctamente a GitHub.")
	}
	//======================= GITHUB ==================================

}
