package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"` // Added to control streaming
	System string `json:"system"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: query <natural language request>")
		os.Exit(1)
	}

	query := strings.Join(os.Args[1:], " ")
	kubectlCommand := generateKubectlCommand(query)

	fmt.Printf("--> %s\n", kubectlCommand)
}

func generateKubectlCommand(query string) string {
	ollamaUrl := "http://localhost:11434/api/generate"

	// Create the request payload
	request := OllamaRequest{
		Model:  "mistral", // Adjust model name as needed
		Prompt: "Generate only a kubectl command (starting with 'kubectl ') for this request and nothing else. No explanation, no description, only command: " + query,
		Stream: false, // Disable streaming for simpler response handling
		System: "You are an expert kubectl command generator, that only generates valid kubectl commands. You should never provide any explanations. You should always output raw shell commands as text with ```. You can use this documentation as reference https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands",
	}

	body, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshaling request: %s\n", err)
		return ""
	}

	r, err := http.NewRequest("POST", ollamaUrl, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return ""
	}

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return ""
	}
	defer res.Body.Close()

	// Read the full response
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return ""
	}

	// Print raw response for debugging
	// fmt.Printf("Raw Ollama response: %s\n", string(bodyBytes))

	// Parse the response
	var response OllamaResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		fmt.Printf("Error parsing Ollama response: %s\n", err)
		return ""
	}

	// Clean and extract the command
	command := strings.TrimSpace(response.Response)

	// Remove surrounding triple backticks if present
	command = strings.TrimPrefix(command, "```")
	command = strings.TrimSuffix(command, "```")
	command = strings.TrimSpace(command) // Ensure no leading/trailing spaces

	// Ensure response contains a valid kubectl command
	// Validate that it's a kubectl command
	if !strings.HasPrefix(command, "kubectl ") {
		fmt.Println("Invalid command received:", command)
		return ""
	}

	return command
}
