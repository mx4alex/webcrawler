package extracter

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func reverseSortByValue(myMap map[string]float64) PairList {
	pl := make(PairList, len(myMap))
	i := 0
	for k, v := range myMap {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value float64
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func RegexSplitWords() *regexp.Regexp {
	return regexp.MustCompile("[\\p{L}\\d_]+")
}

func RegexSplitSentences() *regexp.Regexp {
	return regexp.MustCompile(`[.,\/#!$%\^&\*;:{}=\-_~()]`)
}

func RegexStopWords(stopWordsSlice []string) *regexp.Regexp {
	stopWordRegexList := []string{}

	for _, word := range stopWordsSlice {
		wordRegex := fmt.Sprintf(`(?:\A|\z|\s)%s(?:\A|\z|\s)`, word)
		stopWordRegexList = append(stopWordRegexList, wordRegex)
	}

	re := regexp.MustCompile(`(?i)` + strings.Join(stopWordRegexList, "|"))
	return re
}

func IsNumber(str string) bool {
	if strings.Contains(str, ".") {
		if _, err := strconv.ParseFloat(str, 32); err != nil {
			return false
		}
	} else {
		if _, err := strconv.ParseInt(str, 10, 32); err != nil {
			return false
		}
	}

	return true
}

func SeperateWords(text string) []string {
	words := []string{}

	splitWords := RegexSplitWords().FindAllString(text, -1)
	for _, singleword := range splitWords {
		currentword := strings.ToLower(strings.TrimSpace(singleword))
		if currentword != "" && !IsNumber(currentword) {
			words = append(words, currentword)
		}
	}

	return words
}

func SplitSentences(text string) []string {
	splitText := RegexSplitSentences().ReplaceAllString(strings.TrimSpace(text), "\n")
	return strings.Split(splitText, "\n")
}

func GenerateCandidateKeywords(sentenceList []string, stopWordPattern *regexp.Regexp) []string {
	phraseList := []string{}

	for _, sentence := range sentenceList {
		tmp := stopWordPattern.ReplaceAllString(strings.TrimSpace(sentence), " | ")
		for {
			abc := len(tmp)
			tmp = stopWordPattern.ReplaceAllString(strings.TrimSpace(tmp), " | ")
			if abc == len(tmp) {
				break
			}
		}

		multipleWhiteSpaceRe := regexp.MustCompile(`\s\s+`)
		tmp = multipleWhiteSpaceRe.ReplaceAllString(strings.TrimSpace(tmp), " ")

		phrases := strings.Split(tmp, "|")
		for _, phrase := range phrases {
			phrase = strings.ToLower(strings.TrimSpace(phrase))
			if phrase != "" {
				phraseList = append(phraseList, phrase)
			}
		}
	}

	return phraseList
}

func CalculateWordScores(phraseList []string) map[string]float64 {
	wordFrequency := map[string]int{}
	wordDegree := map[string]int{}

	for _, phrase := range phraseList {
		wordList := SeperateWords(phrase)
		wordListLength := len(wordList)
		wordListDegree := wordListLength - 1

		for _, word := range wordList {
			SetDefaultStringInt(wordFrequency, word, 0)
			wordFrequency[word]++

			SetDefaultStringInt(wordDegree, word, 0)
			wordDegree[word] += wordListDegree
		}
	}

	for key := range wordFrequency {
		wordDegree[key] = wordDegree[key] + wordFrequency[key]
	}

	wordScore := map[string]float64{}
	for key := range wordFrequency {
		SetDefaultStringFloat64(wordScore, key, 0)
		wordScore[key] = float64(wordDegree[key]) / float64(wordFrequency[key])
	}

	return wordScore
}

func GenerateCandidateKeywordScores(phraseList []string, wordScore map[string]float64) map[string]float64 {
	keywordCandidates := map[string]float64{}

	for _, phrase := range phraseList {
		SetDefaultStringFloat64(keywordCandidates, phrase, 0)
		wordList := SeperateWords(phrase)
		candidateScore := float64(0.0)

		for _, word := range wordList {
			candidateScore = candidateScore + wordScore[word]
		}

		keywordCandidates[phrase] = candidateScore
	}

	return keywordCandidates
}

func SetDefaultStringInt(h map[string]int, k string, v int) (set bool, r int) {
	if r, set = h[k]; !set {
		h[k] = v
		r = v
		set = true
	}
	return
}

func SetDefaultStringFloat64(h map[string]float64, k string, v float64) (set bool, r float64) {
	if r, set = h[k]; !set {
		h[k] = v
		r = v
		set = true
	}
	return
}

func IsRus(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}

func RunRake(text string) PairList {
	var words []string

	sentenceList := SplitSentences(text)

	if IsRus(text) {
		words = StopWordsSliceRu
	} else {
		words = StopWordsSliceEn
	}

	stopWordPattern := RegexStopWords(words)

	phraseList := GenerateCandidateKeywords(sentenceList, stopWordPattern)

	wordScores := CalculateWordScores(phraseList)

	keywordCandidates := GenerateCandidateKeywordScores(phraseList, wordScores)
	sorted := reverseSortByValue(keywordCandidates)
	return sorted
}
