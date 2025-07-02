//go:build unit

package responses

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressReporter_Processing(t *testing.T) {
	progressChan := make(chan Message, 1)
	reporter := NewProgressReporter(progressChan)
	testMessage := "Step 1: Validating data..."

	reporter.Processing(testMessage)

	select {
	case msg := <-progressChan:
		assert.Equal(t, "processing", msg.Status)
		assert.Equal(t, testMessage, msg.Message)
		assert.Nil(t, msg.Data)
		assert.Empty(t, msg.Error)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for message on channel")
	}
}

func TestProgressReporter_Success(t *testing.T) {
	progressChan := make(chan Message, 1)
	reporter := NewProgressReporter(progressChan)
	testMessage := "Operation completed"
	testData := struct {
		UserID int `json:"user_id"`
	}{UserID: 123}

	reporter.Success(testMessage, testData)

	select {
	case msg := <-progressChan:
		assert.Equal(t, "success", msg.Status)
		assert.Equal(t, testMessage, msg.Message)
		assert.Equal(t, testData, msg.Data)
		assert.Empty(t, msg.Error)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for message on channel")
	}
}

func TestProgressReporter_Error(t *testing.T) {
	t.Run("With a non-nil error", func(t *testing.T) {
		progressChan := make(chan Message, 1)
		reporter := NewProgressReporter(progressChan)
		testMessage := "Failed to connect to database"
		expectedError := errors.New("connection refused")

		reporter.Error(testMessage, expectedError)

		select {
		case msg := <-progressChan:
			assert.Equal(t, "error", msg.Status)
			assert.Equal(t, testMessage, msg.Message)
			assert.Equal(t, expectedError.Error(), msg.Error)
			assert.Nil(t, msg.Data)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for message on channel")
		}
	})

	t.Run("With a nil error", func(t *testing.T) {
		progressChan := make(chan Message, 1)
		reporter := NewProgressReporter(progressChan)
		testMessage := "An unknown error occurred"

		reporter.Error(testMessage, nil)

		select {
		case msg := <-progressChan:
			assert.Equal(t, "error", msg.Status)
			assert.Equal(t, testMessage, msg.Message)
			assert.Empty(t, msg.Error)
			assert.Nil(t, msg.Data)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for message on channel")
		}
	})
}

func TestNewProgressReporter(t *testing.T) {
	progressChan := make(chan Message, 1)

	reporter := NewProgressReporter(progressChan)
	require.NotNil(t, reporter)
	assert.NotNil(t, reporter.channel)
}
