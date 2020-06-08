package config

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}
