/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	crerr "github.com/cockroachdb/errors"
	"github.com/naosuke884/dirnote/note"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var interactive bool

var viewCmd = &cobra.Command{
	Use:     "view",
	Short:   "View a note in the current directory.",
	Aliases: []string{"v"},
	RunE:    viewCmdRun,
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactively select a directory to view its note")
}

func viewCmdRun(cmd *cobra.Command, args []string) error {
	var directory string
	var err error

	if interactive {
		directory, err = selectDirectory()
		if err == NoSelectDirectoryError {
			fmt.Println("No directory selected.")
			return nil
		}
		if err != nil {
			return crerr.Wrap(err, "failed to select a directory")
		}
	} else {
		directory, err = os.Getwd()
		if err != nil {
			return crerr.Wrap(err, "failed to get the current directory")
		}
	}

	noteBucket := []byte("Notes")
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(noteBucket)
		note_byte := bucket.Get([]byte(directory))
		if note_byte == nil {
			fmt.Println("No note found for the current directory.")
			return nil
		}
		var note note.Note
		if err := gob.NewDecoder(bytes.NewReader(note_byte)).Decode(&note); err != nil {
			return crerr.Wrap(err, "failed to decode the note")
		}
		fmt.Println(note.Body)
		return nil
	})
	if err != nil {
		return crerr.Wrap(err, "failed to view the note")
	}
	return nil
}

var NoSelectDirectoryError = crerr.New("no directory selected")

func selectDirectory() (string, error) {
	var directories []string
	noteBucket := []byte("Notes")

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(noteBucket)
		return bucket.ForEach(func(k, _ []byte) error {
			directories = append(directories, string(k))
			return nil
		})
	})
	if err != nil {
		return "", crerr.Wrap(err, "failed to fetch directories")
	}

	if len(directories) == 0 {
		return "", crerr.New("no notes available to select")
	}

	model := directorySelectorModel{directories: directories}
	p := tea.NewProgram(model, tea.WithAltScreen()) // Enable alternate screen
	result, err := p.Run()
	if err != nil {
		return "", crerr.Wrap(err, "failed to run interactive selection")
	}

	selectedModel, ok := result.(directorySelectorModel)
	if !ok || selectedModel.selected == "" {
		return "", NoSelectDirectoryError
	}

	return selectedModel.selected, nil
}

type directorySelectorModel struct {
	directories []string
	cursor      int
	selected    string
}

func (m directorySelectorModel) Init() tea.Cmd {
	return nil
}

func (m directorySelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.directories)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.directories[m.cursor]
			return m, tea.Quit
		case "esc", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m directorySelectorModel) View() string {
	s := "Select a directory:\n\n"
	for i, dir := range m.directories {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor
		}
		s += fmt.Sprintf("%s %s\n", cursor, dir)
	}
	return s
}
