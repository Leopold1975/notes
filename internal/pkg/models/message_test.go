package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"notes/internal/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
	assert.NoError(t, err)
	n := models.Note{
		ID:          1,
		Title:       "test",
		Description: "test",
		DateAdded:   tm,
		DateNotify:  tm,
		Delay:       time.Minute * 20,
	}

	m, err := models.NoteToMessage(n)
	assert.NoError(t, err)

	var id uint64
	err = json.Unmarshal(m.Key, &id)
	assert.NoError(t, err)
	assert.Equal(t, id, n.ID)

	var testNote models.Note
	err = json.Unmarshal(m.Value, &testNote)
	assert.NoError(t, err)
	assert.Equal(t, n, testNote)
}
