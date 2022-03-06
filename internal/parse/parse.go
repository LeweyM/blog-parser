package parse

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Handler struct {
	root, imgPath string
}

func NewHandler(root, imgPath string) *Handler {
	return &Handler{root: root, imgPath: imgPath}
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
			h.parseFile(filepath.Join(draftPath, draftFile.Name()), true)
		}
	}

	for _, file := range files {
		if isMarkdown(file) {
			h.parseFile(filepath.Join(postsPath, file.Name()), false)
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

func (h *Handler) parseFile(path string, isDraft bool) {
	tempFile := createTempFile()
	defer os.Remove(tempFile.Name())

	contents := getContentsFromFile(path)

	h.transformAndCopyImageFiles(getImageLinkLocations(contents))

	contents = parseImageLinks(contents)
	contents = parseInternalLinks(contents)
	contents = addHeader(contents, getTitleFromPath(path), isDraft)

	writeToTempFile(tempFile, contents)
	dest := filepath.Join(".", "out", "posts", fmt.Sprintf("%s.md", sanitize(getTitleFromPath(path))))
	copyFile(tempFile.Name(), dest)
}

func getLinkLocations(contents string, reg *regexp.Regexp) []ContentLink {
	var locations []ContentLink
	found := true
	offset := 0
	for found {
		loc := reg.FindIndex([]byte(contents))
		if len(loc) > 0 {
			l, r := loc[0], loc[1]
			locations = append(locations, ContentLink{
				l:       l + offset,
				r:       r + offset,
				content: contents[l:r],
			})
			offset += r
			contents = contents[r:]
		} else {
			found = false
		}
	}
	return locations
}

func (h *Handler) transformAndCopyImageFiles(locations []ContentLink) {
	for _, location := range locations {
		src := filepath.Join(h.root, getImageFileName(location.content))
		dest := filepath.Join("out", "img", sanitizeImageName(location))
		copyFile(src, dest)
	}
}

func sanitize(title string) string {
	title = regexp.MustCompile(`[^a-zA-Z0-9-_\. ]`).ReplaceAllString(title, "")
	title = regexp.MustCompile(` `).ReplaceAllString(title, "-")
	return title
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

func copyFile(src, dest string) {
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
