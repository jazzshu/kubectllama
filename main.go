package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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
	// Place holder
	return "kubectl get pods -n default"
}

func executeKubectlCommand(cmdStr string) {
	parts := strings.Fields(cmdStr)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}
