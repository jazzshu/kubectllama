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

	var version = "dev"

	// Create the root command using cobra
	var rootCmd = &cobra.Command{
		Use:     "kubectllama",
		Short:   "kubectllama is a CLI tool to generate kubectl commands using AI",
		Long:    `kubectllama generates kubectl commands based on natural language input, using AI models to understand your requests.`,
		Version: version,
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

			kubectlCommand, description := generateKubectlCommand(query, model, ollamaUrl)

			// Stop the animation and clear the line
			stopAnimation <- true
			fmt.Printf("\r%s\r", strings.Repeat(" ", 50)) // Clear the line

			fmt.Printf("\033[32m%s\033[0m\n", kubectlCommand) // Command in green
			fmt.Printf("%s\n", description)

			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Printf("Execute this command? [Y/n] ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(strings.ToLower(input))

				switch input {
				case "y", "yes", "":
					// Execute kubectl command
					execCmd := exec.Command("kubectl", strings.Fields(kubectlCommand)[1:]...)
					execCmd.Stdout = os.Stdout
					execCmd.Stderr = os.Stderr
					if err := execCmd.Run(); err != nil {
						fmt.Printf("Error executing command: %s\n", err)
					}
					return // Exit after execution
				case "n", "no":
					os.Exit(0)
				default:
					fmt.Println("Invalid input. Please enter 'Y' or 'n'.")
					continue
				}
			}
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

func generateKubectlCommand(query string, model string, ollamaUrl string) (command string, description string) {
	// Ensure the URL ends with the correct endpoint
	if !strings.HasSuffix(ollamaUrl, "/api/generate") {
		ollamaUrl = strings.TrimSuffix(ollamaUrl, "/") + "/api/generate"
	}
	// Create the request payload
	request := OllamaRequest{
		Model:  model,
		Prompt: "Generate a kubectl command (starting with 'kubectl ') for this request and then on a new line, give a one line explanation of the kubectl command: " + query,
		Stream: false, // Disable streaming for simpler response handling
		System: "You are an expert kubectl command generator, that generates only valid kubectl command. You should never provide any links or request for any additional prompts. You must use this documentation as reference https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands",
	}

	body, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshaling request: %s\n", err)
		return "", ""
	}

	r, err := http.NewRequest("POST", ollamaUrl, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return "", ""
	}

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return "", ""
	}
	defer res.Body.Close()

	// Read the full response
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return "", ""
	}

	// Use for debugging
	// fmt.Printf("Raw Ollama response: %s\n", string(bodyBytes))

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
		return "", ""
	}

	// Parse the response
	var response OllamaResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		fmt.Printf("Error parsing Ollama response: %s\n", err)
		return "", ""
	}

	// Clean and extract the command
	modelResponse := strings.TrimSpace(response.Response)

	lines := strings.SplitN(modelResponse, "\n\n", 2)
	command = strings.TrimSpace(strings.Trim(lines[0], "`")) //Remove backticks
	description = ""
	if len(lines) > 1 {
		description = strings.TrimSpace(lines[1])
	}

	return command, description
}

// goroutine to show animation while in background API is sending the response
func showThinkingAnimation() {
	spinner := []string{"|", "/", "-", "\\"}
	thinkingText := "Generating kubectl command "
	i := 0
	for {
		select {
		// if channel stopAnimation sends true, stop the animation and return
		case <-stopAnimation:
			fmt.Printf("\r%s... Done! 	\n", thinkingText)
			return
		// else keep running the goroutine
		default:
			// Print the spinner with carriage return so it updates in place.
			fmt.Printf("\r%s%s", thinkingText, spinner[i%len(spinner)])
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}
