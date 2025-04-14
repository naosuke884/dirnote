/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	crerr "github.com/cockroachdb/errors"
	"github.com/naosuke884/dirnote/note"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Edit a note in the current directory.",
	Aliases: []string{"e"},
	RunE:    editCmdRun,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func editCmdRun(cmd *cobra.Command, args []string) error {
	noteBucket := []byte("Notes")

	cwd, _ := os.Getwd()
	directory, err := note.NewDirectory(cwd)
	if err != nil {
		return crerr.Wrap(err, "failed to get the current directory")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(noteBucket)
		if bucket == nil {
			return crerr.New("bucket not found")
		}

		old_encoded_note := bucket.Get([]byte(directory.Path))
		if old_encoded_note == nil {
			return crerr.New("note not found")
		}

		var old_note note.Note
		if err := gob.NewDecoder(bytes.NewReader(old_encoded_note)).Decode(&old_note); err != nil {
			return crerr.Wrap(err, "failed to decode the note")
		}

		new_body, err := editWithVim(old_note.Body)
		if err != nil {
			return crerr.Wrap(err, "failed to edit the note")
		}

		new_note := note.Note{
			Directory: directory,
			Body:      new_body,
		}

		var encoded_note bytes.Buffer
		if err := gob.NewEncoder(&encoded_note).Encode(new_note); err != nil {
			return crerr.Wrap(err, "failed to encode the note")
		}
		err = bucket.Put([]byte(directory.Path), encoded_note.Bytes())
		if err != nil {
			return crerr.Wrap(err, "failed to put the note")
		}
		return nil
	})
	if err != nil {
		return crerr.Wrap(err, "failed to update the note")
	}
	err = viewCmdRun(cmd, args)
	if err != nil {
		return crerr.Wrap(err, "failed to view the note")
	}
	return nil
}

func editWithVim(input string) (string, error) {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("vim_edit_%d.txt", time.Now().UnixNano()))

	if err := os.WriteFile(tmpFile, []byte(input), 0644); err != nil {
		return "", crerr.Wrap(err, "failed to write to temporary file")
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("vim", "-c", "set encoding=utf-8 | set fileencoding=utf-8 | set fileencodings=utf-8 | set nobomb", tmpFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", crerr.Wrap(err, "failed to run vim")
	}

	edited, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", crerr.Wrap(err, "failed to read the temporary file")
	}

	return strings.TrimRight(string(edited), "\r\n"), nil
}
