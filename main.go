package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
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

// Add new type for Ollama error response
type OllamaErrorResponse struct {
	Error string `json:"error"`
}

// Create a channel to control the thinking animation
var stopAnimation chan bool

func main() {
	var model string
	var ollamaUrl string

	// Create the root command using cobra
	var rootCmd = &cobra.Command{
		Use:   "kubectllama",
		Short: "kubectllama is a CLI tool to generate kubectl commands using AI",
		Long:  `kubectllama generates kubectl commands based on natural language input, using AI models to understand your requests.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Ensure there is a query argument
			if len(args) < 1 {
				fmt.Println("Usage: kubectllama <natural language request>")
				os.Exit(1)
			}

			// Get the query (the rest of the args)
			query := strings.Join(args, " ")

			// Initialize the stop channel
			// This channel sends a true when the Ollama API responds.
			// When the goroutine receives the "true" it stops the goroutine
			// aka stops the animation
			stopAnimation = make(chan bool)

			// Start the thinking animation.
			// This is a goroutine
			go showThinkingAnimation()

			kubectlCommand := generateKubectlCommand(query, model, ollamaUrl)

			// Stop the animation and clear the line
			stopAnimation <- true
			fmt.Printf("\r%s\r", strings.Repeat(" ", 50)) // Clear the line

			fmt.Printf("--> %s\n", kubectlCommand)
		},
	}

	// Add the --model flag to the root command
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "mistral", "Specify the AI model to use (default: mistral)")
	rootCmd.PersistentFlags().StringVarP(&ollamaUrl, "url", "u", "http://localhost:11434", "Specify the Ollama API URL (default: http://localhost:11434)")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateKubectlCommand(query string, model string, ollamaUrl string) string {
	// Ensure the URL ends with the correct endpoint
	if !strings.HasSuffix(ollamaUrl, "/api/generate") {
		ollamaUrl = strings.TrimSuffix(ollamaUrl, "/") + "/api/generate"
	}
	// Create the request payload
	request := OllamaRequest{
		Model:  model,
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

	// First try to unmarshal as an error response
	var errorResponse OllamaErrorResponse
	if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error != "" {
		// Check if the error indicates the model is not found
		if strings.Contains(strings.ToLower(errorResponse.Error), "model") &&
			strings.Contains(strings.ToLower(errorResponse.Error), "not found") {
			fmt.Printf("\rError: Model '%s' is not installed locally. Please download it first using:\n", model)
			fmt.Printf("ollama pull %s\n", model)
			os.Exit(1)
		}
		// Handle other types of errors
		fmt.Printf("\rError from Ollama: %s\n", errorResponse.Error)
		return ""
	}

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

// goroutine to show animation while in background API is sending the response
func showThinkingAnimation() {
	thinkingText := "Thinking of the kubectl command"
	dots := []string{".", "..", "..."}
	for {
		select {
		// if channel stopAnimation sends true, stop the animation and return
		case <-stopAnimation:
			return
		// else keep running the goroutine
		default:
			// loop through the dots and don't stop until you receive (<- stopAnimation)
			// a true from the channel
			for _, dot := range dots {
				select {
				case <-stopAnimation:
					return
				default:
					fmt.Printf("\r%s%s", thinkingText, dot)
					time.Sleep(500 * time.Millisecond)
				}
			}
		}
	}
}
