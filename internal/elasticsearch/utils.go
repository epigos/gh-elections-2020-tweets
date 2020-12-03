package elasticsearch

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/epigos/gh-elections-2020-tweets/internal/config"
	"regexp"
	"strings"
)

var (
	sourceReg = regexp.MustCompile(`<a.+?>twitter\s+?for\s+?(.*)<\/a>`)
)

func getParty(h string) string {
	hashtag := fmt.Sprintf("#%s", h)

	if has(hashtag, config.NPPTags) {
		return "NPP"
	} else if has(hashtag, config.NDCTags) {
		return "NDC"
	}
	return ""
}

func parseLocation(l string) string {
	loc := strings.ToLower(l)

	if strings.Contains(loc, "accra") {
		return "Accra"
	} else if strings.Contains(loc, "kumasi") {
		return "Kumasi"
	}
	return l
}

func parseSource(s string) string {
	src := sourceReg.FindStringSubmatch(strings.ToLower(s))
	if len(src) > 0 {
		return src[len(src)-1]
	}
	return s
}

func parseUserMentions(t *twitter.Tweet) []userMention {
	var userMentions []userMention
	for _, u := range t.Entities.UserMentions {
		userMentions = append(userMentions, userMention{u.Name, u.ScreenName, u.IDStr})
	}

	return userMentions
}
func has(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
