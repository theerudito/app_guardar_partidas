package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyToFolder(folderPath, logPath string, gamesPath []GamesList, excludeFolder map[string]bool) error {

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
			return fmt.Errorf("no se pudo crear el directorio destino: %w", err)
		}
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("no se pudo abrir el archivo de log: %w", err)
	}

	defer logFile.Close()

	for _, game := range gamesPath {
		if _, err := os.Stat(game.GamePath); os.IsNotExist(err) {
			msg := fmt.Sprintf("Directorio no encontrado: %s", game.GamePath)
			fmt.Println(msg)
			logsManager(logPath, "WARN", msg)
			continue
		}

		if err := zipGameFolder(game.GamePath, folderPath, logPath, excludeFolder, game.GameName); err != nil {
			msg := fmt.Sprintf("Error al comprimir %s (%s): %v", game.GameName, game.GamePath, err)
			fmt.Println(msg)
			logsManager(logPath, "ERROR", msg)
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
		return fmt.Errorf("no se pudo crear archivo zip: %w", err)
	}

	success := false

	defer func() {
		zipFile.Close()
		if !success {
			os.Remove(zipPath)
		}
	}()

	zipWriter := zip.NewWriter(zipFile)

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

		header, err := zip.FileInfoHeader(info)

		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)

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
		return fmt.Errorf("error al leer archivos para el zip: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("error al finalizar la estructura del zip: %w", err)
	}

	if err := zipFile.Sync(); err != nil {
		return fmt.Errorf("error al sincronizar zip con el disco: %w", err)
	}

	success = true

	logsManager(logPath, "INFO", fmt.Sprintf("Backup creado correctamente: %s", zipName))

	return nil

}
