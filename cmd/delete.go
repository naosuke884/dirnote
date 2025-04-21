/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	crerr "github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a note in the current directory.",
	Aliases: []string{"d"},
	RunE:    deleteCmdRun,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func deleteCmdRun(cmd *cobra.Command, args []string) error {
	directory, err := os.Getwd()
	if err != nil {
		return crerr.Wrap(err, "failed to get the current directory")
	}

	fmt.Print("Are you sure you want to delete the note? (y/N): ")
	var response string
	_, err = fmt.Scanln(&response)
	if err != nil {
		return crerr.Wrap(err, "failed to read user input")
	}
	if response != "y" && response != "Y" {
		fmt.Println("Operation canceled.")
		return nil
	}

	noteBucket := []byte("Notes")
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(noteBucket)
		if bucket == nil {
			return crerr.New("bucket does not exist")
		}
		if bucket.Get([]byte(directory)) == nil {
			return crerr.New("no note found for the current directory")
		}
		if err := bucket.Delete([]byte(directory)); err != nil {
			return crerr.Wrap(err, "failed to delete the note")
		}
		return nil
	})
	if err != nil {
		return crerr.Wrap(err, "failed to delete the note")
	}
	fmt.Println("Note deleted successfully.")
	return nil
}
