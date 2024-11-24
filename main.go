package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var receivedURL string

// Struct to hold the incoming URL from the extension
type RequestData struct {
	URL         string `json:"url"`
	Description string `json:"description`
	APIKey      string `json:"apiKey"`
}

// function that solves the problem and gives string based output
func solveProblem(apiKey, problem, description string) (string, error) {
	//setting up prompt
	promptTemplate := `You are an expert problem-solving assistant. Your task is to guide users in solving LeetCode problems step by step without providing any code.

    Problem: "{{problem_name}}"
    
    User's Description/Approach: "{{user_description}}"

    Based on the user's input, provide:
    1. An analysis of their current approach (if provided)
    2. Step-by-step hints and guidance
    3. Common pitfalls to avoid
    4. Optimization suggestions

    Focus on the logic, strategy, and algorithmic thinking but strictly avoid writing or suggesting any code.
    
    Steps to solve:`

	prompt := strings.ReplaceAll(promptTemplate, "{{problem_name}}", problem)
	prompt = strings.ReplaceAll(prompt, "{{user_description}}", description)

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

func formatstring(title string) string {
	words := strings.Split(title, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper((word[:1])) + word[1:]
		}
	}
	return strings.Join(words, " ")
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

		if reqData.APIKey == "" {
			http.Error(w, "API key is required", http.StatusBadRequest)
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
		// apiKey := os.Getenv("API_KEY")
		// if apiKey == "" {
		// 	log.Fatal("API_KEY not set in evnvironment")
		// }

		// Solve the problem and get the hint
		hint := ""
		var err1 error
		if !check_Leetcode {
			problem := string(s[2])
			hint, err1 = solveProblem(reqData.APIKey, problem, reqData.Description)
			if err1 != nil {
				fmt.Println(err1)
			}
			//fmt.Println(hint)
			fmt.Println("Done reviewing the problem")
		}

		// Create a response map to send back to the extension
		var response map[string]string
		if check_Leetcode {
			response = map[string]string{
				"Problem_Title": "Go to any leetcode problem page to use this extension",
				"message":       "Cannot extract any problem name from the URL",
				"hint":          "No hint available",
			}
		} else {
			response = map[string]string{
				"Problem_Title": formatstring(string(s[2])),
				"message":       "This hint is AI-generated. Please review LeetCode’s terms and conditions to ensure compliance before use.",
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
