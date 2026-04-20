package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type GameBackup struct {
	Path       string
	CustomName string
}

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

	folderGames := []GameBackup{
		{
			Path:       filepath.Join(userDirectory, "Saved Games"),
			CustomName: "",
		},
		{
			Path:       filepath.Join(userDirectory, "Downloads", "God Of War", "PS3", "dev_hdd0", "home"),
			CustomName: "",
		},
		{
			Path:       filepath.Join(userDirectory, "AppData", "LocalLow", "Team Cherry"),
			CustomName: "",
		},
		{
			Path:       `C:\Games\Assassins Creed Shadows\saves`,
			CustomName: "Assassin's Creed Shadows",
		},
		{
			Path:       `C:\Users\Usuario\Documents\Criterion Games\Need For Speed(TM) Most Wanted`,
			CustomName: "Need For Speed(TM) Most Wanted",
		},
		{
			Path:       `C:\Users\Usuario\Documents\NFS Most Wanted`,
			CustomName: "NFS Most Wanted",
		},
	}

	dropboxPath := filepath.Join(userDirectory, "Dropbox", "Partidas Juegos")
	logPath := filepath.Join(dropboxPath, "backup_log.txt")

	var excludeFolder = map[string]bool{
		"trophy": true,
	}

	err = copyToDropbox(dropboxPath, logPath, folderGames, excludeFolder)
	if err != nil {
		fmt.Println("error copiar a carpeta de dropbox:", err)
	} else {
		fmt.Println("✅ Archivos copiados correctamente a Dropbox.")
	}

	err = copyToGithub(dropboxPath, logPath)
	if err != nil {
		fmt.Println("Error subir a GitHub:", err)
	} else {
		fmt.Println("✅ Archivos subidos correctamente a GitHub.")
	}
}
