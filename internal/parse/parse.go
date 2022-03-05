package parse

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Handler struct {
	root string
}

func NewHandler(root string) *Handler {
	return &Handler{root: root}
}

func (h *Handler) Handle() error {
	postsPath := filepath.Join(h.root, "Blog posts")
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
			parseFile(filepath.Join(draftPath, draftFile.Name()), true)
		}
	}

	for _, file := range files {
		if isMarkdown(file) {
			parseFile(filepath.Join(postsPath, file.Name()), false)
		}
	}

	return nil
}

func getDraftPath(path string) string {
	return filepath.Join(path, "drafts")
}

func isMarkdown(file fs.FileInfo) bool {
	ext := filepath.Ext(file.Name())
	return ext == ".md"
}

func parseFile(path string, isDraft bool) {
	tempFile := createTempFile()
	defer os.Remove(tempFile.Name())

	contents := getContentsFromFile(path)
	contents = parseImageLinks(contents)
	contents = addHeader(contents, getTitleFromPath(path), isDraft)

	writeToTempFile(tempFile, contents)
	copyToOut(tempFile.Name(), getTitleFromPath(path))
}

func writeToTempFile(tempFile *os.File, contents string) {
	_, err := tempFile.WriteString(contents)
	if err != nil {
		panic(err)
	}
}

func getContentsFromFile(path string) string {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	contents := string(file)
	return contents
}

func createTempFile() *os.File {
	dir, err := ioutil.TempDir("", "*")
	tempFile, err := ioutil.TempFile(dir, "tmp")
	if err != nil {
		panic(err)
	}
	return tempFile
}

func copyToOut(src, title string) {
	dest := filepath.Join(".", "out", fmt.Sprintf("%s.md", title))
	os.Create(dest)
	source, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	if err != nil {
		fmt.Println(err)
	}
}

func getTitleFromPath(path string) string {
	_, fileName := filepath.Split(path)
	fileName = strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	return fileName
}

func addHeader(contents, title string, isDraft bool) string {
	header := fmt.Sprintf(`---
title: %s
draft: %t
---`, title, isDraft)
	return header + "\n\r" + contents
}
