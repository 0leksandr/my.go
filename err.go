package my

func PanicIf(err error) {
	if err != nil {
		Dump(err)
		//log.Fatal(err)
		panic(err)
	}
}
func Must(err error) {
	PanicIf(err)
}
