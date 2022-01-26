package regex

import (
	"regexp"
	"sync"
)

var (
	regexCache     = make(map[string]*regexp.Regexp)
	regexCacheLock = sync.RWMutex{}
)

func GetCompiledRegex(pattern string) *regexp.Regexp {
	// try in cache first
	regexCacheLock.RLock()
	re := regexCache[pattern]
	regexCacheLock.RUnlock()
	if re != nil {
		return re
	}

	// should we panic or return the error?
	re = regexp.MustCompile(pattern)

	// lock for concurrent access and save the compiled expression in cache
	regexCacheLock.Lock()
	regexCache[pattern] = re
	regexCacheLock.Unlock()

	return re
}

func FindNamedRegexMatch(pattern, text string) map[string]string {
	// error ignored because mustCompile will cause a panic
	re := GetCompiledRegex(pattern)
	match := re.FindStringSubmatch(text)
	result := make(map[string]string)
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
	re := GetCompiledRegex(pattern)
	match := re.FindAllStringSubmatch(text, -1)
	var results []map[string]string
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
	re := GetCompiledRegex(pattern)
	return re.ReplaceAllString(text, replaceText)
}

func MatchString(pattern, text string) bool {
	re := GetCompiledRegex(pattern)
	return re.MatchString(text)
}
