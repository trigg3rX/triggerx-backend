package types

import (
	"encoding/json"
	"math/big"
	"reflect"
	"strconv"
)

// BigInt is a wrapper around *big.Int that provides custom JSON marshaling/unmarshaling
type BigInt struct {
	*big.Int
}

// NewBigInt creates a new BigInt from a *big.Int
func NewBigInt(i *big.Int) *BigInt {
	if i == nil {
		return nil
	}
	return &BigInt{Int: i}
}

// MarshalJSON implements json.Marshaler interface
func (b *BigInt) MarshalJSON() ([]byte, error) {
	if b == nil || b.Int == nil {
		return []byte("null"), nil
	}
	// Marshal as string to avoid scientific notation
	return json.Marshal(b.Int.String())
}

// UnmarshalJSON implements json.Unmarshaler interface
func (b *BigInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		b.Int = nil
		return nil
	}

	// Only accept string format for 256-bit integers
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return &json.UnmarshalTypeError{
			Value:  string(data),
			Type:   reflect.TypeOf(""),
			Offset: 0,
			Struct: "BigInt",
			Field:  "Int",
		}
	}

	// Parse string as big.Int
	i, ok := new(big.Int).SetString(str, 10)
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  "string",
			Type:   reflect.TypeOf(big.Int{}),
			Offset: 0,
			Struct: "BigInt",
			Field:  "Int",
		}
	}
	b.Int = i
	return nil
}

// ToBigInt converts BigInt to *big.Int
func (b *BigInt) ToBigInt() *big.Int {
	if b == nil {
		return nil
	}
	return b.Int
}

// FromBigInt creates a BigInt from *big.Int
func FromBigInt(i *big.Int) *BigInt {
	return NewBigInt(i)
}

// String implements fmt.Stringer interface
func (b *BigInt) String() string {
	if b == nil || b.Int == nil {
		return "<nil>"
	}
	return b.Int.String()
}

// SetString sets the value from a string
func (b *BigInt) SetString(s string, base int) (*BigInt, bool) {
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	i, ok := b.Int.SetString(s, base)
	if !ok {
		return nil, false
	}
	b.Int = i
	return b, true
}

// Add adds x and y and stores the result in b
func (b *BigInt) Add(x, y *BigInt) *BigInt {
	if b == nil {
		b = &BigInt{Int: new(big.Int)}
	}
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	if x == nil || x.Int == nil {
		x = &BigInt{Int: big.NewInt(0)}
	}
	if y == nil || y.Int == nil {
		y = &BigInt{Int: big.NewInt(0)}
	}
	b.Int.Add(x.Int, y.Int)
	return b
}

// Sub subtracts y from x and stores the result in b
func (b *BigInt) Sub(x, y *BigInt) *BigInt {
	if b == nil {
		b = &BigInt{Int: new(big.Int)}
	}
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	if x == nil || x.Int == nil {
		x = &BigInt{Int: big.NewInt(0)}
	}
	if y == nil || y.Int == nil {
		y = &BigInt{Int: big.NewInt(0)}
	}
	b.Int.Sub(x.Int, y.Int)
	return b
}

// Mul multiplies x and y and stores the result in b
func (b *BigInt) Mul(x, y *BigInt) *BigInt {
	if b == nil {
		b = &BigInt{Int: new(big.Int)}
	}
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	if x == nil || x.Int == nil {
		x = &BigInt{Int: big.NewInt(0)}
	}
	if y == nil || y.Int == nil {
		y = &BigInt{Int: big.NewInt(0)}
	}
	b.Int.Mul(x.Int, y.Int)
	return b
}

// Div divides x by y and stores the result in b
func (b *BigInt) Div(x, y *BigInt) *BigInt {
	if b == nil {
		b = &BigInt{Int: new(big.Int)}
	}
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	if x == nil || x.Int == nil {
		x = &BigInt{Int: big.NewInt(0)}
	}
	if y == nil || y.Int == nil {
		y = &BigInt{Int: big.NewInt(1)}
	}
	b.Int.Div(x.Int, y.Int)
	return b
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (b *BigInt) Cmp(x *BigInt) int {
	if b == nil || b.Int == nil {
		b = &BigInt{Int: big.NewInt(0)}
	}
	if x == nil || x.Int == nil {
		x = &BigInt{Int: big.NewInt(0)}
	}
	return b.Int.Cmp(x.Int)
}

// Equal returns true if x equals y
func (b *BigInt) Equal(x *BigInt) bool {
	return b.Cmp(x) == 0
}

// Less returns true if x is less than y
func (b *BigInt) Less(x *BigInt) bool {
	return b.Cmp(x) < 0
}

// Greater returns true if x is greater than y
func (b *BigInt) Greater(x *BigInt) bool {
	return b.Cmp(x) > 0
}

// IsZero returns true if the value is zero
func (b *BigInt) IsZero() bool {
	return b == nil || b.Int == nil || b.Sign() == 0
}

// IsNegative returns true if the value is negative
func (b *BigInt) IsNegative() bool {
	return b != nil && b.Int != nil && b.Sign() < 0
}

// IsPositive returns true if the value is positive
func (b *BigInt) IsPositive() bool {
	return b != nil && b.Int != nil && b.Sign() > 0
}

// Clone creates a copy of the BigInt
func (b *BigInt) Clone() *BigInt {
	if b == nil || b.Int == nil {
		return nil
	}
	return &BigInt{Int: new(big.Int).Set(b.Int)}
}

// ParseBigInt parses a string as a BigInt
func ParseBigInt(s string) (*BigInt, error) {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, &strconv.NumError{
			Func: "ParseBigInt",
			Num:  s,
			Err:  strconv.ErrSyntax,
		}
	}
	return &BigInt{Int: i}, nil
}

// MustParseBigInt parses a string as a BigInt, panicking on error
func MustParseBigInt(s string) *BigInt {
	b, err := ParseBigInt(s)
	if err != nil {
		panic(err)
	}
	return b
}
