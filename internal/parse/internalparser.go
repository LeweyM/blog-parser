package parse

import (
	"fmt"
	"regexp"
)

const internalLinkRegex = "[^!]\\[\\[(.*)]]"

func parseInternalLinks(contents string) string {
	newPost := ""
	internalLinkLocations := getInternalLinkLocations(contents)

	l := 0
	for _, internalLink := range internalLinkLocations {
		newPost += contents[l:internalLink.l]
		newPost += transformInternalLink(internalLink)
		l = internalLink.r
	}
	newPost += contents[l:]

	return newPost
}

func getInternalLinkLocations(contents string) []ContentLink {
	return getLinkLocations(contents, regexp.MustCompile(internalLinkRegex))
}

func transformInternalLink(internalLink ContentLink) string {
	sanitizedLink := sanitize(internalLink.content)
	reg := regexp.MustCompile(internalLinkRegex)
	linkTitle := reg.FindStringSubmatch(internalLink.content)[1]
	return fmt.Sprintf("[%s]({{< ref \"%s\" >}})", linkTitle, sanitizedLink)
}
