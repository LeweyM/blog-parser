package parse

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
)

func Handle(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.New(fmt.Sprintf("could not read path %s", path))
	}

	draftFiles, err := ioutil.ReadDir(getDraftPath(path))
	if err != nil {
		return errors.New(fmt.Sprintf("could not read draft path %s", getDraftPath(path)))
	}

	for _, draftFile := range draftFiles {
		if isMarkdown(draftFile) {
			parseDraft(draftFile)
		}
	}

	for _, file := range files {
		if isMarkdown(file) {
			parseFile(file)
		}
	}

	return nil
}

func getDraftPath(path string) string {
	return filepath.Join(path, "drafts")
}

func parseDraft(draft fs.FileInfo) {
	name := draft.Name()
	fmt.Printf("\nDRAFT: %s", name)
}

func isMarkdown(file fs.FileInfo) bool {
	ext := filepath.Ext(file.Name())
	return ext == ".md"
}

func parseFile(file fs.FileInfo) {
	name := file.Name()
	fmt.Printf("\npost: %s", name)
}
