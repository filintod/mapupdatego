package prefix

import (
	"regexp"
	"strconv"
	"strings"
)

func getMapIdxOfNameRegex(re *regexp.Regexp) map[string]int {
	m := make(map[string]int)
	for i, n := range re.SubexpNames() {
		if i != 0 {
			m[n] = i
		}
	}
	return m
}

func NewPrefix() Prefix {
	p := Prefix{}
	p.remove = REMOVE_PREFIX
	p.removeAll = REMOVE_ALL_PREFIX
	p.append = APPEND_PREFIX
	p.appendAll = APPEND_ALL_PREFIX
	p.indexRegex = regexp.MustCompile(INDEX_PREFIX_REGEX)
	p.indexRegexAppend = regexp.MustCompile(INDEX_PREFIX_REGEX_APPEND)
	p.indexRegexAppendAll = regexp.MustCompile(INDEX_PREFIX_REGEX_APPEND_ALL)

	return p
}

// Default values for prefixes for appending and remove
const (
	REMOVE_PREFIX                 = `(-)`
	REMOVE_ALL_PREFIX             = `(--)`
	APPEND_PREFIX                 = `(+)`
	APPEND_ALL_PREFIX             = `(++)`
	INDEX_PREFIX_REGEX            = `^\((?P<index>\d+)\)(?P<key>.+)`
	INDEX_PREFIX_REGEX_APPEND     = `^\((?P<index>\d+)\+\)(?P<key>.+)`
	INDEX_PREFIX_REGEX_APPEND_ALL = `^\((?P<index>\d+)\+\+\)(?P<key>.+)`
)

type Prefix struct {
	remove                 string
	removeAll              string
	append                 string
	appendAll              string
	indexRegex             *regexp.Regexp
	indexRegexMap          map[string]int
	indexRegexAppend       *regexp.Regexp
	indexRegexAppendMap    map[string]int
	indexRegexAppendAll    *regexp.Regexp
	indexRegexAppendAllMap map[string]int
	lastMatch              []string
	lastRegexMap           map[string]int
}

func (p Prefix) HasRemove(other string) bool {
	return strings.HasPrefix(other, p.remove)
}

func (p Prefix) TrimRemove(other string) string {
	return strings.TrimPrefix(other, p.remove)
}

func (p Prefix) HasRemoveAll(other string) bool {
	return strings.HasPrefix(other, p.removeAll)
}

func (p Prefix) TrimRemoveAll(other string) string {
	return strings.TrimPrefix(other, p.removeAll)
}

func (p Prefix) HasAppend(other string) bool {
	return strings.HasPrefix(other, p.append)
}

func (p Prefix) TrimAppend(other string) string {
	return strings.TrimPrefix(other, p.append)
}

func (p Prefix) HasAppendAll(other string) bool {
	return strings.HasPrefix(other, p.appendAll)
}

func (p Prefix) TrimAppendAll(other string) string {
	return strings.TrimPrefix(other, p.appendAll)
}

func (p Prefix) TrimAll(other string) string {
	return p.TrimAppend(p.TrimAppendAll(p.TrimRemove(p.TrimRemoveAll(other))))
}

// hasRegex is the only pointer receiver method of Prefix to store the last match regex expression
//			this deviate from standard procedure to make it more readable
func (p *Prefix) hasRegex(other string, re *regexp.Regexp) bool {
	p.lastRegexMap = getMapIdxOfNameRegex(re)
	p.lastMatch = re.FindStringSubmatch(other)
	return p.lastMatch != nil
}

func (p Prefix) GetLastIndex() int {
	if p.lastMatch != nil {
		v, _ := strconv.Atoi(p.lastMatch[p.lastRegexMap["index"]])
		return v
	}
	return -1
}

// GetLastKey retrieves the string representing the key that we want to index from last regex check
func (p Prefix) GetLastKey() string {
	return p.lastMatch[p.lastRegexMap["key"]]
}

// HasIndex checks whether the key has a prefix with the regex for index by default (\d+)key
func (p *Prefix) HasIndex(other string) bool {
	return p.hasRegex(other, p.indexRegex)
}

// HasIndexAppend checks whether the key has a prefix with the regex for index by default (\d++)key
func (p *Prefix) HasIndexAppend(other string) bool {
	return p.hasRegex(other, p.indexRegexAppend)
}

func (p *Prefix) HasIndexAppendAll(other string) bool {
	return p.hasRegex(other, p.indexRegexAppendAll)
}
