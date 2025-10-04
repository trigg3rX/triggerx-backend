package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMockGocqlxQueryer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGocqlxQueryer := NewMockGocqlxQueryer(ctrl)

	t.Run("BindStruct", func(t *testing.T) {
		data := struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{
			ID:   1,
			Name: "test",
		}
		expectedQueryer := NewMockGocqlxQueryer(ctrl)

		mockGocqlxQueryer.EXPECT().BindStruct(data).Return(expectedQueryer)

		result := mockGocqlxQueryer.BindStruct(data)
		assert.Equal(t, expectedQueryer, result)
	})

	t.Run("ExecRelease", func(t *testing.T) {
		expectedError := errors.New("exec release failed")

		mockGocqlxQueryer.EXPECT().ExecRelease().Return(expectedError)

		result := mockGocqlxQueryer.ExecRelease()
		assert.Equal(t, expectedError, result)
	})

	t.Run("ExecRelease_Success", func(t *testing.T) {
		mockGocqlxQueryer.EXPECT().ExecRelease().Return(nil)

		result := mockGocqlxQueryer.ExecRelease()
		assert.NoError(t, result)
	})

	t.Run("GetRelease", func(t *testing.T) {
		dest := struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}
		expectedError := errors.New("get release failed")

		mockGocqlxQueryer.EXPECT().GetRelease(&dest).Return(expectedError)

		result := mockGocqlxQueryer.GetRelease(&dest)
		assert.Equal(t, expectedError, result)
	})

	t.Run("GetRelease_Success", func(t *testing.T) {
		dest := struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}

		mockGocqlxQueryer.EXPECT().GetRelease(&dest).Return(nil)

		result := mockGocqlxQueryer.GetRelease(&dest)
		assert.NoError(t, result)
	})

	t.Run("Select", func(t *testing.T) {
		dest := []struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}
		expectedError := errors.New("select failed")

		mockGocqlxQueryer.EXPECT().Select(&dest).Return(expectedError)

		result := mockGocqlxQueryer.Select(&dest)
		assert.Equal(t, expectedError, result)
	})

	t.Run("Select_Success", func(t *testing.T) {
		dest := []struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}

		mockGocqlxQueryer.EXPECT().Select(&dest).Return(nil)

		result := mockGocqlxQueryer.Select(&dest)
		assert.NoError(t, result)
	})

	t.Run("SelectRelease", func(t *testing.T) {
		dest := []struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}
		expectedError := errors.New("select release failed")

		mockGocqlxQueryer.EXPECT().SelectRelease(&dest).Return(expectedError)

		result := mockGocqlxQueryer.SelectRelease(&dest)
		assert.Equal(t, expectedError, result)
	})

	t.Run("SelectRelease_Success", func(t *testing.T) {
		dest := []struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}

		mockGocqlxQueryer.EXPECT().SelectRelease(&dest).Return(nil)

		result := mockGocqlxQueryer.SelectRelease(&dest)
		assert.NoError(t, result)
	})

	t.Run("WithContext", func(t *testing.T) {
		ctx := context.Background()
		expectedQueryer := NewMockGocqlxQueryer(ctrl)

		mockGocqlxQueryer.EXPECT().WithContext(ctx).Return(expectedQueryer)

		result := mockGocqlxQueryer.WithContext(ctx)
		assert.Equal(t, expectedQueryer, result)
	})

	t.Run("WithContext_Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()
		expectedQueryer := NewMockGocqlxQueryer(ctrl)

		mockGocqlxQueryer.EXPECT().WithContext(ctx).Return(expectedQueryer)

		result := mockGocqlxQueryer.WithContext(ctx)
		assert.Equal(t, expectedQueryer, result)
	})
}

// TestMockGocqlxQueryerChaining tests method chaining
func TestMockGocqlxQueryerChaining(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueryer := NewMockGocqlxQueryer(ctrl)
	mockQueryerWithContext := NewMockGocqlxQueryer(ctrl)
	mockQueryerWithBind := NewMockGocqlxQueryer(ctrl)

	t.Run("WithContextAndBindStruct", func(t *testing.T) {
		ctx := context.Background()
		data := struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{
			ID:   1,
			Name: "test",
		}

		// Setup expectations for chaining
		mockQueryer.EXPECT().WithContext(ctx).Return(mockQueryerWithContext)
		mockQueryerWithContext.EXPECT().BindStruct(data).Return(mockQueryerWithBind)
		mockQueryerWithBind.EXPECT().ExecRelease().Return(nil)

		// Execute chained calls
		result := mockQueryer.WithContext(ctx).BindStruct(data)
		err := result.ExecRelease()

		// Verify
		assert.Equal(t, mockQueryerWithBind, result)
		assert.NoError(t, err)
	})
}

