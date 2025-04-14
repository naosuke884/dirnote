/*
Copyright © 2025 Naoki Hayashi 884naoki.general@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

var configFile string
var storageDir string
var log zerolog.Logger
var db *bolt.DB

var rootCmd = &cobra.Command{
	Use:                "dirnote",
	Short:              "A simple note management CLI application. Notes are managed by directories.",
	PersistentPreRunE:  persistentPreRunE,
	PersistentPostRunE: persistentPostRunE,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Str("stacktrace", fmt.Sprintf("%+v", err)).Msg("failed to execute the command")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger, initStorageDir, initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.dirnote.yaml)")
}

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func initStorageDir() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get the home directory")
	}
	storageDir = filepath.Join(home, ".dirnote")
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		err := os.Mkdir(storageDir, 0700)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create the configuration directory")
		}
	}
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(storageDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".dirnote")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Info().Msgf("Using config file: %s", viper.ConfigFileUsed())
	}
}

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	dbMiddleware := &dbMiddleware{}
	if err := dbMiddleware.PreRunE(cmd, args); err != nil {
		log.Error().Err(err).Msg("failed to open the database")
		return err

	}
	return nil
}

func persistentPostRunE(cmd *cobra.Command, args []string) error {
	dbMiddleware := &dbMiddleware{}
	if err := dbMiddleware.PostRunE(cmd, args); err != nil {
		log.Error().Err(err).Msg("failed to close the database")
		return err
	}
	return nil
}

type Middleware interface {
	PreRunE(cmd *cobra.Command, args []string) error
	PostRunE(cmd *cobra.Command, args []string) error
}

type dbMiddleware struct{}

func (m *dbMiddleware) PreRunE(cmd *cobra.Command, args []string) error {
	var err error
	db, err = bolt.Open(filepath.Join(storageDir, "dirnote.db"), 0600, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to open the database")
		return err
	}
	return nil
}

func (m *dbMiddleware) PostRunE(cmd *cobra.Command, args []string) error {
	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close the database")
		return err
	}
	return nil
}
