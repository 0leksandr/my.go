package my

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"regexp"
	"strings"
)

func Dump2(values ...interface{}) {
	spew.Config.DisableCapacities = true
	spew.Config.Indent = "    "

	for _, val := range values {
		str := spew.Sdump(val)
		lines := strings.Split(str, "\n")
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
		fmt.Printf("%v\n", fileLine(1))
		for _, line := range lines { fmt.Println(line) }
	}
}
