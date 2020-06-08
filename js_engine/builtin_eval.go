package js_engine

import "strings"

func (j *JS_ENGINE) Eval(filename, script string) {
	// eat crunchbang
	nl := 0
	if len(script) > 3 && script[0:2] == "#!" {
		nl = strings.Index(script, "\n")
		if nl == -1 {
			nl = 0
		}
	}

	_, err := j.VM.RunScript(filename, script[nl:len(script)])
	if err != nil {
		j.Errors = append(j.Errors, err.Error())
		j.VM.Interrupt("halt")
	}
}
