package js_engine

import (
	"io/ioutil"
	"path/filepath"
)

func (j *JS_ENGINE) Require(filename string) {
	filename, err := filepath.Abs(j.JailDir + "/" + filename)
	if err != nil {
		j.Errors = append(j.Errors, "Require(): "+err.Error())
		j.VM.Interrupt("halt")
		return
	}

	inJailDir, err := filepath.Match(j.JailDir+"/*", filename)
	if err != nil {
		j.Errors = append(j.Errors, "Require(): "+err.Error())
		j.VM.Interrupt("halt")
		return
	}

	if !inJailDir {
		j.Errors = append(j.Errors, "Require(): "+filename+" is outside of the jail directory.")
		j.VM.Interrupt("halt")
		return
	}

	s, err := ioutil.ReadFile(filename)
	if err != nil {
		j.Errors = append(j.Errors, "Require(): "+err.Error())
		j.VM.Interrupt("halt")
		return
	}
	j.Eval(filename, string(s))
}
