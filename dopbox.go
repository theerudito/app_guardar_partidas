package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyToDropbox(dropboxPath, logPath string, gamesPath []GameBackup, excludeFolder map[string]bool) error {
	if _, err := os.Stat(dropboxPath); os.IsNotExist(err) {
		err = os.MkdirAll(dropboxPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("no se pudo crear el directorio destino: %w", err)
		}
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo de log: %w", err)
	}
	defer logFile.Close()

	for _, game := range gamesPath {
		baseDir := game.Path

		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			fmt.Printf("Directorio no encontrado: %s\n", baseDir)
			logsManager(logPath, "WARN", fmt.Sprintf("Directorio no encontrado: %s", baseDir))
			continue
		}

		switch game.CustomName {
		case "Assassin's Creed Shadows", "Need For Speed(TM) Most Wanted", "NFS Most Wanted", "Assassins Creed Black Flag Resynced":
			entries, err := os.ReadDir(baseDir)
			if err != nil {
				msg := fmt.Sprintf("Error al leer %s: %v", baseDir, err)
				fmt.Println(msg)
				logsManager(logPath, "ERROR", msg)
				continue
			}

			foundSubfolder := false
			for _, entry := range entries {
				if entry.IsDir() {
					subDir := filepath.Join(baseDir, entry.Name())
					err = zipGameFolder(subDir, dropboxPath, logPath, excludeFolder, game.CustomName)
					if err != nil {
						msg := fmt.Sprintf("Error al comprimir %s: %v", subDir, err)
						fmt.Println(msg)
						logsManager(logPath, "ERROR", msg)
					}
					foundSubfolder = true
					break
				}
			}

			if !foundSubfolder {
				msg := fmt.Sprintf("No se encontró subcarpeta dentro de %s", baseDir)
				fmt.Println(msg)
				logsManager(logPath, "WARN", msg)
			}
			continue
		}

		entries, err := os.ReadDir(baseDir)
		if err != nil {
			fmt.Printf("Error al leer %s: %v\n", baseDir, err)
			logsManager(logPath, "ERROR", fmt.Sprintf("Error al leer %s: %v", baseDir, err))
			continue
		}

		hasSubfolders := false
		for _, entry := range entries {
			if entry.IsDir() {
				hasSubfolders = true
				subDir := filepath.Join(baseDir, entry.Name())

				err = zipGameFolder(subDir, dropboxPath, logPath, excludeFolder, "")
				if err != nil {
					fmt.Printf("Error al comprimir %s: %v\n", subDir, err)
					logsManager(logPath, "ERROR", fmt.Sprintf("Error al comprimir %s: %v", subDir, err))
				}
			}
		}

		if !hasSubfolders {
			err = zipGameFolder(baseDir, dropboxPath, logPath, excludeFolder, game.CustomName)
			if err != nil {
				fmt.Printf("Error al comprimir %s: %v\n", baseDir, err)
				logsManager(logPath, "ERROR", fmt.Sprintf("Error al comprimir %s: %v", baseDir, err))
			}
		}
	}

	return nil
}

func zipGameFolder(folderPath, dropboxPath, logPath string, excludeFolder map[string]bool, customName string) error {
	gameName := customName
	if gameName == "" {
		gameName = filepath.Base(folderPath)
	}

	zipName := fmt.Sprintf("%s.zip", gameName)
	zipPath := filepath.Join(dropboxPath, zipName)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := zipFile.Close(); err != nil {
			fmt.Printf("Error al cerrar %s: %v\n", zipPath, err)
			logsManager(logPath, "ERROR", fmt.Sprintf("Error al cerrar %s: %v", zipPath, err))
		}
	}()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			fmt.Printf("Error al cerrar zip %s: %v\n", zipPath, err)
			logsManager(logPath, "ERROR", fmt.Sprintf("Error al cerrar zip %s: %v", zipPath, err))
		}
	}()

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if excludeFolder[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		return err
	}

	logsManager(logPath, "INFO", fmt.Sprintf("Backup creado: %s", zipName))
	return nil
}
