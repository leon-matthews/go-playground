// Package mocks provides mocks implementation of models
package mocks

import (
	"time"

	"local.dev/snippetbox/internal/models"
)

var mockSnippet = models.Snippet{
	ID:      1,
	Title:   "An old silent pond",
	Content: "An old silent pond...",
	Created: time.Now(),
	Expires: time.Now(),
}

// SnippetModel mocks database access to snippets table
type SnippetModel struct{}

// Insert always 'succeeds', returning ID of 2
func (m *SnippetModel) Insert(title, content string, expires int) (int, error) {
	return 2, nil
}

// Get returns the hard-coded mockSnippet for ID of 1, otherwise fails with [models.ErrNoRecord]
func (m *SnippetModel) Get(id int) (models.Snippet, error) {
	switch id {
	case 1:
		return mockSnippet, nil
	default:
		return models.Snippet{}, models.ErrNoRecord
	}
}

// Latest always returns a slice of length one with just the hard-coded mockSnippet
func (m *SnippetModel) Latest() ([]models.Snippet, error) {
	return []models.Snippet{mockSnippet}, nil
}
