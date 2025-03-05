package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/m4tveevm/GoCalc/calc"
)

type Task struct {
	ID         int    `json:"id"`
	Expression string `json:"expression"`
}

type TaskResponse struct {
	Task Task `json:"task"`
}

type ResultPayload struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

func worker(workerID int, orchestratorURL string, pollInterval time.Duration) {
	client := &http.Client{Timeout: 5 * time.Second}
	for {
		resp, err := client.Get(orchestratorURL + "/internal/task")
		if err != nil {
			log.Printf("[Worker %d] Error fetching task: %v", workerID, err)
			time.Sleep(pollInterval)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			time.Sleep(pollInterval)
			continue
		}
		var taskResp TaskResponse
		if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
			log.Printf("[Worker %d] Error decoding task: %v", workerID, err)
			resp.Body.Close()
			time.Sleep(pollInterval)
			continue
		}
		resp.Body.Close()
		log.Printf("[Worker %d] Received task %d: %s", workerID, taskResp.Task.ID, taskResp.Task.Expression)

		calculator := calc.NewBasicCalculator()
		result, err := calculator.Calculate(taskResp.Task.Expression)
		if err != nil {
			log.Printf("[Worker %d] Error computing expression: %v", workerID, err)
			continue
		}
		delay := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
		time.Sleep(delay)

		resPayload := ResultPayload{
			ID:     taskResp.Task.ID,
			Result: result,
		}
		data, _ := json.Marshal(resPayload)
		res, err := client.Post(orchestratorURL+"/internal/task", "application/json", bytes.NewReader(data))
		if err != nil {
			log.Printf("[Worker %d] Error sending result: %v", workerID, err)
			continue
		}
		res.Body.Close()
		log.Printf("[Worker %d] Sent result for task %d: %v", workerID, taskResp.Task.ID, result)
	}
}

func main() {
	workers := 1
	if val := os.Getenv("COMPUTING_POWER"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			workers = n
		}
	}
	orchestratorURL := os.Getenv("ORCHESTRATOR_URL")
	if orchestratorURL == "" {
		orchestratorURL = "http://orchestrator:8080"
	}
	pollInterval := 2 * time.Second

	log.Printf("Agent started with %d workers", workers)
	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= workers; i++ {
		go worker(i, orchestratorURL, pollInterval)
	}
	select {}
}
