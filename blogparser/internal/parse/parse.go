package parse

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Handle(root, imagesPath, outPath string) error {
	postsPath := filepath.Join(root, "Blog posts")
	files, err := ioutil.ReadDir(postsPath)
	if err != nil {
		return errors.New(fmt.Sprintf("could not read path %s", postsPath))
	}

	draftPath := getDraftPath(postsPath)
	draftFiles, err := ioutil.ReadDir(draftPath)
	if err != nil {
		return errors.New(fmt.Sprintf("could not read draft path %s", draftPath))
	}

	for _, draftFile := range draftFiles {
		if isMarkdown(draftFile) {
			parseDraft(draftFile)
		}
	}

	for _, file := range files {
		if isMarkdown(file) {
			name := file.Name()
			fmt.Printf("\npost: %s", name)
			parseFile(filepath.Join(postsPath, file.Name()))
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

func parseFile(path string) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("\n%v", err)
		return
	}

	dir, err := ioutil.TempDir("", "*")
	tempFile, err := ioutil.TempFile(dir, "tmp")
	if err != nil {
		fmt.Printf("\n%v", err)
		return
	}
	defer os.Remove(tempFile.Name())

	contents := string(file)
	contents = parseImageLinks(contents)

	_, err = tempFile.WriteString(contents)
	if err != nil {
		fmt.Printf("\n%v", err)
		return
	}

	fmt.Println("")
	fmt.Println(tempFile.Name())
}
