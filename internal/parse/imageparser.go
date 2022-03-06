package parse

import (
	"fmt"
	"regexp"
)

type ImageLink struct {
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

func transformImageLink(imageLink ImageLink) string {
	imageName := sanitizeImageName(imageLink)
	return fmt.Sprintf("![%s](/img/%s)", imageName, imageName)
}

func sanitizeImageName(location ImageLink) string {
	return sanitize(getImageFileName(location.content))
}

func getImageFileName(content string) string {
	name, ext := parseObsidianImageLink(content)
	return name + ext
}

func parseObsidianImageLink(content string) (string, string) {
	search := regexp.MustCompile("!\\[\\[(.*)(\\.png)]]")
	matches := search.FindStringSubmatch(content)
	name := matches[1]
	ext := matches[2]
	return name, ext
}

func getImageLinkLocations(contents string) []ImageLink {
	var locations []ImageLink
	search := regexp.MustCompile("!\\[\\[(.*\\.png)]]")
	found := true
	offset := 0
	for found {
		loc := search.FindIndex([]byte(contents))
		if len(loc) > 0 {
			l, r := loc[0], loc[1]
			locations = append(locations, ImageLink{
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

