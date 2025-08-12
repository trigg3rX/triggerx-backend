package types

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBigInt_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "large job ID",
			value:    "300749528249665590178224313442040528409305273634097553067152835846309151049",
			expected: `"300749528249665590178224313442040528409305273634097553067152835846309151049"`,
		},
		{
			name:     "zero",
			value:    "0",
			expected: `"0"`,
		},
		{
			name:     "one",
			value:    "1",
			expected: `"1"`,
		},
		{
			name:     "large number",
			value:    "1234567890123456789012345678901234567890",
			expected: `"1234567890123456789012345678901234567890"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the value
			i, ok := new(big.Int).SetString(tt.value, 10)
			require.True(t, ok, "Failed to parse test value")

			// Create BigInt
			bigInt := NewBigInt(i)

			// Marshal to JSON
			data, err := json.Marshal(bigInt)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Unmarshal back
			var result BigInt
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.value, result.String())
		})
	}
}

func TestBigInt_JSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		expected    string
	}{
		{
			name:        "valid large job ID",
			jsonData:    `"300749528249665590178224313442040528409305273634097553067152835846309151049"`,
			expectError: false,
			expected:    "300749528249665590178224313442040528409305273634097553067152835846309151049",
		},
		{
			name:        "valid zero",
			jsonData:    `"0"`,
			expectError: false,
			expected:    "0",
		},
		{
			name:        "null value",
			jsonData:    `null`,
			expectError: false,
			expected:    "<nil>",
		},
		{
			name:        "invalid number format",
			jsonData:    `300749528249665590178224313442040528409305273634097553067152835846309151049`,
			expectError: true,
		},
		{
			name:        "invalid string",
			jsonData:    `"not_a_number"`,
			expectError: true,
		},
		{
			name:        "empty string",
			jsonData:    `""`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result BigInt
			err := json.Unmarshal([]byte(tt.jsonData), &result)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.String())
			}
		})
	}
}

func TestBigInt_StructMarshaling(t *testing.T) {
	type TestStruct struct {
		JobID *BigInt `json:"job_id"`
		Name  string  `json:"name"`
	}

	// Test with the problematic job ID from the error
	jobIDStr := "300749528249665590178224313442040528409305273634097553067152835846309151049"
	jobID, _ := new(big.Int).SetString(jobIDStr, 10)

	testStruct := TestStruct{
		JobID: NewBigInt(jobID),
		Name:  "test job",
	}

	// Marshal
	data, err := json.Marshal(testStruct)
	require.NoError(t, err)

	expected := `{"job_id":"300749528249665590178224313442040528409305273634097553067152835846309151049","name":"test job"}`
	assert.Equal(t, expected, string(data))

	// Unmarshal back
	var result TestStruct
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, jobIDStr, result.JobID.String())
	assert.Equal(t, "test job", result.Name)
}

func TestBigInt_ConversionMethods(t *testing.T) {
	// Test conversion methods
	original := big.NewInt(123456789)
	bigInt := NewBigInt(original)

	// Test ToBigInt
	converted := bigInt.ToBigInt()
	assert.Equal(t, original, converted)

	// Test FromBigInt
	fromBigInt := FromBigInt(original)
	assert.Equal(t, original, fromBigInt.ToBigInt())

	// Test String
	assert.Equal(t, "123456789", bigInt.String())

	// Test Clone
	cloned := bigInt.Clone()
	assert.Equal(t, bigInt.String(), cloned.String())
	assert.NotSame(t, bigInt.Int, cloned.Int) // Should be different pointers
}

func TestBigInt_ArithmeticOperations(t *testing.T) {
	a := NewBigInt(big.NewInt(100))
	b := NewBigInt(big.NewInt(50))

	// Test Add
	result := NewBigInt(big.NewInt(0))
	result.Add(a, b)
	assert.Equal(t, "150", result.String())

	// Test Sub
	result = NewBigInt(big.NewInt(0))
	result.Sub(a, b)
	assert.Equal(t, "50", result.String())

	// Test Mul
	result = NewBigInt(big.NewInt(0))
	result.Mul(a, b)
	assert.Equal(t, "5000", result.String())

	// Test Div
	result = NewBigInt(big.NewInt(0))
	result.Div(a, b)
	assert.Equal(t, "2", result.String())
}

func TestBigInt_ComparisonMethods(t *testing.T) {
	a := NewBigInt(big.NewInt(100))
	b := NewBigInt(big.NewInt(50))
	c := NewBigInt(big.NewInt(100))

	// Test Cmp
	assert.Equal(t, 1, a.Cmp(b))  // a > b
	assert.Equal(t, -1, b.Cmp(a)) // b < a
	assert.Equal(t, 0, a.Cmp(c))  // a == c

	// Test comparison methods
	assert.True(t, a.Greater(b))
	assert.True(t, b.Less(a))
	assert.True(t, a.Equal(c))
	assert.False(t, a.Equal(b))
}

func TestBigInt_UtilityMethods(t *testing.T) {
	zero := NewBigInt(big.NewInt(0))
	positive := NewBigInt(big.NewInt(100))
	negative := NewBigInt(big.NewInt(-100))
	nilBigInt := NewBigInt(nil)

	// Test IsZero
	assert.True(t, zero.IsZero())
	assert.False(t, positive.IsZero())
	assert.False(t, negative.IsZero())
	assert.True(t, nilBigInt.IsZero())

	// Test IsPositive
	assert.False(t, zero.IsPositive())
	assert.True(t, positive.IsPositive())
	assert.False(t, negative.IsPositive())
	assert.False(t, nilBigInt.IsPositive())

	// Test IsNegative
	assert.False(t, zero.IsNegative())
	assert.False(t, positive.IsNegative())
	assert.True(t, negative.IsNegative())
	assert.False(t, nilBigInt.IsNegative())
}
