/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	crerr "github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all notes.",
	Aliases: []string{"l"},
	RunE:    runListCmd,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runListCmd(cmd *cobra.Command, args []string) error {
	directory, err := os.Getwd()
	if err != nil {
		return crerr.Wrap(err, "failed to get the current directory")
	}
	home_directory, err := os.UserHomeDir()
	if err != nil {
		return crerr.Wrap(err, "failed to get the home directory")
	}

	noteBucket := []byte("Notes")
	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(noteBucket)
		bucket.ForEach(func(k, v []byte) error {
			simbol := " "
			if string(k) == directory {
				simbol = "*"
			}
			path := string(k)
			if len(path) >= len(home_directory) && path[:len(home_directory)] == home_directory {
				path = "~" + path[len(home_directory):]
			}
			fmt.Printf("\033[32m%s\033[0m %s\n", simbol, path)
			return nil
		})
		return nil
	})
	if err != nil {
		return crerr.Wrap(err, "failed to list notes")
	}
	return nil
}
