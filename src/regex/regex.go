package regex

import (
	"regexp"
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

var (
	regexCache     = make(map[string]*regexp.Regexp)
	regexCacheLock = sync.RWMutex{}
)

const (
	LINK = `(?P<STR>\x1b]8;;(.+)\x1b\\(?P<TEXT>.+)\x1b]8;;\x1b\\)`
)

func GetCompiledRegex(pattern string) (*regexp.Regexp, error) {
	// try in cache first
	regexCacheLock.RLock()
	re := regexCache[pattern]
	regexCacheLock.RUnlock()
	if re != nil {
		return re, nil
	}

	// should we panic or return the error?
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// lock for concurrent access and save the compiled expression in cache
	regexCacheLock.Lock()
	regexCache[pattern] = re
	regexCacheLock.Unlock()

	return re, nil
}

func FindNamedRegexMatch(pattern, text string) map[string]string {
	result := make(map[string]string)

	re, err := GetCompiledRegex(pattern)
	if err != nil {
		return result
	}

	match := re.FindStringSubmatch(text)
	if len(match) == 0 {
		return result
	}

	for i, name := range re.SubexpNames() {
		if i == 0 {
			continue
		}
		result[name] = match[i]
	}

	return result
}

func FindAllNamedRegexMatch(pattern, text string) []map[string]string {
	var results []map[string]string

	re, err := GetCompiledRegex(pattern)
	if err != nil {
		return results
	}

	match := re.FindAllStringSubmatch(text, -1)

	if len(match) == 0 {
		return results
	}

	for _, set := range match {
		result := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i == 0 {
				result["text"] = set[i]
				continue
			}
			result[name] = set[i]
		}
		results = append(results, result)
	}

	return results
}

func ReplaceAllString(pattern, text, replaceText string) string {
	re, err := GetCompiledRegex(pattern)
	if err != nil {
		return text
	}

	return re.ReplaceAllString(text, replaceText)
}

func MatchString(pattern, text string) bool {
	re, err := GetCompiledRegex(pattern)
	if err != nil {
		return false
	}

	return re.MatchString(text)
}

func FindStringMatch(pattern, text string, index int) (string, bool) {
	re, err := GetCompiledRegex(pattern)
	if err != nil {
		return text, false
	}

	matches := re.FindStringSubmatch(text)
	if len(matches) <= index {
		return text, false
	}

	match := matches[index]
	if len(match) == 0 {
		return text, false
	}

	return match, true
}
