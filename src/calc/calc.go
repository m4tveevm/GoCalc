package calc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Calculator interface {
	Calculate(expression string) (float64, error)
}

type BasicCalculator struct{}

func NewBasicCalculator() *BasicCalculator {
	return &BasicCalculator{}
}

func (c *BasicCalculator) Calculate(expression string) (float64, error) {
	if err := validateParentheses(expression); err != nil {
		return 0, err
	}
	tokens, err := getTokenString(expression)
	if err != nil {
		return 0, err
	}
	postfix, err := infixToPostfix(tokens)
	if err != nil {
		return 0, err
	}
	return evaluatePostfix(postfix)
}

func validateParentheses(expression string) error {
	var stack []rune
	for _, ch := range expression {
		if ch == '(' {
			stack = append(stack, ch)
		} else if ch == ')' {
			if len(stack) == 0 {
				return errors.New("string is not valid")
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) != 0 {
		return errors.New("string is not valid")
	}
	return nil
}

func getTokenString(line string) ([]string, error) {
	var tokens []string
	var number strings.Builder

	for i := 0; i < len(line); i++ {
		elem := line[i]
		if unicode.IsDigit(rune(elem)) || elem == '.' {
			number.WriteByte(elem)
		} else if unicode.IsSpace(rune(elem)) {
			continue
		} else {
			if number.Len() > 0 {
				tokens = append(tokens, number.String())
				number.Reset()
			}
			if strings.ContainsRune("+-*/()", rune(elem)) {
				tokens = append(tokens, string(elem))
			} else {
				return nil, fmt.Errorf("undefined token: %c", elem)
			}
		}
	}
	if number.Len() > 0 {
		tokens = append(tokens, number.String())
	}
	return tokens, nil
}

func infixToPostfix(tokens []string) ([]string, error) {
	var answ []string
	var stack []string

	precedence := func(oper string) int {
		switch oper {
		case "+", "-":
			return 1
		case "*", "/":
			return 2
		default:
			return 0
		}
	}

	for _, token := range tokens {
		if isNumeric(token) {
			answ = append(answ, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				answ = append(answ, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1]
		} else if isOperator(token) {
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if isOperator(top) && precedence(top) >= precedence(token) {
					answ = append(answ, top)
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, token)
		} else {
			return nil, fmt.Errorf("undefined token: %s", token)
		}
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		answ = append(answ, top)
		stack = stack[:len(stack)-1]
	}
	return answ, nil
}

func evaluatePostfix(postfix []string) (float64, error) {
	var stack []float64

	for _, token := range postfix {
		if isNumeric(token) {
			num, _ := strconv.ParseFloat(token, 64)
			stack = append(stack, num)
		} else if isOperator(token) {
			if len(stack) < 2 {
				return 0, errors.New("undefined line")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			var res float64
			switch token {
			case "+":
				res = a + b
			case "-":
				res = a - b
			case "*":
				res = a * b
			case "/":
				if b == 0 {
					return 0, errors.New("division by zero")
				}
				res = a / b
			default:
				return 0, fmt.Errorf("undefined token: %s", token)
			}
			stack = append(stack, res)
		} else {
			return 0, fmt.Errorf("undefined token: %s", token)
		}
	}

	if len(stack) != 1 {
		return 0, errors.New("undefined line")
	}
	return stack[0], nil
}

func isNumeric(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

func isOperator(token string) bool {
	return strings.Contains("+-*/", token)
}
