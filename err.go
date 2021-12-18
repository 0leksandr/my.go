package my

func panicIf(err error) {
	if err != nil {
		dumpAt(2, err)
		//log.Fatal(err)
		panic(err)
	}
}
func PanicIf(err error) {
	panicIf(err)
}
func Must(err error) {
	panicIf(err)
}
