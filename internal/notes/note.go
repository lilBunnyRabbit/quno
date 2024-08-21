package notes

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Note struct {
	Title     string
	Content   string
	CreatedAt time.Time
	Path      string
}

func (note *Note) ToString() string {
	return fmt.Sprintf(
		"%s\n%s\n%s",
		color.CyanString(note.Title),
		strings.Repeat("-", len(note.Title)),
		note.Content,
	)
}

type PartialNote struct {
	Title     string
	CreatedAt time.Time
	Path      string
}

func NoteFromPartial(partial PartialNote, content string) Note {
	return Note{
		Title:     partial.Title,
		Content:   content,
		CreatedAt: partial.CreatedAt,
		Path:      partial.Path,
	}
}
