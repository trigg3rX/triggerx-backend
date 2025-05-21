package converter

import "math/big"

func ConvertBigIntToStrings(bigInts []*big.Int) []string {
	strings := make([]string, len(bigInts))
	for i, bigInt := range bigInts {
		strings[i] = bigInt.String()
	}
	return strings
}
