package mocks

import (
	"snippetbox.i4o.dev/internal/models"
	"time"
)

var mockSnippet = models.Snippet{
	ID:      1,
	Title:   "Echoes",
	Content: "Overhead the albatross, hangs motionless upon the air",
	Created: time.Now(),
	Expires: time.Now(),
}

type SnippetModel struct {
}

func (model *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	return 2, nil
}

func (model *SnippetModel) Get(id int) (models.Snippet, error) {
	switch id {
	case 1:
		return mockSnippet, nil
	default:
		return models.Snippet{}, models.ErrNoRecord
	}
}

func (model *SnippetModel) Latest() ([]models.Snippet, error) {
	return []models.Snippet{mockSnippet}, nil
}
