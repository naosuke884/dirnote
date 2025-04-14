/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/gob"
	"os"

	crerr "github.com/cockroachdb/errors"
	"github.com/naosuke884/dirnote/note"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
)

var addCmd = &cobra.Command{
	Use:     "add [note]",
	Short:   "Add a note to the current directory.",
	Aliases: []string{"a"},
	Args:    cobra.ExactArgs(1),
	RunE:    addCmdRun,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func addCmdRun(cmd *cobra.Command, args []string) error {
	noteBucket := []byte("Notes")
	err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(noteBucket)
		return crerr.Wrap(err, "failed to create bucket")
	})
	if err != nil {
		return crerr.Wrap(err, "failed to create bucket")
	}

	cwd, _ := os.Getwd()
	directory, err := note.NewDirectory(cwd)
	if err != nil {
		return crerr.Wrap(err, "failed to get the current directory")
	}

	body := args[0]

	err = db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(noteBucket)
		old_encoded_note := bucket.Get([]byte(directory.Path))
		var new_note note.Note
		if old_encoded_note == nil {
			new_note = note.Note{
				Directory: directory,
				Body:      body,
			}
		} else {
			var old_note note.Note
			if err := gob.NewDecoder(bytes.NewReader(old_encoded_note)).Decode(&old_note); err != nil {
				return crerr.Wrap(err, "failed to decode the note")
			}
			new_note = note.Note{
				Directory: directory,
				Body:      old_note.Body + "\n" + body,
			}
		}

		var encoded_note bytes.Buffer
		if err := gob.NewEncoder(&encoded_note).Encode(new_note); err != nil {
			return crerr.Wrap(err, "failed to encode the note")
		}
		return bucket.Put([]byte(directory.Path), encoded_note.Bytes())
	})
	if err != nil {
		return crerr.Wrap(err, "failed to put the note")
	}
	err = viewCmdRun(cmd, args)
	if err != nil {
		return crerr.Wrap(err, "failed to view the note")
	}
	return nil
}
