package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBigInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "large job ID",
			value:    "300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: "300749528249665590178224313442040528409305273634097553067152835846309151049",
		},
		{
			name:     "zero",
			value:    "0",
			expected: "0",
		},
		{
			name:     "one",
			value:    "1",
			expected: "1",
		},
		{
			name:     "negative one",
			value:    "-1",
			expected: "-1",
		},
		{
			name:     "large negative number",
			value:    "-300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: "-300749528249665590178224313442040528409305273634097553067152835846309151049",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBigInt(tt.value)
			assert.NoError(t, err)

			expected := new(big.Int)
			expected.SetString(tt.expected, 10)
			assert.Equal(t, 0, result.Cmp(expected), "Expected %s but got %s", tt.expected, result.String())
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		expected string
	}{
		{
			name:     "large addition",
			x:        "300749528249665590178224313442040528409305273634097553067152835846309151049",
			y:        "300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: "601499056499331180356448626884081056818610547268195106134305671692618302098",
		},
		{
			name:     "zero addition",
			x:        "0",
			y:        "0",
			expected: "0",
		},
		{
			name:     "add to zero",
			x:        "100",
			y:        "0",
			expected: "100",
		},
		{
			name:     "negative addition",
			x:        "-50",
			y:        "-25",
			expected: "-75",
		},
		{
			name:     "mixed signs",
			x:        "100",
			y:        "-50",
			expected: "50",
		},
		{
			name:     "invalid x",
			x:        "invalid",
			y:        "100",
			expected: "",
		},
		{
			name:     "invalid y",
			x:        "100",
			y:        "invalid",
			expected: "",
		},
		{
			name:     "empty x",
			x:        "",
			y:        "100",
			expected: "",
		},
		{
			name:     "empty y",
			x:        "100",
			y:        "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.x, tt.y)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		expected string
	}{
		{
			name:     "large subtraction",
			x:        "601499056499331180356448626884081056818610547268195106134305671692618302098",
			y:        "300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: "300749528249665590178224313442040528409305273634097553067152835846309151049",
		},
		{
			name:     "zero subtraction",
			x:        "100",
			y:        "0",
			expected: "100",
		},
		{
			name:     "subtract to zero",
			x:        "100",
			y:        "100",
			expected: "0",
		},
		{
			name:     "negative result",
			x:        "50",
			y:        "100",
			expected: "-50",
		},
		{
			name:     "negative subtraction",
			x:        "-50",
			y:        "-25",
			expected: "-25",
		},
		{
			name:     "invalid x",
			x:        "invalid",
			y:        "100",
			expected: "",
		},
		{
			name:     "invalid y",
			x:        "100",
			y:        "invalid",
			expected: "",
		},
		{
			name:     "empty x",
			x:        "",
			y:        "100",
			expected: "",
		},
		{
			name:     "empty y",
			x:        "100",
			y:        "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sub(tt.x, tt.y)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMul(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		expected string
	}{
		{
			name:     "large multiplication",
			x:        "1000000000000000000",
			y:        "1000000000000000000",
			expected: "1000000000000000000000000000000000000",
		},
		{
			name:     "multiply by zero",
			x:        "100",
			y:        "0",
			expected: "0",
		},
		{
			name:     "multiply by one",
			x:        "100",
			y:        "1",
			expected: "100",
		},
		{
			name:     "negative multiplication",
			x:        "-10",
			y:        "5",
			expected: "-50",
		},
		{
			name:     "both negative",
			x:        "-10",
			y:        "-5",
			expected: "50",
		},
		{
			name:     "invalid x",
			x:        "invalid",
			y:        "100",
			expected: "",
		},
		{
			name:     "invalid y",
			x:        "100",
			y:        "invalid",
			expected: "",
		},
		{
			name:     "empty x",
			x:        "",
			y:        "100",
			expected: "",
		},
		{
			name:     "empty y",
			x:        "100",
			y:        "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Mul(tt.x, tt.y)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiv(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		expected string
	}{
		{
			name:     "large division",
			x:        "1000000000000000000000000000000000000",
			y:        "1000000000000000000",
			expected: "1000000000000000000",
		},
		{
			name:     "divide by one",
			x:        "100",
			y:        "1",
			expected: "100",
		},
		{
			name:     "divide to zero",
			x:        "0",
			y:        "100",
			expected: "0",
		},
		{
			name:     "integer division",
			x:        "10",
			y:        "3",
			expected: "3",
		},
		{
			name:     "negative division",
			x:        "-100",
			y:        "10",
			expected: "-10",
		},
		{
			name:     "both negative",
			x:        "-100",
			y:        "-10",
			expected: "10",
		},
		{
			name:     "invalid x",
			x:        "invalid",
			y:        "100",
			expected: "",
		},
		{
			name:     "invalid y",
			x:        "100",
			y:        "invalid",
			expected: "",
		},
		{
			name:     "empty x",
			x:        "",
			y:        "100",
			expected: "",
		},
		{
			name:     "empty y",
			x:        "100",
			y:        "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Div(tt.x, tt.y)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsEqual(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		expected bool
	}{
		{
			name:     "equal large numbers",
			x:        "300749528249665590178224313442040528409305273634097553067152835846309151049",
			y:        "300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: true,
		},
		{
			name:     "equal zero",
			x:        "0",
			y:        "0",
			expected: true,
		},
		{
			name:     "not equal",
			x:        "100",
			y:        "200",
			expected: false,
		},
		{
			name:     "equal negative",
			x:        "-100",
			y:        "-100",
			expected: true,
		},
		{
			name:     "invalid x",
			x:        "invalid",
			y:        "100",
			expected: false,
		},
		{
			name:     "invalid y",
			x:        "100",
			y:        "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEqual(tt.x, tt.y)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLess(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		expected bool
	}{
		{
			name:     "less than",
			x:        "100",
			y:        "200",
			expected: true,
		},
		{
			name:     "not less than",
			x:        "200",
			y:        "100",
			expected: false,
		},
		{
			name:     "equal not less",
			x:        "100",
			y:        "100",
			expected: false,
		},
		{
			name:     "negative less than positive",
			x:        "-100",
			y:        "100",
			expected: true,
		},
		{
			name:     "large numbers less than",
			x:        "100749528249665590178224313442040528409305273634097553067152835846309151049",
			y:        "300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: true,
		},
		{
			name:     "invalid x",
			x:        "invalid",
			y:        "100",
			expected: false,
		},
		{
			name:     "invalid y",
			x:        "100",
			y:        "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLess(tt.x, tt.y)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsGreater(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		expected bool
	}{
		{
			name:     "greater than",
			x:        "200",
			y:        "100",
			expected: true,
		},
		{
			name:     "not greater than",
			x:        "100",
			y:        "200",
			expected: false,
		},
		{
			name:     "equal not greater",
			x:        "100",
			y:        "100",
			expected: false,
		},
		{
			name:     "positive greater than negative",
			x:        "100",
			y:        "-100",
			expected: true,
		},
		{
			name:     "large numbers greater than",
			x:        "300749528249665590178224313442040528409305273634097553067152835846309151049",
			y:        "100749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: true,
		},
		{
			name:     "invalid x",
			x:        "invalid",
			y:        "100",
			expected: false,
		},
		{
			name:     "invalid y",
			x:        "100",
			y:        "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGreater(tt.x, tt.y)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "zero",
			value:    "0",
			expected: true,
		},
		{
			name:     "positive not zero",
			value:    "100",
			expected: false,
		},
		{
			name:     "negative not zero",
			value:    "-100",
			expected: false,
		},
		{
			name:     "invalid",
			value:    "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsZero(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNegative(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "negative",
			value:    "-100",
			expected: true,
		},
		{
			name:     "positive not negative",
			value:    "100",
			expected: false,
		},
		{
			name:     "zero not negative",
			value:    "0",
			expected: false,
		},
		{
			name:     "large negative",
			value:    "-300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: true,
		},
		{
			name:     "invalid",
			value:    "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNegative(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPositive(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "positive",
			value:    "100",
			expected: true,
		},
		{
			name:     "negative not positive",
			value:    "-100",
			expected: false,
		},
		{
			name:     "zero not positive",
			value:    "0",
			expected: false,
		},
		{
			name:     "large positive",
			value:    "300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: true,
		},
		{
			name:     "invalid",
			value:    "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPositive(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseBigIntErrors(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "invalid string",
			value: "invalid",
		},
		{
			name:  "empty string",
			value: "",
		},
		{
			name:  "hexadecimal",
			value: "0x123",
		},
		{
			name:  "float",
			value: "123.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBigInt(tt.value)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}
