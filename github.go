package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func copyToGithub(pathDropbox string, logPath string) error {
	githubUser := os.Getenv("GITHUB_USER")
	githubToken := os.Getenv("GITHUB_TOKEN")
	repoName := os.Getenv("GITHUB_REPO")

	if githubUser == "" || githubToken == "" || repoName == "" {
		return fmt.Errorf("faltan variables de entorno GITHUB_USER, GITHUB_TOKEN o GITHUB_REPO")
	}

	repoPath := filepath.Join(pathDropbox, repoName)

	// Crear repo en GitHub si no existe
	err := createGithubRepo(githubUser, githubToken, repoName)
	if err != nil {
		return fmt.Errorf("error creando repo en GitHub: %w", err)
	}

	// Si no existe localmente, clonar
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		remoteURL := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", githubUser, githubToken, githubUser, repoName)

		err = runGitCommand(pathDropbox, logPath, "clone", remoteURL, repoName)
		if err != nil {
			return fmt.Errorf("error al clonar repositorio: %w", err)
		}

		// limpiar URL del remote
		cleanURL := fmt.Sprintf("https://github.com/%s/%s.git", githubUser, repoName)
		_ = runGitCommand(repoPath, logPath, "remote", "set-url", "origin", cleanURL)
	}

	// Configurar usuario git en el repo local
	_ = runGitCommand(repoPath, logPath, "config", "user.name", githubUser)
	_ = runGitCommand(repoPath, logPath, "config", "user.email", fmt.Sprintf("%s@users.noreply.github.com", githubUser))

	// Copiar zips al repo
	files, err := os.ReadDir(pathDropbox)
	if err != nil {
		return fmt.Errorf("no se pudo leer la carpeta Dropbox: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(file.Name())) != ".zip" {
			continue
		}

		src := filepath.Join(pathDropbox, file.Name())
		dst := filepath.Join(repoPath, file.Name())

		err := copyFile(src, dst)
		if err != nil {
			logsManager(logPath, "ERROR", fmt.Sprintf("Error copiando %s al repo: %v", file.Name(), err))
			continue
		}
	}

	err = runGitCommand(repoPath, logPath, "add", ".")
	if err != nil {
		return fmt.Errorf("error en git add: %w", err)
	}


	statusOutput, err := runGitCommandOutput(repoPath, logPath, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("error en git status: %w", err)
	}

	if strings.TrimSpace(statusOutput) == "" {
		fmt.Println("ℹ️ No hay cambios para subir a GitHub.")
		logsManager(logPath, "INFO", "No hay cambios para subir a GitHub")
		return nil
	}

	err = runGitCommand(repoPath, logPath, "commit", "-m", "Backup actualizado")
	if err != nil {
		return fmt.Errorf("error en git commit: %w", err)
	}


	remoteURL := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", githubUser, githubToken, githubUser, repoName)
	err = runGitCommand(repoPath, logPath, "remote", "set-url", "origin", remoteURL)
	if err != nil {
		return fmt.Errorf("error configurando remote: %w", err)
	}

	branch, err := getCurrentBranch(repoPath, logPath)
	if err != nil || strings.TrimSpace(branch) == "" {
		branch = "main"
	}

	err = runGitCommand(repoPath, logPath, "push", "-u", "origin", strings.TrimSpace(branch))
	if err != nil {
		return fmt.Errorf("error en git push: %w", err)
	}


	cleanURL := fmt.Sprintf("https://github.com/%s/%s.git", githubUser, repoName)
	_ = runGitCommand(repoPath, logPath, "remote", "set-url", "origin", cleanURL)

	logsManager(logPath, "INFO", "Archivos subidos correctamente a GitHub con git push")
	return nil
}

func createGithubRepo(_, token, repo string) error {
	url := "https://api.github.com/user/repos"

	body := map[string]interface{}{
		"name":    repo,
		"private": false,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		fmt.Println("✅ Repo creado en GitHub")
		return nil
	}

	if resp.StatusCode == 422 {
		fmt.Println("ℹ️ Repo ya existe en GitHub")
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("github api (%d): %s", resp.StatusCode, string(bodyBytes))
}

func runGitCommand(workDir, logPath string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		logsManager(logPath, "ERROR", fmt.Sprintf("git %s -> %s", strings.Join(args, " "), string(output)))
		return fmt.Errorf("%s", string(output))
	}

	logsManager(logPath, "INFO", fmt.Sprintf("git %s", strings.Join(args, " ")))
	return nil
}

func runGitCommandOutput(workDir, logPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		logsManager(logPath, "ERROR", fmt.Sprintf("git %s -> %s", strings.Join(args, " "), string(output)))
		return "", fmt.Errorf("%s", string(output))
	}

	return string(output), nil
}

func getCurrentBranch(workDir, logPath string) (string, error) {
	output, err := runGitCommandOutput(workDir, logPath, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
