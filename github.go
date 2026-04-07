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

	// La misma carpeta de Dropbox será el repo local
	repoPath := pathDropbox

	// Crear repo remoto si no existe
	if err := createGithubRepo(githubToken, repoName); err != nil {
		return fmt.Errorf("error creando repo en GitHub: %w", err)
	}

	// Inicializar git local si no existe
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		if err := runGitCommand(repoPath, logPath, "init"); err != nil {
			return fmt.Errorf("error en git init: %w", err)
		}
	}

	// Configuración de usuario
	_ = runGitCommand(repoPath, logPath, "config", "user.name", githubUser)
	_ = runGitCommand(repoPath, logPath, "config", "user.email", fmt.Sprintf("%s@users.noreply.github.com", githubUser))

	// .gitignore opcional
	if err := ensureGitIgnore(repoPath); err != nil {
		logsManager(logPath, "WARN", fmt.Sprintf("No se pudo crear .gitignore: %v", err))
	}

	remoteURLWithToken := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", githubUser, githubToken, githubUser, repoName)
	remoteURLClean := fmt.Sprintf("https://github.com/%s/%s.git", githubUser, repoName)

	// Crear o actualizar origin
	if err := setOrigin(repoPath, logPath, remoteURLWithToken); err != nil {
		return fmt.Errorf("error configurando remote origin: %w", err)
	}

	// Forzar rama main
	currentBranch, _ := getCurrentBranch(repoPath, logPath)
	if strings.TrimSpace(currentBranch) == "" {
		if err := runGitCommand(repoPath, logPath, "checkout", "--orphan", "main"); err != nil {
			return fmt.Errorf("error creando rama main: %w", err)
		}
	} else if strings.TrimSpace(currentBranch) != "main" {
		if err := runGitCommand(repoPath, logPath, "branch", "-M", "main"); err != nil {
			return fmt.Errorf("error renombrando rama a main: %w", err)
		}
	}

	// Agregar todo lo actual
	if err := runGitCommand(repoPath, logPath, "add", "."); err != nil {
		return fmt.Errorf("error en git add: %w", err)
	}

	statusOutput, err := runGitCommandOutput(repoPath, logPath, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("error en git status: %w", err)
	}

	if strings.TrimSpace(statusOutput) == "" {
		fmt.Println("ℹ️ No hay cambios para subir a GitHub.")
		logsManager(logPath, "INFO", "No hay cambios para subir a GitHub")
		_ = runGitCommand(repoPath, logPath, "remote", "set-url", "origin", remoteURLClean)
		return nil
	}

	// Commit
	if err := runGitCommand(repoPath, logPath, "commit", "-m", "Backup actualizado"); err != nil {
		return fmt.Errorf("error en git commit: %w", err)
	}

	// Subir forzando lo local sobre remoto
	if err := runGitCommand(repoPath, logPath, "push", "-u", "origin", "main", "--force"); err != nil {
		return fmt.Errorf("error en git push: %w", err)
	}

	// Limpiar token del remote
	_ = runGitCommand(repoPath, logPath, "remote", "set-url", "origin", remoteURLClean)

	logsManager(logPath, "INFO", "Archivos subidos correctamente a GitHub con git push --force")
	return nil
}

func createGithubRepo(token, repo string) error {
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

func setOrigin(repoPath, logPath, remoteURL string) error {
	err := runGitCommand(repoPath, logPath, "remote", "get-url", "origin")
	if err != nil {
		return runGitCommand(repoPath, logPath, "remote", "add", "origin", remoteURL)
	}
	return runGitCommand(repoPath, logPath, "remote", "set-url", "origin", remoteURL)
}

func ensureGitIgnore(repoPath string) error {
	gitIgnorePath := filepath.Join(repoPath, ".gitignore")

	if _, err := os.Stat(gitIgnorePath); err == nil {
		return nil
	}

	content := []byte(".env\nbackup_log.txt\nsaved_games/\n")
	return os.WriteFile(gitIgnorePath, content, 0644)
}

func runGitCommand(workDir, logPath string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		logsManager(logPath, "ERROR", fmt.Sprintf("git %s -> %s", strings.Join(args, " "), string(output)))
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
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
		return "", fmt.Errorf("%s", strings.TrimSpace(string(output)))
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
