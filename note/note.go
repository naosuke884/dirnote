package note

import (
	"os"

	crerr "github.com/cockroachdb/errors"
)

type Note struct {
	Directory Directory
	Body      string
}

type Directory struct {
	Path         string
	CreationTime string
}

func NewDirectory(path string) (Directory, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Directory{}, crerr.Wrap(err, "failed to get directory info")
	}
	if !info.IsDir() {
		return Directory{}, crerr.New(path + " is not a directory")
	}
	return Directory{
		Path:         path,
		CreationTime: info.ModTime().Format("2006-01-02 15:04:05"),
	}, nil
}
