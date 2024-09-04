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

func (note Note) String() string {
	dashes := strings.Repeat("-", max(len(note.Title), 32))

	return fmt.Sprintf(
		"\n%s\n%s\n%s\n\n%s",
		dashes,
		color.CyanString(note.Title),
		dashes,
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
