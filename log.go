package my

import (
	"fmt"
	"os"
	"time"
)

func Log(vars ...any) {
	trace := Trace{}.New().SkipFile(1)

	f, _ := os.OpenFile(
		getProjectRoot(trace[0].File) + "my.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	for _, v := range vars {
		_, errWrite := f.Write([]byte(fmt.Sprintf(
			"%s %s: %s\n",
			time.Now().Format(time.DateTime),
			trace.Local()[0].String(),
			formatter{}.New().Format(v),
		)))
		PanicIf(errWrite)
	}

    Must(f.Close())
}
