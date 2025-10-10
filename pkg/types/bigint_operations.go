package types

import (
	"math/big"
	"strconv"
)

// ParseBigInt parses a string as a big.Int
func ParseBigInt(s string) (*big.Int, error) {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, &strconv.NumError{
			Func: "ParseBigInt",
			Num:  s,
			Err:  strconv.ErrSyntax,
		}
	}
	return i, nil
}

// Add adds x and y and stores the result in b
func Add(x, y string) string {
	i, ok := new(big.Int).SetString(x, 10)
	if !ok {
		return ""
	}
	j, ok := new(big.Int).SetString(y, 10)
	if !ok {
		return ""
	}
	if x == "" {
		return ""
	}
	if y == "" {
		return ""
	}
	i.Add(i, j)
	return i.String()
}

// Sub subtracts y from x and stores the result in b
func Sub(x, y string) string {
	i, ok := new(big.Int).SetString(x, 10)
	if !ok {
		return ""
	}
	j, ok := new(big.Int).SetString(y, 10)
	if !ok {
		return ""
	}
	if x == "" {
		return ""
	}
	if y == "" {
		return ""
	}
	i.Sub(i, j)
	return i.String()
}

// Mul multiplies x and y and stores the result in b
func Mul(x, y string) string {
	i, ok := new(big.Int).SetString(x, 10)
	if !ok {
		return ""
	}
	j, ok := new(big.Int).SetString(y, 10)
	if !ok {
		return ""
	}
	if x == "" {
		return ""
	}
	if y == "" {
		return ""
	}
	i.Mul(i, j)
	return i.String()
}

// Div divides x by y and stores the result in b
func Div(x, y string) string {
	i, ok := new(big.Int).SetString(x, 10)
	if !ok {
		return ""
	}
	j, ok := new(big.Int).SetString(y, 10)
	if !ok {
		return ""
	}
	if x == "" {
		return ""
	}
	if y == "" {
		return ""
	}
	i.Div(i, j)
	return i.String()
}

// Equal returns true if x equals y
func IsEqual(x string, y string) bool {
	i, ok := new(big.Int).SetString(x, 10)
	if !ok {
		return false
	}
	j, ok := new(big.Int).SetString(y, 10)
	if !ok {
		return false
	}
	return i.Cmp(j) == 0
}

// Less returns true if x is less than y
func IsLess(x string, y string) bool {
	i, ok := new(big.Int).SetString(x, 10)
	if !ok {
		return false
	}
	j, ok := new(big.Int).SetString(y, 10)
	if !ok {
		return false
	}
	return i.Cmp(j) < 0
}

// Greater returns true if x is greater than y
func IsGreater(x string, y string) bool {
	i, ok := new(big.Int).SetString(x, 10)
	if !ok {
		return false
	}
	j, ok := new(big.Int).SetString(y, 10)
	if !ok {
		return false
	}
	return i.Cmp(j) > 0
}

// IsZero returns true if the value is zero
func IsZero(s string) bool {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return false
	}
	return i != nil && i.Sign() == 0
}

// IsNegative returns true if the value is negative
func IsNegative(s string) bool {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return false
	}
	return i != nil && i.Sign() < 0
}

// IsPositive returns true if the value is positive
func IsPositive(s string) bool {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return false
	}
	return i != nil && i.Sign() > 0
}
