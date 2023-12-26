package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

type PatternType struct {
	pattern string
	start   rune
	ends    []rune
}

func (p *PatternType) ContainsInEnds(c rune) bool {
	for _, end := range p.ends {
		if c == end {
			return true
		}
	}
	return false
}

// GeneratePatternType generate default patter type
func GeneratePatternType(keyword string) (patterns []PatternType) {
	for _, p := range []PatternType{
		{pattern: `"%s"`, start: ':', ends: []rune{',', '}'}},
		{pattern: "<%s>", start: 0, ends: []rune{'<'}},
		{pattern: "%s", start: ':', ends: []rune{' ', ','}},
		{pattern: "%s", start: '=', ends: []rune{'&'}},
	} {
		patterns = append(patterns, PatternType{
			pattern: fmt.Sprintf(p.pattern, keyword), start: p.start, ends: p.ends,
		})
	}
	return patterns
}

type Masker interface {
	Mask(text string) string
}

type maskImpl struct {
	keywords []string
}

// NewMasker create new logger masker with pattern, if patternType is empty will be create with default pattern type password
func NewMasker(keywords ...string) Masker {
	if len(keywords) == 0 { // default pattern
		keywords = []string{"password"}
	}
	return &maskImpl{
		keywords: keywords,
	}
}

func (r *maskImpl) Mask(text string) string {
	isJSON := json.Valid([]byte(text))
	for _, keyword := range r.keywords {
		for _, pat := range GeneratePatternType(keyword) {
			idx := strings.Index(text, pat.pattern)
			if idx < 0 {
				continue
			}
			isStart, isValueStart := pat.start == 0, false
			isMask := false
			maskStart, maskEnd := 0, 0
			start := idx + len(pat.pattern)
			for i, c := range text[start:] {
				i += start
				isSpace := unicode.IsSpace(c)

				if isStart && !isSpace {
					isValueStart = true
				}
				switch c {
				case pat.start:
					isStart = true
				}

				if isStart {
					if isValueStart {
						isMask = true
						if maskStart == 0 {
							maskStart = i
						}

						if (!isJSON && isSpace) || pat.ContainsInEnds(c) || i == len(text)-1 {
							isValueStart = false
							maskEnd = i
							break
						}
					}
				} else if !isSpace {
					break
				}
			}
			if isMask && maskStart > 0 && maskEnd > maskStart && maskStart < len(text) && maskEnd < len(text) {
				mask := `"xxxxx"`
				if maskEnd == len(text)-1 && !pat.ContainsInEnds(rune(text[maskEnd])) {
					maskEnd = len(text)
				}
				text = text[:maskStart] + mask + text[maskEnd:]
				break
			}
		}
	}
	return text
}
