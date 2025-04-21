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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	crerr "github.com/cockroachdb/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cockroachdberrors "github.com/yumemi-inc/zerolog-cockroachdb-errors"
	bolt "go.etcd.io/bbolt"
)

var configFile string
var storageDir string
var isDebug bool
var log zerolog.Logger
var db *bolt.DB

var rootCmd = &cobra.Command{
	Use:                "dirnote",
	Short:              "A simple note management CLI application. Notes are managed by directories.",
	PersistentPreRunE:  persistentPreRunE,
	PersistentPostRunE: persistentPostRunE,
	SilenceErrors:      !isDebug,
	SilenceUsage:       !isDebug,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil && isDebug {
		log.Error().Err(err).Msg("failed to execute the command")
	}
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger, initStorageDir, initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.dirnote.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&isDebug, "debug", "d", false, "enable debug mode")
}

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = cockroachdberrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	consoleWriter := getConsoleWriter()
	log = zerolog.New(consoleWriter).
		With().
		Timestamp().
		Stack().
		Logger()
}

func getConsoleWriter() zerolog.ConsoleWriter {
	const (
		cFunc  = "\x1b[36m" // シアン：関数名
		cFile  = "\x1b[33m" // 黄色：ファイル名
		cLine  = "\x1b[35m" // マゼンタ：行番号
		cReset = "\x1b[0m"  // リセット
	)
	return zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: false,
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("%s:", i)
		},
		FormatFieldValue: func(v interface{}) string {
			// JSON 化されたスタックトレースをパース
			var arr []interface{}
			switch val := v.(type) {
			case []byte:
				if err := json.Unmarshal(val, &arr); err != nil {
					return string(val)
				}
			case []interface{}:
				arr = val
			default:
				return fmt.Sprint(v)
			}

			var b strings.Builder
			b.WriteRune('\n') // 先頭に空行

			for _, entry := range arr {
				m := entry.(map[string]interface{})

				// --- 最初のスタックフレームだけを取り出す ---
				if traces, ok := m["stacktrace"].([]interface{}); ok && len(traces) > 0 {
					f := traces[0].(map[string]interface{})
					b.WriteString("  ")
					// ファイル:行
					b.WriteString(cFile)
					b.WriteString(fmt.Sprintf("%s", f["source"]))
					b.WriteString(cReset)
					b.WriteString(":")
					b.WriteString(cLine)
					b.WriteString(fmt.Sprintf("%v", f["line"]))
					b.WriteString(cReset)
					// 関数名
					b.WriteString(" ")
					b.WriteString(cFunc)
					b.WriteString(fmt.Sprintf("%s()", f["func"]))
					b.WriteString(cReset)
					b.WriteRune('\n')
				}

				// --- そのスタックエントリの details（コメント）だけを出力 ---
				if details, ok := m["details"].([]interface{}); ok && len(details) > 0 {
					for _, d := range details {
						b.WriteString("    ")
						b.WriteString(fmt.Sprint(d))
						b.WriteRune('\n')
					}
				}
			}

			return b.String()
		},
	}
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
		return crerr.Wrap(err, "failed to open the database")
	}
	return nil
}

func persistentPostRunE(cmd *cobra.Command, args []string) error {
	dbMiddleware := &dbMiddleware{}
	if err := dbMiddleware.PostRunE(cmd, args); err != nil {
		return crerr.Wrap(err, "failed to close the database")
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
		return crerr.Wrap(err, "failed to open the database")
	}
	return nil
}

func (m *dbMiddleware) PostRunE(cmd *cobra.Command, args []string) error {
	if err := db.Close(); err != nil {
		return crerr.Wrap(err, "failed to close the database")
	}
	return nil
}
