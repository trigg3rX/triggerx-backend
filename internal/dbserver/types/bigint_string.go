package types

import (
	"encoding/json"
	"math/big"
)

type BigIntString struct {
	*big.Int
}

func (b BigIntString) MarshalJSON() ([]byte, error) {
	if b.Int == nil {
		return []byte(`"0"`), nil
	}
	return json.Marshal(b.String())
}

func (b *BigIntString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	i := new(big.Int)
	i.SetString(s, 10)
	b.Int = i
	return nil
}
