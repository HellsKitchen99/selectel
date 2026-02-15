package loglint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// Тест checkLowerCase - Успех
func TestCheckLowerCaseSuccess(t *testing.T) {
	// preparing
	msg := "the bay harbour butcher"
	expectedResult := true

	// test
	result := checkLowerCase(msg)

	// assert
	if result != expectedResult {
		t.Errorf("expected result - %v", expectedResult)
	}
}

// Тест checkLowerCase - Провал
func TestCheckLowerCaseFailure(t *testing.T) {
	// preparing
	msg := "The Bay Harbour Butcher"
	expectedResult := false

	// test
	result := checkLowerCase(msg)

	// assert
	if result != expectedResult {
		t.Errorf("expected result - %v", expectedResult)
	}
}

// Тест checkEnglish - Успех
func TestCheckEnglishSuccess(t *testing.T) {
	// preparing
	msg := "something on english"
	expectedResult := true

	// test
	result := checkEnglish(msg)

	// assert
	if result != expectedResult {
		t.Errorf("expected result - %v", expectedResult)
	}
}

// Тест checkEnglish - Провал
func TestCheckEnglishFailure(t *testing.T) {
	// preparing
	msg := "что то на русском"
	expectedResult := false

	// test
	result := checkEnglish(msg)

	// assert
	if result != expectedResult {
		t.Errorf("expected result - %v", expectedResult)
	}
}

// Тест checkNoSpecialChars - Успех
func TestCheckNoSpecialCharsSuccess(t *testing.T) {
	// preparing
	msg := "abc_-123 "
	expectedResult := true

	// test
	result := checkNoSpecialChars(msg)

	// assert
	if result != expectedResult {
		t.Errorf("expected result - %v", expectedResult)
	}
}

// Тест checkNoSpecialChars - Провал
func TestCheckNoSpecialCharsFailure(t *testing.T) {
	// preparing
	msg := "@%😶‍🌫️🥶"
	expectedResult := false

	// test
	result := checkNoSpecialChars(msg)

	// assert
	if result != expectedResult {
		t.Errorf("expected result - %v", expectedResult)
	}
}

// Тест checkSensitive - Успех
func TestCheckSensitiveSuccess(t *testing.T) {
	// preparing
	word := "Smth"
	exprs, err := checkSensitivePreparing(word)
	if err != nil {
		t.Errorf("error while trying to parse go code to ast.File: %v", err)
	}
	expectedResult := true

	// test
	result := checkSensitive(exprs)

	// assert
	if result != expectedResult {
		t.Errorf("expected result - %v", expectedResult)
	}
}

// Тест checkSensitive - Провал
func TestCheckSensitiveFailureBasicLit(t *testing.T) {
	// preparing

	// test

	// assert
}

func checkSensitivePreparing(word string) ([]ast.Expr, error) {
	msg := fmt.Sprintf(`package main

		func main() {
			fmt.Println("%v")
		}`, word)
	set := token.NewFileSet()
	node, err := parser.ParseFile(set, "", msg, 0)
	if err != nil {
		return []ast.Expr{}, err
	}

	var exprs []ast.Expr
	ast.Inspect(node, func(n ast.Node) bool {
		exp, ok := n.(ast.Expr)
		if !ok {
			return false
		}
		exprs = append(exprs, exp)
		return true
	})
	return exprs, nil
}
