package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetGlobals() {
	mu.Lock()
	defer mu.Unlock()
	nextID = 1
	tasks = make(map[int]*Calculation)
	queue = []int{}
}

func TestHandleCalculate(t *testing.T) {
	resetGlobals()
	body := bytes.NewBufferString(`{"expression": "1+1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", body)
	w := httptest.NewRecorder()
	handleCalculate(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, res.StatusCode)
	}
	var out map[string]int
	json.NewDecoder(res.Body).Decode(&out)
	if out["id"] != 1 {
		t.Fatalf("expected id 1, got %d", out["id"])
	}
}

func TestHandleCalculateInvalid(t *testing.T) {
	resetGlobals()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(`{"expression":""}`))
	w := httptest.NewRecorder()
	handleCalculate(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected %d, got %d", http.StatusUnprocessableEntity, res.StatusCode)
	}
}

func TestHandleCalculateMethodNotAllowed(t *testing.T) {
	resetGlobals()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/calculate", nil)
	w := httptest.NewRecorder()
	handleCalculate(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, res.StatusCode)
	}
}

func TestHandleListExpressions(t *testing.T) {
	resetGlobals()
	body := bytes.NewBufferString(`{"expression": "2+2"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", body)
	w := httptest.NewRecorder()
	handleCalculate(w, req)
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	wList := httptest.NewRecorder()
	handleListExpressions(wList, reqList)
	res := wList.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, res.StatusCode)
	}
	var out map[string][]*Calculation
	json.NewDecoder(res.Body).Decode(&out)
	if len(out["expressions"]) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(out["expressions"]))
	}
}

func TestHandleGetExpression(t *testing.T) {
	resetGlobals()
	body := bytes.NewBufferString(`{"expression": "3*3"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", body)
	w := httptest.NewRecorder()
	handleCalculate(w, req)
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/1", nil)
	wGet := httptest.NewRecorder()
	handleGetExpression(wGet, reqGet)
	res := wGet.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, res.StatusCode)
	}
	var out map[string]*Calculation
	json.NewDecoder(res.Body).Decode(&out)
	if out["expression"].ID != 1 {
		t.Fatalf("expected id 1, got %d", out["expression"].ID)
	}
}

func TestHandleGetExpressionNotFound(t *testing.T) {
	resetGlobals()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/999", nil)
	w := httptest.NewRecorder()
	handleGetExpression(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestHandleInternalTaskGetPost(t *testing.T) {
	resetGlobals()
	body := bytes.NewBufferString(`{"expression": "4/2"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", body)
	w := httptest.NewRecorder()
	handleCalculate(w, req)
	reqInternal := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
	wInternal := httptest.NewRecorder()
	handleInternalTask(wInternal, reqInternal)
	res := wInternal.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, res.StatusCode)
	}
	var taskResp TaskResponse
	json.NewDecoder(res.Body).Decode(&taskResp)
	if taskResp.Task.ID != 1 || taskResp.Task.Expression != "4/2" {
		t.Fatalf("unexpected task: %+v", taskResp.Task)
	}
	resultPayload := ResultPayload{ID: 1, Result: 2}
	data, _ := json.Marshal(resultPayload)
	reqPost := httptest.NewRequest(http.MethodPost, "/internal/task", bytes.NewBuffer(data))
	wPost := httptest.NewRecorder()
	handleInternalTask(wPost, reqPost)
	resPost := wPost.Result()
	if resPost.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resPost.StatusCode)
	}
	mu.Lock()
	task, exists := tasks[1]
	mu.Unlock()
	if !exists || task.Status != "done" || task.Result == nil || *task.Result != 2 {
		t.Fatalf("task not updated correctly")
	}
}

func TestInternalTaskNoTask(t *testing.T) {
	resetGlobals()
	req := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
	w := httptest.NewRecorder()
	handleInternalTask(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestHandleGetExpressionInvalidID(t *testing.T) {
	resetGlobals()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/abc", nil)
	w := httptest.NewRecorder()
	handleGetExpression(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}

func TestHandleListExpressionsExtra(t *testing.T) {
	resetGlobals()
	expressions := []string{"1+1", "2*3", "10-4"}
	for _, expr := range expressions {
		body := bytes.NewBufferString(`{"expression": "` + expr + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", body)
		w := httptest.NewRecorder()
		handleCalculate(w, req)
	}
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	wList := httptest.NewRecorder()
	handleListExpressions(wList, reqList)
	res := wList.Result()
	var out map[string][]*Calculation
	json.NewDecoder(res.Body).Decode(&out)
	if len(out["expressions"]) != len(expressions) {
		t.Fatalf("expected %d expressions, got %d", len(expressions), len(out["expressions"]))
	}
	idSum := 0
	for _, task := range out["expressions"] {
		idSum += task.ID
	}
	if idSum != 6 {
		t.Fatalf("expected id sum 6, got %d", idSum)
	}
}
