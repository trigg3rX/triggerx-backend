package aggregator

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorVariables ensures the exported error variables are defined and behave as expected.
// This test is primarily a safeguard against accidental changes to these sentinel errors.
func TestErrorVariables(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectedText string
	}{
		{
			name:         "ErrInvalidKey",
			err:          ErrInvalidKey,
			expectedText: "invalid key",
		},
		{
			name:         "ErrSigningFailed",
			err:          ErrSigningFailed,
			expectedText: "signing operation failed",
		},
		{
			name:         "ErrRPCFailed",
			err:          ErrRPCFailed,
			expectedText: "RPC operation failed",
		},
		{
			name:         "ErrMarshalFailed",
			err:          ErrMarshalFailed,
			expectedText: "marshaling operation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check the error message.
			assert.EqualError(t, tc.err, tc.expectedText)

			// Check that the error can be wrapped and correctly identified.
			wrappedErr := fmt.Errorf("additional context: %w", tc.err)
			assert.True(t, errors.Is(wrappedErr, tc.err), "errors.Is should be able to identify the wrapped error")
		})
	}
}
