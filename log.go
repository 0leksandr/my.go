package my

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"
)

var mx sync.Mutex

func Log(vars ...any) {
	mx.Lock()
	defer mx.Unlock()

	trace := Trace{}.New().SkipFile(1)
	root := getProjectRoot(trace[0].File)
	logFilename := root + "my_log.go"

	logFile := MustFirst(os.OpenFile(
		logFilename,
		os.O_CREATE|os.O_WRONLY,
		0644,
	))

	readFirstLine := func(filename string) string {
		file := MustFirst(os.OpenFile(filename, os.O_RDONLY, 0))
		defer func() {
			Must(file.Close())
		}()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			return scanner.Text()
		} else {
			return ""
		}
	}

	regexPackage := regexp.MustCompile("^package \\w+$")
	if !regexPackage.MatchString(readFirstLine(logFilename)) {
		for _, siblingFile := range MustFirst(os.ReadDir(root)) {
			if !siblingFile.IsDir() {
				if regexp.MustCompile(".+\\.go$").MatchString(siblingFile.Name()) {
					firstLine := readFirstLine(root + siblingFile.Name())
					if regexPackage.MatchString(firstLine) {
						MustFirst(logFile.WriteAt([]byte(firstLine), 0))
						break
					}
				}
			}
		}
	}

	for _, v := range vars {
		MustFirst(logFile.WriteAt(
			[]byte(fmt.Sprintf(
				"// %s %s\nvar _= %s\n",
				time.Now().Format(time.DateTime),
				trace.Local()[0].String(),
				formatter{}.New().Format(v),
			)),
			MustFirst(logFile.Stat()).Size(),
		))
	}

    Must(logFile.Close())
}
