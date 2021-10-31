package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

var (
	regexCache     = make(map[string]*regexp.Regexp)
	regexCacheLock = sync.RWMutex{}
)

func getCompiledRegex(pattern string) *regexp.Regexp {
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

func findNamedRegexMatch(pattern, text string) map[string]string {
	// error ignored because mustCompile will cause a panic
	re := getCompiledRegex(pattern)
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

func findAllNamedRegexMatch(pattern, text string) []map[string]string {
	re := getCompiledRegex(pattern)
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

func replaceAllString(pattern, text, replaceText string) string {
	re := getCompiledRegex(pattern)
	return re.ReplaceAllString(text, replaceText)
}

func matchString(pattern, text string) bool {
	re := getCompiledRegex(pattern)
	return re.MatchString(text)
}

func dirMatchesOneOf(env environmentInfo, dir string, regexes []string) bool {
	normalizedCwd := strings.ReplaceAll(dir, "\\", "/")
	normalizedHomeDir := strings.ReplaceAll(env.homeDir(), "\\", "/")

	for _, element := range regexes {
		normalizedElement := strings.ReplaceAll(element, "\\\\", "/")
		if strings.HasPrefix(normalizedElement, "~") {
			normalizedElement = normalizedHomeDir + normalizedElement[1:]
		}
		pattern := fmt.Sprintf("^%s$", normalizedElement)
		goos := env.getRuntimeGOOS()
		if goos == windowsPlatform || goos == darwinPlatform {
			pattern = "(?i)" + pattern
		}
		matched := matchString(pattern, normalizedCwd)
		if matched {
			return true
		}
	}
	return false
}
