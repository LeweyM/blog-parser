package parse

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
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
	postEntries, err := os.ReadDir(postsPath)
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
			h.parseFile(filepath.Join(draftPath, draftFile.Name()), true, "")
		}
	}

	for _, postEntry := range postEntries {
		if postEntry.IsDir() {
			seriesEntries, _ := os.ReadDir(filepath.Join(postsPath, postEntry.Name()))
			for _, seriesEntry := range seriesEntries {
				// no nested series
				// special folder drafts is hardcoded
				if seriesEntry.Name() == "_index.md" {
					h.parseSeriesMetadata(filepath.Join(postsPath, postEntry.Name(), seriesEntry.Name()), postEntry.Name())
				} else if !seriesEntry.IsDir() && postEntry.Name() != "drafts" {
					h.parseSeries(filepath.Join(postsPath, postEntry.Name(), seriesEntry.Name()), false, postEntry.Name())
				}
			}
		} else {
			fileInfo, _ := postEntry.Info()
			if isMarkdown(fileInfo) {
				h.parseFile(filepath.Join(postsPath, postEntry.Name()), false, "")
			}
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

func (h *Handler) parseFile(path string, isDraft bool, series string) {

	fmt.Printf("\n%s, %s", path, series)

	tempFile := createTempFile()
	defer os.Remove(tempFile.Name())

	contents := getContentsFromFile(path)

	h.transformAndCopyImageFiles(getImageLinkLocations(contents))

	contents = parseImageLinks(contents)
	contents = parseInternalLinks(contents)
	contents = addHeader(contents, getTitleFromPath(path), isDraft, series)

	writeToTempFile(tempFile, contents)
	dest := filepath.Join(".", "out", "content", "posts", fmt.Sprintf("%s.md", sanitize(getTitleFromPath(path))))
	os.MkdirAll(filepath.Join(".", "out", "content", "posts"), 0750)
	copyFile(tempFile.Name(), dest)
}

func parseCustomSearchShortcodes(contents string) string {
	shortcodeRegex := regexp.MustCompile("{{< search(.*)>}}")

	for indices := shortcodeRegex.FindStringIndex(contents); len(indices) != 0; indices = shortcodeRegex.FindStringIndex(contents) {
		beginning := indices[0] // inclusive
		end := indices[1]       // exclusive

		htmlFilePath := runSearchHTMLFileBuilder(contents[beginning:end])

		contents = contents[:beginning] + fmt.Sprintf(`{{< iframe src="%s" >}}`, htmlFilePath) + contents[end:]
	}

	return contents
}

func runSearchHTMLFileBuilder(contents string) string {
	shortcodeRegex := regexp.MustCompile("{{< search(.*)>}}")

	submatch := shortcodeRegex.FindStringSubmatch(contents)
	commandStr := strings.TrimSpace(submatch[1])

	args := getArgsFromCommandString(commandStr)

	path := filepath.Join("html")

	fileName := getFileName(args)
	filePath := filepath.Join(path, fileName)

	searchArgs := append(append([]string{args[0], "out"}, args[1:]...), filepath.Join("out", filePath))
	cmd := exec.Command("search", searchArgs...)

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return filePath
}

func getArgsFromCommandString(commandStr string) []string {
	res := []string{}

	currentWord := ""
	isInDoubleQuotes := false
	isInSingleQuotes := false
	for _, char := range commandStr {
		s := string(char)
		print(s)

		if char == '"' && !isInDoubleQuotes && !isInSingleQuotes {
			isInDoubleQuotes = true
		} else if char == '"' && isInDoubleQuotes {
			if len(strings.TrimSpace(currentWord)) > 0 {
				res = append(res, currentWord)
			}
			currentWord = ""
		} else if char == '\'' && !isInDoubleQuotes && !isInSingleQuotes {
			isInSingleQuotes = true
		} else if char == '\'' && isInSingleQuotes {
			if len(strings.TrimSpace(currentWord)) > 0 {
				res = append(res, currentWord)
			}
			currentWord = ""
		} else if char != ' ' && !isInDoubleQuotes && !isInSingleQuotes {
			currentWord += string(char)
		} else if isInDoubleQuotes || isInDoubleQuotes {
			currentWord += string(char)
		} else {
			if len(strings.TrimSpace(currentWord)) > 0 {
				res = append(res, currentWord)
			}
			currentWord = ""
			isInDoubleQuotes = false
			isInSingleQuotes = false
		}
	}

	if len(strings.TrimSpace(currentWord)) > 0 {
		res = append(res, currentWord)
	}

	return trimArgSpacesAndQuotations(res)
}

func trimArgSpacesAndQuotations(ss []string) []string {
	res := []string{}

	for _, s := range ss {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			res = append(res, s[1:len(s)-1])
		} else if strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`) {
			res = append(res, s[1:len(s)-1])
		} else {
			res = append(res, s)
		}
	}
	return res
}

func getFileName(args []string) string {
	//fileName := strings.Join(args, "-")
	hash := md5.New()
	hash.Write([]byte(strings.Join(args, "-")))
	res := hash.Sum(nil)

	return fmt.Sprintf("%x", res)
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
		os.MkdirAll(filepath.Join("out", "img"), 0750)
		dest := filepath.Join("out", "img", sanitizeImageName(location))
		copyFile(src, dest)
	}
}
func (h *Handler) parseSeries(path string, isDraft bool, series string) {
	fmt.Printf("\n%s, %s", path, series)

	tempFile := createTempFile()
	defer os.Remove(tempFile.Name())

	contents := getContentsFromFile(path)

	h.transformAndCopyImageFiles(getImageLinkLocations(contents))

	contents = parseImageLinks(contents)
	contents = parseInternalLinks(contents)
	contents = addHeader(contents, getTitleFromPath(path), isDraft, series)
	contents = parseCustomSearchShortcodes(contents)

	writeToTempFile(tempFile, contents)
	folderPath := filepath.Join(".", "out", "content", "series", series)
	os.MkdirAll(folderPath, 0750)
	dest := filepath.Join(folderPath, fmt.Sprintf("%s.md", sanitize(getTitleFromPath(path))))
	copyFile(tempFile.Name(), dest)
}

func (h *Handler) parseSeriesMetadata(path, series string) {
	metadata := getContentsFromFile(path)
	header := extractHeader(metadata)
	body := extractContents(metadata)

	header.seriesDescription = fmt.Sprintf(`["%s"]`, series)
	header.title = series

	imageLinkParts := regexp.MustCompile(imageLinkRegex).FindStringSubmatch(header.thumbnailSrc)
	header.thumbnailSrc = filepath.Join("/img", sanitize(fmt.Sprintf(imageLinkParts[1]+imageLinkParts[2])))
	h.transformAndCopyImageFiles(getImageLinkLocations(metadata))

	newFile := strings.Join([]string{buildHeader(header), body}, "\n")

	tempFile := createTempFile()
	defer os.Remove(tempFile.Name())
	writeToTempFile(tempFile, newFile)
	os.MkdirAll(filepath.Join(".", "out", "content", "series-descriptions"), 0750)
	copyFile(tempFile.Name(), filepath.Join(".", "out", "content", "series-descriptions", fmt.Sprintf("%s.md", series)))
}

type Header struct {
	title             string
	seriesDescription string
	thumbnailSrc      string
}

func extractHeader(fileContents string) Header {
	h := Header{}
	for _, s := range strings.Split(fileContents, "\n") {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "title:") {
			s = strings.Replace(s, "title:", "", 1)
			s = strings.TrimSpace(s)
			h.title = s
		}
		if strings.HasPrefix(s, "image:") {
			s = strings.Replace(s, "image:", "", 1)
			s = strings.TrimSpace(s)
			h.thumbnailSrc = s
		}
	}
	return h
}

func extractContents(fileContents string) string {
	return strings.TrimSpace(regexp.MustCompile("(?s)---(.*)---").ReplaceAllString(fileContents, ""))
}

func buildHeader(header Header) string {
	res := []string{}
	if header.title != "" {
		res = append(res, fmt.Sprintf("title: %s", header.title))
	}
	if header.thumbnailSrc != "" {
		res = append(res, fmt.Sprintf("image: %s", header.thumbnailSrc))
	}
	if header.seriesDescription != "" {
		res = append(res, fmt.Sprintf("seriesdesc: %s", header.seriesDescription))
	}

	return fmt.Sprintf(`---
%s
---`, strings.Join(res, "\n"))
}

func sanitize(title string) string {
	title = regexp.MustCompile(`[^a-zA-Z0-9-_\. ]`).ReplaceAllString(title, "")
	title = strings.TrimSpace(title)
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

func addHeader(contents, title string, isDraft bool, series string) string {
	ops := []string{
		fmt.Sprintf("title: %s", title),
		fmt.Sprintf("draft: %t", isDraft),
	}
	if series != "" {
		ops = append(ops, fmt.Sprintf(`series: ["%s"]`, series))
	}

	header := fmt.Sprintf(`---
%s
---`, strings.Join(ops, "\n"))

	return header + "\n\r" + contents
}
