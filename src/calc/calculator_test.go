package calc

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	calculator := NewBasicCalculator()

	testCases := []struct {
		expression string
		expected   float64
		expectErr  bool
	}{
		{"3 + 4 * 5", 23, false},
		{"(3 + 4) * 5", 35, false},
		{"6 / 2 * 3", 9, false},
		{"10 - 3 + 5", 12, false},
		{"", 0, true},
		{"invalid_string", 0, true},
		{"(3 + 4", 0, true},
		{"3 + 4)", 0, true},
	}

	for _, tc := range testCases {
		result, err := calculator.Calculate(tc.expression)
		if tc.expectErr {
			if err == nil {
				t.Errorf("expected error for expression %q, got none", tc.expression)
			}
		} else {
			if err != nil {
				t.Errorf("did not expect error for expression %q, got %v", tc.expression, err)
			}
			if result != tc.expected {
				t.Errorf("expected %v for expression %q, got %v", tc.expected, tc.expression, result)
			}
		}
	}
}
