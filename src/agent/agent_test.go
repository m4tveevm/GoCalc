package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type FakeOrchestrator struct {
	mu        sync.Mutex
	taskSent  bool
	postedID  int
	postedRes float64
}

func (f *FakeOrchestrator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/internal/task" {
		f.mu.Lock()
		if !f.taskSent {
			resp := map[string]interface{}{
				"task": map[string]interface{}{
					"id":         42,
					"expression": "2+2",
				},
			}
			f.taskSent = true
			f.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		f.mu.Unlock()
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "No task available"})
	} else if r.Method == http.MethodPost && r.URL.Path == "/internal/task" {
		var rp ResultPayload
		json.NewDecoder(r.Body).Decode(&rp)
		f.mu.Lock()
		f.postedID = rp.ID
		f.postedRes = rp.Result
		f.mu.Unlock()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "result accepted"})
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func TestWorker(t *testing.T) {
	fake := &FakeOrchestrator{}
	srv := httptest.NewServer(fake)
	defer srv.Close()
	go worker(1, srv.URL, 100*time.Millisecond)
	time.Sleep(3 * time.Second)
	fake.mu.Lock()
	postedID := fake.postedID
	postedRes := fake.postedRes
	fake.mu.Unlock()
	if postedID != 42 {
		t.Fatalf("expected posted id 42, got %d", postedID)
	}
	if postedRes != 4 {
		t.Fatalf("expected posted result 4, got %v", postedRes)
	}
}

func TestWorkerNoTask(t *testing.T) {
	fake := &FakeOrchestrator{}
	fake.taskSent = true
	srv := httptest.NewServer(fake)
	defer srv.Close()
	doneCh := make(chan bool)
	go func() {
		worker(2, srv.URL, 50*time.Millisecond)
		doneCh <- true
	}()
	select {
	case <-doneCh:
		t.Fatal("worker terminated unexpectedly")
	case <-time.After(500 * time.Millisecond):
		// expected: no task available, worker keeps polling
	}
}

func TestMultipleWorker(t *testing.T) {
	fake := &FakeOrchestrator{}
	srv := httptest.NewServer(fake)
	defer srv.Close()
	for i := 1; i <= 3; i++ {
		go worker(i, srv.URL, 100*time.Millisecond)
	}
	time.Sleep(3 * time.Second)
	fake.mu.Lock()
	id := fake.postedID
	res := fake.postedRes
	fake.mu.Unlock()
	if id != 42 {
		t.Fatalf("expected id 42, got %d", id)
	}
	if res != 4 {
		t.Fatalf("expected result 4, got %v", res)
	}
}

func TestFakeOrchestratorEndpoints(t *testing.T) {
	fake := &FakeOrchestrator{}
	reqGet := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
	wGet := httptest.NewRecorder()
	fake.ServeHTTP(wGet, reqGet)
	resGet := wGet.Result()
	if resGet.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resGet.StatusCode)
	}
	var tr TaskResponse
	json.NewDecoder(resGet.Body).Decode(&tr)
	if tr.Task.ID != 42 {
		t.Fatalf("expected task id 42, got %d", tr.Task.ID)
	}
	payload := ResultPayload{ID: 42, Result: 4}
	data, _ := json.Marshal(payload)
	reqPost := httptest.NewRequest(http.MethodPost, "/internal/task", bytes.NewBuffer(data))
	wPost := httptest.NewRecorder()
	fake.ServeHTTP(wPost, reqPost)
	resPost := wPost.Result()
	if resPost.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resPost.StatusCode)
	}
	bodyBytes, _ := ioutil.ReadAll(resPost.Body)
	if !strings.Contains(string(bodyBytes), "result accepted") {
		t.Fatalf("unexpected response: %s", string(bodyBytes))
	}
}
