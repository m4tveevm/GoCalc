package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/m4tveevm/GoCalc/calc"
)

func TestCalculateHandler(t *testing.T) {
	calculator := calc.NewBasicCalculator()
	handler := CalculateHandler(calculator)

	testCases := []struct {
		name         string
		requestBody  string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Valid expression",
			requestBody:  `{"expression": "3 + 4 * 5"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"result":23}`,
		},
		{
			name:         "Invalid expression",
			requestBody:  `{"expression": "3 + invalid"}`,
			expectedCode: http.StatusUnprocessableEntity,
			expectedBody: `{"error":"Expression is not valid"}`,
		},
		{
			name:         "Empty expression",
			requestBody:  `{"expression": ""}`,
			expectedCode: http.StatusUnprocessableEntity,
			expectedBody: `{"error":"Expression is not valid"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler(rec, req)

			res := rec.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}
			}(res.Body)

			if res.StatusCode != tc.expectedCode {
				t.Errorf("expected status code %v, got %v", tc.expectedCode, res.StatusCode)
			}

			var body bytes.Buffer
			_, err := body.ReadFrom(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if strings.TrimSpace(body.String()) != tc.expectedBody {
				t.Errorf("expected body %v, got %v", tc.expectedBody, strings.TrimSpace(body.String()))
			}
		})
	}
}
