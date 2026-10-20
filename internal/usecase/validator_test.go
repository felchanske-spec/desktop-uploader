package usecase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"desktop-uploader/internal/domain"

	"github.com/stretchr/testify/assert"
)

type mockTaskQueue struct {
	savedTask domain.Task
	pushErr   error
}

func (m *mockTaskQueue) Push(task domain.Task) error {
	m.savedTask = task
	return m.pushErr
}

func TestUploadInteractor_ValidateAndQueue_TableDriven(t *testing.T) {
	tmpDir := t.TempDir()

	goodFilePath := filepath.Join(tmpDir, "good_file.txt")
	err := os.WriteFile(goodFilePath, make([]byte, 100), 0644)
	assert.NoError(t, err)

	hugeFilePath := filepath.Join(tmpDir, "huge_file.txt")
	hugeFile, err := os.Create(hugeFilePath)
	assert.NoError(t, err)
	_, err = hugeFile.WriteAt([]byte{0}, 51*1024*1024)
	assert.NoError(t, err)
	hugeFile.Close()

	type testCase struct {
		name          string
		filePath      string
		target        string
		expectedErr   error
		expectToQueue bool
	}

	tests := []testCase{
		{
			name:          "Success scenario with valid file",
			filePath:      goodFilePath,
			target:        "123456",
			expectedErr:   nil,
			expectToQueue: true,
		},
		{
			name:          "Error when file does not exist",
			filePath:      filepath.Join(tmpDir, "ghost_file.txt"),
			target:        "123456",
			expectedErr:   ErrFileNotFound,
			expectToQueue: false,
		},
		{
			name:          "Error when path points to a directory",
			filePath:      tmpDir,
			target:        "123456",
			expectedErr:   ErrCannotSendDirectory,
			expectToQueue: false,
		},
		{
			name:          "Error when file exceeds 50MB limit",
			filePath:      hugeFilePath,
			target:        "123456",
			expectedErr:   ErrFileTooLarge,
			expectToQueue: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockQueue := &mockTaskQueue{}
			interactor := NewUploadInteractor(mockQueue)

			startTime := time.Now()
			err := interactor.ValidateAndQueue(tc.filePath, tc.target)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.expectedErr)
			}

			if tc.expectToQueue {
				assert.Equal(t, tc.filePath, mockQueue.savedTask.FilePath)
				assert.Equal(t, tc.target, mockQueue.savedTask.Target)
				assert.NotEmpty(t, mockQueue.savedTask.ID)
				assert.True(t, mockQueue.savedTask.CreatedAt.After(startTime) || mockQueue.savedTask.CreatedAt.Equal(startTime))
			} else {
				assert.Empty(t, mockQueue.savedTask)
			}
		})
	}
}
