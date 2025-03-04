package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type FunctionRequest struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	Requirements string `json:"requirements"`
	Config       string `json:"config"`
}

func main() {
	http.HandleFunc("/deploy", deployHandler)
	http.ListenAndServe(":8080", nil)
}

func deployHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req FunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create directory for the function
	functionDir := filepath.Join("functions", req.Name)
	if err := os.MkdirAll(functionDir, 0755); err != nil {
		http.Error(w, "Failed to create function directory", http.StatusInternalServerError)
		return
	}

	// Write the function code to a file
	codeFilePath := filepath.Join(functionDir, "handler.py")
	if err := os.WriteFile(codeFilePath, []byte(req.Code), 0644); err != nil {
		http.Error(w, "Failed to write function code", http.StatusInternalServerError)
		return
	}

	// Write the requirements.txt file
	requirementsFilePath := filepath.Join(functionDir, "requirements.txt")
	if err := os.WriteFile(requirementsFilePath, []byte(req.Requirements), 0644); err != nil {
		http.Error(w, "Failed to write requirements.txt", http.StatusInternalServerError)
		return
	}

	// Write the skyscale.yaml file
	skyscaleFilePath := filepath.Join(functionDir, "skyscale.yaml")
	if err := os.WriteFile(skyscaleFilePath, []byte(req.Config), 0644); err != nil {
		http.Error(w, "Failed to write skyscale.yaml", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Function '%s' deployed successfully.", req.Name)
}
