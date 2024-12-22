package main

import (
	"encoding/json"
	"fmt"
	calculator "github.com/m4tveevm/GoCalc/calc"
	"log"
	"net/http"
)

func calculateHandler(calc calculator.Calculator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			Expression string `json:"expression"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Expression == "" {
			http.Error(w, `{"error":"Expression is not valid"}`, http.StatusUnprocessableEntity)
			return
		}

		result, err := calc.Calculate(request.Expression)
		if err != nil {
			http.Error(w, `{"error":"Expression is not valid"}`, http.StatusUnprocessableEntity)
			return
		}

		response := struct {
			Result float64 `json:"result"`
		}{
			Result: result,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			log.Printf("Failed to encode response: %v", err)
			return
		}
	}
}

func main() {
	calc := calculator.NewBasicCalculator()
	http.HandleFunc("/api/v1/calculate", calculateHandler(calc))
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
