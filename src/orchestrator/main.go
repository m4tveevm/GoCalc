package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Calculation struct {
	ID         int      `json:"id"`
	Expression string   `json:"expression"`
	Status     string   `json:"status"`
	Result     *float64 `json:"result,omitempty"`
}

var (
	mu     sync.Mutex
	nextID = 1
	tasks  = make(map[int]*Calculation)
	queue  []int
)

type CalcRequest struct {
	Expression string `json:"expression"`
}

type TaskResponse struct {
	Task struct {
		ID         int    `json:"id"`
		Expression string `json:"expression"`
	} `json:"task"`
}

type ResultPayload struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

func handleCalculate(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req CalcRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil || req.Expression == "" {
		http.Error(writer, `{"error":"Invalid expression"}`, http.StatusUnprocessableEntity)
		return
	}
	mu.Lock()
	id := nextID
	nextID++
	task := &Calculation{
		ID:         id,
		Expression: req.Expression,
		Status:     "pending",
	}
	tasks[id] = task
	queue = append(queue, id)
	mu.Unlock()

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(map[string]int{"id": id})
}

func handleListExpressions(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	mu.Lock()
	var list []*Calculation
	for i := 1; i < nextID; i++ {
		if task, ok := tasks[i]; ok {
			list = append(list, task)
		}
	}
	mu.Unlock()
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]interface{}{"expressions": list})
}

func handleGetExpression(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(request.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(writer, `{"error":"ID not provided"}`, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[4])
	if err != nil {
		http.Error(writer, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}
	mu.Lock()
	task, exists := tasks[id]
	mu.Unlock()
	if !exists {
		http.Error(writer, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]*Calculation{"expression": task})
}

func handleInternalTask(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		mu.Lock()
		if len(queue) == 0 {
			mu.Unlock()
			writer.WriteHeader(http.StatusNotFound)
			json.NewEncoder(writer).Encode(map[string]string{"error": "No task available"})
			return
		}
		id := queue[0]
		queue = queue[1:]
		task := tasks[id]
		task.Status = "in_progress"
		mu.Unlock()

		writer.Header().Set("Content-Type", "application/json")
		var resp TaskResponse
		resp.Task.ID = task.ID
		resp.Task.Expression = task.Expression
		json.NewEncoder(writer).Encode(resp)
	} else if request.Method == http.MethodPost {
		var res ResultPayload
		if err := json.NewDecoder(request.Body).Decode(&res); err != nil {
			http.Error(writer, `{"error":"Invalid data"}`, http.StatusUnprocessableEntity)
			return
		}
		mu.Lock()
		task, exists := tasks[res.ID]
		if !exists {
			mu.Unlock()
			http.Error(writer, `{"error":"Task not found"}`, http.StatusNotFound)
			return
		}
		task.Result = &res.Result
		task.Status = "done"
		mu.Unlock()
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(map[string]string{"status": "result accepted"})
	} else {
		http.Error(writer, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", handleCalculate)
	mux.HandleFunc("/api/v1/expressions", handleListExpressions)
	mux.HandleFunc("/api/v1/expressions/", handleGetExpression)
	mux.HandleFunc("/internal/task", handleInternalTask)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Println("Orchestrator running on :8080")
	log.Fatal(srv.ListenAndServe())
}
