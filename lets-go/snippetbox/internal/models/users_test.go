package models

import (
	"testing"

	"local.dev/snippetbox/internal/assert"
)

func TestUserModelExists(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	tests := map[string]struct {
		userID int
		want   bool
	}{
		"Valid ID returns true": {
			userID: 1,
			want:   true,
		},
		"Zero ID returns false": {
			userID: 0,
			want:   false,
		},
		"Non-existent ID returns false": {
			userID: 2,
			want:   false,
		},
	}

	db := newTestDBContainer(t)
	userModel := UserModel{DB: db}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exists, err := userModel.Exists(tt.userID)
			assert.Equal(t, tt.want, exists)
			assert.Nil(t, err)
		})
	}
}
