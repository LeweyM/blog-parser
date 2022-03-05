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
	search := regexp.MustCompile("!\\[\\[(.*)(\\.png)]]")
	matches := search.FindStringSubmatch(imageLink.content)
	name := matches[1]
	ext := matches[2]
	return fmt.Sprintf("![%s](/static/img/%s%s)", name, name, ext)
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

