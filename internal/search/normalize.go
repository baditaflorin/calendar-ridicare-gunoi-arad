package search

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	wordNoise = regexp.MustCompile(`\b(strada|str|calea|bulevardul|bd|piata|pta|aleea|cartier|cartierul|municipiul|arad)\b`)
	spaces    = regexp.MustCompile(`\s+`)
)

func Normalize(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = strings.ReplaceAll(s, "ș", "s")
	s = strings.ReplaceAll(s, "ş", "s")
	s = strings.ReplaceAll(s, "ț", "t")
	s = strings.ReplaceAll(s, "ţ", "t")
	s = strings.ReplaceAll(s, "ă", "a")
	s = strings.ReplaceAll(s, "â", "a")
	s = strings.ReplaceAll(s, "î", "i")
	t := norm.NFKD.String(s)
	var b strings.Builder
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		b.WriteByte(' ')
	}
	out := wordNoise.ReplaceAllString(b.String(), " ")
	return spaces.ReplaceAllString(strings.TrimSpace(out), " ")
}

func Aliases(cartier, street string) []string {
	base := Normalize(street)
	withCartier := strings.TrimSpace(Normalize(cartier + " " + street))
	aliases := []string{base}
	if withCartier != "" && withCartier != base {
		aliases = append(aliases, withCartier)
	}
	if strings.Contains(base, "densuseanu") {
		aliases = append(aliases, strings.ReplaceAll(base, "densuseanu", "densusianu"))
	}
	return unique(aliases)
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := values[:0]
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
