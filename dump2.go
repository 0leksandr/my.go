package my

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"regexp"
	"strings"
)

func Sdump2(value any) string {
	spew.Config.DisableCapacities = true
	spew.Config.Indent = "    "

	str := spew.Sdump(value)
	lines := strings.Split(str, "\n")
	if len(lines) > 1 { lines = lines[:len(lines)-1] }
	for lineNr, line := range lines {
		type Replacement struct {
			re *regexp.Regexp
			to string
		}
		for _, replacement := range []Replacement{
			{ // opening bracket
				regexp.MustCompile("^( *(?:\\w+: )?)\\((\\[])?(?:main\\.)?([^)]+)\\)(?: \\(len=\\d+\\))? {$"),
				"$1$2$3{",
			},
			{ // simple value
				regexp.MustCompile("^( *\\w+: )\\([\\w.]+\\) (.*[^,]),?$"),
				"$1$2,",
			},
			{ // closing bracket
				regexp.MustCompile("^( +})$"),
				"$1,",
			},
		} {
			if replacement.re.MatchString(line) {
				line = replacement.re.ReplaceAllString(line, replacement.to)
			}
		}
		lines[lineNr] = line
	}
	return strings.Join(lines, "\n")
}
func Dump2(values ...any) {
	for _, value := range values {
		fmt.Printf("%v\n%s\n", fileLine(1), Sdump2(value))
	}
}
