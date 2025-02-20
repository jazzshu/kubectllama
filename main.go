package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func main() {
	fmt.Printf("Enter your request: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	query := scanner.Text()

	kubectlCommand := generateKubectlCommand(query)

	fmt.Printf("--> : %s\n", kubectlCommand)
	fmt.Printf("Press Enter to execute or Ctrl+C to cancel")
	scanner.Scan()

	executeKubectlCommand(kubectlCommand)
}

func generateKubectlCommand(query string) string {
	ollamaUrl := "http://localhost:11434/api/generate"

	body := []byte(`{
		"model": "mistral",
		"prompt": "Strictly return only a valid kubectl command as plain text. No explanations, no newlines, no formatting: ` + query + `",
		"stream": false,
		"format": "json"
	}`)

	r, err := http.NewRequest("POST", ollamaUrl, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Read raw response
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("Raw response from Ollama:", string(bodyBytes)) // Debugging

	var response struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		fmt.Printf("Error parsing Ollama response: %s\n", err)
		return ""
	}

	// Ensure response contains a valid kubectl command
	command := strings.TrimSpace(response.Response)
	if !strings.HasPrefix(command, "kubectl ") {
		fmt.Println("Invalid command received:", command)
		return ""
	}

	return command
}

func executeKubectlCommand(cmdStr string) {
	if cmdStr == "" {
		fmt.Println("Error: empty command")
		return
	}

	parts := strings.Fields(cmdStr)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %s\n", err)
	}
}
