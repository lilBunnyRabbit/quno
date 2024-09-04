package notes

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const NotesPath = "./tmp"

// Maybe?
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func CreateNote(title, content string) (string, error) {
	date := strconv.FormatInt(time.Now().Unix(), 10)

	fileName := fmt.Sprintf("%s/%s#%s.txt", NotesPath, date, title)

	file, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", fileName, err)
	}
	defer file.Close()

	text := fmt.Sprintf("[Title]\n%s\n\n[Content]\n%s", title, content)

	_, err = file.WriteString(text)
	if err != nil {
		return "", fmt.Errorf("failed to write to file %s: %w", fileName, err)
	}

	return fileName, nil
}

func GetNotes() ([]PartialNote, error) {
	entries, err := os.ReadDir(NotesPath)
	if err != nil {
		return []PartialNote{}, fmt.Errorf("failed to read directory %s: %w", NotesPath, err)
	}

	var notes []PartialNote

	for _, e := range entries {
		name := e.Name()

		splits := strings.SplitN(name, "#", 2)

		if len(splits) < 2 {
			fmt.Printf("WARNING: Skipping file %s due to invalid name", name)
			continue
		}

		unix, err := strconv.ParseInt(splits[0], 10, 64)
		if err != nil {
			fmt.Printf("WARNING: Skipping file %s due to invalid timestamp", name)
			continue
		}

		unixTime := time.Unix(unix, 0)

		notes = append(notes, PartialNote{
			Title:     strings.Replace(splits[1], ".txt", "", 1),
			CreatedAt: unixTime,
			Path:      fmt.Sprintf("%s/%s", NotesPath, name),
		})
	}

	return notes, nil
}

func GetNote(partial PartialNote) Note {
	content, err := os.ReadFile(partial.Path)
	check(err)

	return NoteFromPartial(partial, string(content))
}
