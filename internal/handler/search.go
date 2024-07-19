package handler

import "strings"

func SearchContext(text, searchTerm string, contextWords int) string {
	words := strings.Fields(text)
	searchTerm = strings.ToLower(searchTerm)
	termWords := strings.Fields(searchTerm)

	var context string
	var contexts []string
	termLength := len(termWords)

	if len(text) < contextWords*2 {
		return text
	}

	for _, w := range termWords {
		for i := 0; i <= len(words)-termLength; i++ {
			if strings.ToLower(words[i]) == strings.ToLower(w) {
				start := max(0, i-contextWords)
				end := min(len(words), i+termLength+contextWords)
				context = strings.Join(words[start:end], " ")
				contexts = append(contexts, context)

				break
			}
		}
	}

	if len(contexts) > 0 {
		return strings.Join(contexts, " ... ")
	}

	return context
}
