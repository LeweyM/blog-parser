package parse

import (
	"fmt"
	"regexp"
)

const imageLinkRegex = "!\\[\\[(.*)(\\.png|\\.gif)]]"

type ContentLink struct {
	l, r    int
	content string
}

func parseImageLinks(contents string) string {
	newPost := ""
	imageLinks := getImageLinkLocations(contents)

	l := 0
	for _, imageLink := range imageLinks {
		newPost += contents[l:imageLink.l]
		newPost += transformImageLink(imageLink)
		l = imageLink.r
	}
	newPost += contents[l:]

	return newPost
}

func transformImageLink(imageLink ContentLink) string {
	imageName := sanitizeImageName(imageLink)
	return fmt.Sprintf("![%s](/img/%s)", imageName, imageName)
}

func sanitizeImageName(location ContentLink) string {
	return sanitize(getImageFileName(location.content))
}

func getImageFileName(content string) string {
	name, ext := parseObsidianImageLink(content)
	return name + ext
}

func parseObsidianImageLink(content string) (string, string) {
	search := regexp.MustCompile(imageLinkRegex)
	matches := search.FindStringSubmatch(content)
	name := matches[1]
	ext := matches[2]
	return name, ext
}

func getImageLinkLocations(contents string) []ContentLink {
	return getLinkLocations(contents, regexp.MustCompile(imageLinkRegex))
}
