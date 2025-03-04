package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skyscale",
	Short: "Skyscale - Serverless Function Management",
	Long:  `Skyscale is a lightweight serverless function management platform powered by Firecracker`,
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(invokeCmd)
	rootCmd.AddCommand(logsCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [function_name]",
	Short: "Initialize a new function project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functionName := args[0]
		err := initializeFunction(functionName)
		if err != nil {
			fmt.Printf("❌ Error initializing function: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Function '%s' initialized successfully.\n", functionName)
	},
}

func initializeFunction(functionName string) error {
	// Define structure
	dirs := []string{
		functionName,
	}

	files := map[string]string{
		filepath.Join(functionName, "handler.py"): `def handler(event, context):
    """Skyscale function entry point"""
    return {"message": "Hello from ` + functionName + `!"}
`,
		filepath.Join(functionName, "requirements.txt"): `# Add your dependencies here`,
		filepath.Join(functionName, "skyscale.yaml"): `name: ` + functionName + `
runtime: python3.9
entrypoint: handler.handler`,
	}

	// Create directories
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create files
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

var deployCmd = &cobra.Command{
	Use:   "deploy [function_name]",
	Short: "Deploy a function to Skyscale",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functionName := args[0]
		err := deployFunction(functionName)
		if err != nil {
			fmt.Printf("❌ Error deploying function: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Function '%s' deployed successfully.\n", functionName)
	},
}

func deployFunction(functionName string) error {
	// Define the function directory
	functionDir := filepath.Join(functionName)
	// Read the handler.py file
	handlerPath := filepath.Join(functionDir, "handler.py")
	handlerCode, err := os.ReadFile(handlerPath)
	if err != nil {
		return fmt.Errorf("failed to read handler.py: %v", err)
	}

	// Read the requirements.txt file
	requirementsPath := filepath.Join(functionDir, "requirements.txt")
	requirements, err := os.ReadFile(requirementsPath)
	if err != nil {
		return fmt.Errorf("failed to read requirements.txt: %v", err)
	}

	// Read the skyscale.yaml file
	configPath := filepath.Join(functionDir, "skyscale.yaml")
	config, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read skyscale.yaml: %v", err)
	}

	// Prepare the function data
	data := map[string]interface{}{
		"name":         functionName,
		"code":         string(handlerCode),
		"requirements": string(requirements),
		"config":       string(config),
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Send POST request to the server
	resp, err := http.Post("http://localhost:8080/deploy", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deploy function, status: %s", resp.Status)
	}

	return nil
}

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Invoke a deployed function",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement function invocation
		fmt.Println("Invoking function...")
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Retrieve function logs",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement log retrieval
		fmt.Println("Retrieving function logs...")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