// TestMockGocqlxQueryerSelectOperations tests various select operations
func TestMockGocqlxQueryerSelectOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueryer := NewMockGocqlxQueryer(ctrl)

	t.Run("SelectMultipleRecords", func(t *testing.T) {
		dest := []struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}

		mockQueryer.EXPECT().Select(&dest).Return(nil)

		result := mockQueryer.Select(&dest)
		assert.NoError(t, result)
	})

	t.Run("SelectReleaseMultipleRecords", func(t *testing.T) {
		dest := []struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}

		mockQueryer.EXPECT().SelectRelease(&dest).Return(nil)

		result := mockQueryer.SelectRelease(&dest)
		assert.NoError(t, result)
	})
}

// TestMockGocqlxQueryerGetOperations tests various get operations
func TestMockGocqlxQueryerGetOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueryer := NewMockGocqlxQueryer(ctrl)

	t.Run("GetReleaseSingleRecord", func(t *testing.T) {
		dest := struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}

		mockQueryer.EXPECT().GetRelease(&dest).Return(nil)

		result := mockQueryer.GetRelease(&dest)
		assert.NoError(t, result)
	})

	t.Run("GetReleaseNotFound", func(t *testing.T) {
		dest := struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}
		expectedError := errors.New("record not found")

		mockQueryer.EXPECT().GetRelease(&dest).Return(expectedError)

		result := mockQueryer.GetRelease(&dest)
		assert.Equal(t, expectedError, result)
	})
}

// TestMockGocqlxQueryerErrorHandling tests error scenarios
func TestMockGocqlxQueryerErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueryer := NewMockGocqlxQueryer(ctrl)

	t.Run("ExecReleaseError", func(t *testing.T) {
		expectedError := errors.New("connection timeout")

		mockQueryer.EXPECT().ExecRelease().Return(expectedError)

		result := mockQueryer.ExecRelease()
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "connection timeout")
	})

	t.Run("SelectError", func(t *testing.T) {
		dest := []struct {
			ID   int    `cql:"id"`
			Name string `cql:"name"`
		}{}
		expectedError := errors.New("invalid query")

		mockQueryer.EXPECT().Select(&dest).Return(expectedError)

		result := mockQueryer.Select(&dest)
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "invalid query")
	})
}

// TestMockGocqlxQueryerComplexDataStructures tests with complex data structures
func TestMockGocqlxQueryerComplexDataStructures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueryer := NewMockGocqlxQueryer(ctrl)

	t.Run("BindStructWithComplexData", func(t *testing.T) {
		type ComplexStruct struct {
			ID       int                    `cql:"id"`
			Name     string                 `cql:"name"`
			Tags     []string               `cql:"tags"`
			Metadata map[string]interface{} `cql:"metadata"`
			IsActive bool                   `cql:"is_active"`
			Score    float64                `cql:"score"`
		}

		data := ComplexStruct{
			ID:       1,
			Name:     "complex_test",
			Tags:     []string{"tag1", "tag2"},
			Metadata: map[string]interface{}{"key": "value"},
			IsActive: true,
			Score:    95.5,
		}
		expectedQueryer := NewMockGocqlxQueryer(ctrl)

		mockQueryer.EXPECT().BindStruct(data).Return(expectedQueryer)

		result := mockQueryer.BindStruct(data)
		assert.Equal(t, expectedQueryer, result)
	})
}

// Benchmark tests for performance validation
func BenchmarkMockGocqlxQueryerBindStruct(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockQueryer := NewMockGocqlxQueryer(ctrl)
	expectedQueryer := NewMockGocqlxQueryer(ctrl)

	data := struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}{ID: 1, Name: "test"}

	mockQueryer.EXPECT().BindStruct(data).Return(expectedQueryer).AnyTimes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockQueryer.BindStruct(data)
	}
}

func BenchmarkMockGocqlxQueryerExecRelease(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockQueryer := NewMockGocqlxQueryer(ctrl)

	mockQueryer.EXPECT().ExecRelease().Return(nil).AnyTimes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockQueryer.ExecRelease()
	}
}

func BenchmarkMockGocqlxQueryerSelect(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockQueryer := NewMockGocqlxQueryer(ctrl)

	dest := []struct {
		ID   int    `cql:"id"`
		Name string `cql:"name"`
	}{}

	mockQueryer.EXPECT().Select(&dest).Return(nil).AnyTimes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockQueryer.Select(&dest)
	}
}
