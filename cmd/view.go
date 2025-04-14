/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"

	crerr "github.com/cockroachdb/errors"
	"github.com/naosuke884/dirnote/note"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var viewCmd = &cobra.Command{
	Use:     "view",
	Short:   "View a note in the current directory.",
	Aliases: []string{"v"},
	RunE:    viewCmdRun,
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

func viewCmdRun(cmd *cobra.Command, args []string) error {
	directory, err := os.Getwd()
	if err != nil {
		return crerr.Wrap(err, "failed to get the current directory")
	}

	noteBucket := []byte("Notes")
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(noteBucket)
		var note note.Note
		if err := gob.NewDecoder(bytes.NewReader(bucket.Get([]byte(directory)))).Decode(&note); err != nil {
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
