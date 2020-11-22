package main

import "regexp"

func findNamedRegexMatch(pattern, text string) map[string]string {
	re := regexp.MustCompile(pattern)
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
	re := regexp.MustCompile(pattern)
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
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(text, replaceText)
}
