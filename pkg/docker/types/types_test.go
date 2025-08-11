package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLanguageFromFile(t *testing.T) {
	tests := []struct {
		name string
		filePath string
		expected Language
	}{
		{name: "go file", filePath: "test.go", expected: LanguageGo},
		{name: "py file", filePath: "test.py", expected: LanguagePy},
		{name: "js file", filePath: "test.js", expected: LanguageJS},
		{name: "ts file", filePath: "test.ts", expected: LanguageTS},
		{name: "node file", filePath: "test.mjs", expected: LanguageNode},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetLanguageFromFile(test.filePath)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetLanguageFromExtension(t *testing.T) {
	tests := []struct {
		name string
		extension string
		expected Language
	}{
		{name: "go extension", extension: ".go", expected: LanguageGo},
		{name: "py extension", extension: ".py", expected: LanguagePy},
		{name: "js extension", extension: "js", expected: LanguageJS},
		{name: "ts extension", extension: "ts", expected: LanguageTS},
		{name: "node extension", extension: "mjs", expected: LanguageNode},
		{name: "unknown extension", extension: ".unknown", expected: LanguageGo},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetLanguageFromExtension(test.extension)
			assert.Equal(t, test.expected, actual)
		})
	}
}