package js_engine

import "fmt"

func (j *JS_ENGINE) print(c ...interface{}) {
	fmt.Fprint(j.IO_OUT, c...)
}

func (j *JS_ENGINE) ConfPrintErrors(v bool) {
	j.PrintErrors = v
}

/* SendErrors if called, errors will be printed to IO_OUT */
func (j *JS_ENGINE) SendErrors() {
	if !j.PrintErrors {
		return
	}
	if len(j.Errors) > 0 {
		for _, e := range j.Errors {
			j.print(e)
			j.print("\n")
		}
	}
}
