package types

import (
	"math/big"
	"strconv"
)

// parseBigIntString is an internal helper function that parses a string to big.Int.
// Returns an error if the string is not a valid integer.
func parseBigIntString(s string) (*big.Int, error) {
	if s == "" {
		return big.NewInt(0), nil
	}
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, &strconv.NumError{
			Func: "parseBigIntString",
			Num:  s,
			Err:  strconv.ErrSyntax,
		}
	}
	return i, nil
}

// Add adds two Wei-based string values and returns the result as a string.
func Add(x, y string) string {
	xBig, _ := parseBigIntString(x)
	yBig, _ := parseBigIntString(y)
	result := new(big.Int).Add(xBig, yBig)
	return result.String()
}

// Sub subtracts y from x (both Wei-based strings) and returns the result as a string.
func Sub(x, y string) string {
	xBig, _ := parseBigIntString(x)
	yBig, _ := parseBigIntString(y)
	result := new(big.Int).Sub(xBig, yBig)
	return result.String()
}

// Mul multiplies two Wei-based string values and returns the result as a string.
func Mul(x, y string) string {
	xBig, _ := parseBigIntString(x)
	yBig, _ := parseBigIntString(y)
	result := new(big.Int).Mul(xBig, yBig)
	return result.String()
}

// Div divides x by y (both Wei-based strings) and returns the result as a string.
// Returns "0" if y is zero or invalid.
func Div(x, y string) string {
	xBig, _ := parseBigIntString(x)
	yBig, _ := parseBigIntString(y)
	if yBig.Sign() == 0 {
		return "0"
	}
	result := new(big.Int).Div(xBig, yBig)
	return result.String()
}

// IsZero returns true if the Wei-based string value is zero.
func IsZero(x string) bool {
	xBig, _ := parseBigIntString(x)
	return xBig.Sign() == 0
}

// IsEqual returns true if two Wei-based string values are equal.
func IsEqual(x, y string) bool {
	xBig, _ := parseBigIntString(x)
	yBig, _ := parseBigIntString(y)
	return xBig.Cmp(yBig) == 0
}

// IsLess returns true if x is less than y (both Wei-based strings).
func IsLess(x, y string) bool {
	xBig, _ := parseBigIntString(x)
	yBig, _ := parseBigIntString(y)
	return xBig.Cmp(yBig) < 0
}

// IsGreater returns true if x is greater than y (both Wei-based strings).
func IsGreater(x, y string) bool {
	xBig, _ := parseBigIntString(x)
	yBig, _ := parseBigIntString(y)
	return xBig.Cmp(yBig) > 0
}

// ConvertToBigInt converts a Wei-based string to *big.Int for use in blockchain operations.
// This function is intended for use with ethClient functions that require *big.Int.
// Returns nil if the string is invalid.
func ConvertToBigInt(s string) *big.Int {
	result, err := parseBigIntString(s)
	if err != nil {
		return nil
	}
	return result
}

// MulByFloat multiplies a Wei-based string by a float64 multiplier and returns the result as a string.
// This is useful for applying multipliers like rewardsBooster (e.g., 1.5) to Wei values.
func MulByFloat(weiStr string, multiplier float64) string {
	if weiStr == "" || multiplier == 0 {
		return "0"
	}
	weiBig, _ := parseBigIntString(weiStr)
	if weiBig == nil {
		return "0"
	}
	// Convert to big.Float for multiplication
	weiFloat := new(big.Float).SetInt(weiBig)
	multiplierFloat := big.NewFloat(multiplier)
	result := new(big.Float).Mul(weiFloat, multiplierFloat)
	// Convert back to big.Int (truncates decimal part, which is fine for Wei)
	resultInt, _ := result.Int(nil)
	return resultInt.String()
}
