package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var receivedURL string

// Struct to hold the incoming URL from the extension
type RequestData struct {
	URL string `json:"url"`
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}
}

// function that solves the problem and gives string based output
func solveProblem(apiKey, problem string) (string, error) {
	//setting up prompt
	promptTemplate := `You are an expert problem-solving assistant. Your task is to guide users in solving problems step by step without providing any code. 

	Given a LeetCode problem name, provide only a high-level explanation of the steps required to solve the problem. Ensure your response focuses on the logic, strategy, and algorithmic thinking but strictly avoids writing or suggesting any code.
	
	Problem: "{{problem_name}}"
	
	Steps to solve:
	1. 
	2. 
	3. 
	`
	prompt := strings.ReplaceAll(promptTemplate, "{{problem_name}}", problem)

	//setting up gemini API for solving the question of giving hint
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate: %w", err)
	}

	var parts []string
	for _, part := range resp.Candidates[0].Content.Parts {
		parts = append(parts, fmt.Sprintf("%v", part))
	}
	//fmt.Printf("%v", parts)
	result := strings.Join(parts, "\n")
	return result, nil

}

// Function to extract the problem name from the URL
func extractProblemName(problemURL string) ([]string, error) {
	parsedURL, err := url.Parse(problemURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	return strings.Split(parsedURL.Path, "/"), nil
}

func urlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		//RequestData struct to hold incoming url, defined above
		var reqData RequestData

		// Decode the incoming JSON to extract the URL and put it in the reqData struct
		err := json.NewDecoder(r.Body).Decode(&reqData)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Extract the problem name from the URL
		//the problem name will have "-" in between the words
		receivedURL = reqData.URL
		s, err := extractProblemName(receivedURL)
		if err != nil {
			fmt.Println("Error extracting problem name:", err)
		}

		//check if the url is from leetcode or not
		check_Leetcode := false
		if s[1] != "problems" {
			check_Leetcode = true
		}

		// Get the API key from the envvironment
		apiKey := os.Getenv("API_KEY")
		if apiKey == "" {
			log.Fatal("API_KEY not set in evnvironment")
		}

		// Solve the problem and get the hint
		hint := ""
		var err1 error
		if !check_Leetcode {
			problem := string(s[2])
			hint, err1 = solveProblem(apiKey, problem)
			if err1 != nil {
				fmt.Println(err1)
			}
			fmt.Println(hint)
		}

		// Create a response map to send back to the extension
		var response map[string]string
		if check_Leetcode {
			response = map[string]string{
				"Problem_Title": "Go to any leetcode problem page to use this extension",
				"message":       "cannot extract",
				"hint":          "no hints",
			}
		} else {
			response = map[string]string{
				"Problem_Title": string(s[2]),
				"message":       "This is AI generated hint so please verify it",
				"hint":          hint,
			}
		}

		// Set headers and encode the response in JSON format
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// Handle non-POST methods
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
	}
}

func main() {

	http.HandleFunc("/capture-url", urlHandler)

	// Start the server on port 8080
	fmt.Println("Server running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
