package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags
	cfgFile string
	// Base URL for the API
	baseURL string
	// API key for authentication
	apiKey string
)

var rootCmd = &cobra.Command{
	Use:   "skyscale",
	Short: "Skyscale - Serverless Function Management",
	Long:  `Skyscale is a lightweight serverless function management platform powered by Firecracker`,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add persistent flags for all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.skyscale.yaml)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "api-url", "http://localhost:8080", "API URL for the Skyscale control plane")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication")

	// Bind flags to viper config
	viper.BindPFlag("api_url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))

	// Add subcommands here
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(invokeCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(generateAPIKeyCmd)
	rootCmd.AddCommand(configCmd)

	// Add flags for generate-api-key command
	generateAPIKeyCmd.Flags().String("user-id", "cli-user", "User ID for the API key")
	generateAPIKeyCmd.Flags().StringSlice("roles", []string{"user"}, "Roles for the API key")
	generateAPIKeyCmd.Flags().Int64("expires-in", 86400, "Expiration time in seconds (default: 24 hours)")

	invokeCmd.Flags().String("input", "", "JSON input for the function")
	invokeCmd.Flags().String("input-file", "", "Path to a JSON file containing input for the function")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".skyscale" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigName(".skyscale")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		// Get values from config
		if viper.IsSet("api_url") {
			baseURL = viper.GetString("api_url")
		}
		if viper.IsSet("api_key") {
			apiKey = viper.GetString("api_key")
		}
	}
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Skyscale configuration",
	Long:  `Configure API URL and authentication for Skyscale CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Save configuration
		home, err := homedir.Dir()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		configPath := filepath.Join(home, ".skyscale.yaml")

		// Set values
		viper.Set("api_url", baseURL)
		viper.Set("api_key", apiKey)

		// Write config file
		err = viper.WriteConfigAs(configPath)
		if err != nil {
			fmt.Printf("❌ Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Configuration saved to %s\n", configPath)
		fmt.Printf("API URL: %s\n", baseURL)
		if apiKey != "" {
			fmt.Println("API Key: [CONFIGURED]")
		} else {
			fmt.Println("API Key: [NOT CONFIGURED]")
		}
	},
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

// makeAuthenticatedRequest makes an HTTP request with authentication headers
func makeAuthenticatedRequest(method, url string, body []byte) (*http.Response, error) {
	// Create a new request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add authentication if API key is provided
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Make the request
	client := &http.Client{}
	return client.Do(req)
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
	data := map[string]any{
		"name":         functionName,
		"runtime":      "python3.9", // Default runtime, could be extracted from config
		"code":         string(handlerCode),
		"requirements": string(requirements),
		"config":       string(config),
		"memory":       256, // Default values
		"timeout":      30,  // Default values
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Send POST request to the server using the correct API endpoint with authentication
	resp, err := makeAuthenticatedRequest("POST", baseURL+"/api/functions", jsonData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResponse map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			if errMsg, ok := errResponse["error"].(string); ok {
				return fmt.Errorf("failed to deploy function: %s", errMsg)
			}
		}
		return fmt.Errorf("failed to deploy function, status: %s", resp.Status)
	}

	return nil
}

// InvokeRequest represents a request to invoke a function
type InvokeRequest struct {
	Input   map[string]interface{} `json:"input"`
	Context map[string]interface{} `json:"context,omitempty"`
	Sync    bool                   `json:"sync"`
}

var invokeCmd = &cobra.Command{
	Use:   "invoke [function_name]",
	Short: "Invoke a deployed function",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functionName := args[0]

		// Get input from flag or file
		inputJSON, _ := cmd.Flags().GetString("input")
		inputFile, _ := cmd.Flags().GetString("input-file")

		// Parse input data
		input := map[string]any{}

		if inputFile != "" {
			// Read from file
			data, err := os.ReadFile(inputFile)
			if err != nil {
				fmt.Printf("❌ Error reading input file: %v\n", err)
				os.Exit(1)
			}

			if err := json.Unmarshal(data, &input); err != nil {
				fmt.Printf("❌ Error parsing input JSON from file: %v\n", err)
				os.Exit(1)
			}
		} else if inputJSON != "" {
			// Parse JSON string
			if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
				fmt.Printf("❌ Error parsing input JSON: %v\n", err)
				os.Exit(1)
			}
		}

		err := invokeFunction(functionName, input)
		if err != nil {
			fmt.Printf("❌ Error invoking function: %v\n", err)
			os.Exit(1)
		}
	},
}

func invokeFunction(functionName string, input map[string]any) error {
	// Prepare the invoke data with proper context
	context := map[string]any{
		"function_name": functionName,
		"invoked_at":    time.Now().Format(time.RFC3339),
		"client":        "skyscale-cli",
	}

	req := InvokeRequest{
		Input:   input,   // Use event instead of input
		Context: context, // Add proper context
		Sync:    true,    // Synchronous invocation
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Send POST request to the invoke endpoint with authentication
	resp, err := makeAuthenticatedRequest(
		"POST",
		baseURL+"/api/functions/name/"+functionName+"/invoke",
		jsonData,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResponse map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			if errMsg, ok := errResponse["error"].(string); ok {
				return fmt.Errorf("failed to invoke function: %s", errMsg)
			}
		}
		return fmt.Errorf("failed to invoke function, status: %s", resp.Status)
	}

	// Parse and print the response
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// Pretty print the result
	fmt.Println("Function Result:")
	outputJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format result: %v", err)
	}
	fmt.Println(string(outputJSON))

	return nil
}

var logsCmd = &cobra.Command{
	Use:   "logs [function_name]",
	Short: "Retrieve function logs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functionName := args[0]
		err := getLogs(functionName)
		if err != nil {
			fmt.Printf("❌ Error retrieving logs: %v\n", err)
			os.Exit(1)
		}
	},
}

func getLogs(functionName string) error {
	// First, get the function ID by name
	req, err := http.NewRequest("GET", baseURL+"/api/functions/name/"+functionName, nil)
	if err != nil {
		return err
	}

	// Add authentication if API key is provided
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("function not found: %s", resp.Status)
	}

	var function map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&function); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	functionID, ok := function["id"].(string)
	if !ok {
		return fmt.Errorf("invalid function response, missing ID")
	}

	// Then, get the executions for that function with authentication
	req, err = http.NewRequest("GET", baseURL+"/api/executions/function/"+functionID, nil)
	if err != nil {
		return err
	}

	// Add authentication if API key is provided
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Make the request
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to retrieve logs: %s", resp.Status)
	}

	var executions []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&executions); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if len(executions) == 0 {
		fmt.Println("No executions found for this function.")
		return nil
	}

	// Display the logs
	fmt.Printf("Logs for function '%s':\n\n", functionName)
	for i, execution := range executions {
		requestID, _ := execution["request_id"].(string)
		statusCode, _ := execution["status_code"].(float64)
		output, _ := execution["output"].(string)
		errorMsg, _ := execution["error_message"].(string)
		duration, _ := execution["duration_ms"].(float64)

		fmt.Printf("Execution #%d (ID: %s)\n", i+1, requestID)
		fmt.Printf("Status: %d\n", int(statusCode))
		fmt.Printf("Duration: %.2f ms\n", duration)

		if errorMsg != "" {
			fmt.Printf("Error: %s\n", errorMsg)
		}

		fmt.Printf("Output:\n%s\n\n", output)
		fmt.Println("---")
	}

	return nil
}

var generateAPIKeyCmd = &cobra.Command{
	Use:   "generate-api-key",
	Short: "Generate a new API key",
	Run: func(cmd *cobra.Command, args []string) {
		userID, _ := cmd.Flags().GetString("user-id")
		roles, _ := cmd.Flags().GetStringSlice("roles")
		expiresIn, _ := cmd.Flags().GetInt64("expires-in")

		apiKey, err := generateAPIKey(userID, roles, expiresIn)
		if err != nil {
			fmt.Printf("❌ Error generating API key: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ API key generated successfully:\n%s\n", apiKey)
		fmt.Println("\nUse this API key with the --api-key flag in subsequent commands.")
	},
}

func generateAPIKey(userID string, roles []string, expiresIn int64) (string, error) {
	// Prepare the request data
	data := map[string]any{
		"user_id":    userID,
		"roles":      roles,
		"expires_in": expiresIn,
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Send POST request to generate API key
	resp, err := http.Post(
		baseURL+"/api/auth/api-key",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResponse map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			if errMsg, ok := errResponse["error"].(string); ok {
				return "", fmt.Errorf("failed to generate API key: %s", errMsg)
			}
		}
		return "", fmt.Errorf("failed to generate API key, status: %s", resp.Status)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Return the API key
	return result["api_key"].(string), nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
