package js_engine

import (
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)

// New will return a fresh JS_ENGINE{} to run javascript.
func New(jaildir string) *JS_ENGINE {
	j := &JS_ENGINE{}
	jaildir, err := filepath.Abs(jaildir)
	if err != nil {
		j.Errors = append(j.Errors, err.Error())
		return j
	}

	j.VM = goja.New()
	j.IO_OUT = os.Stdout
	j.JailDir = jaildir
	j.PrintErrors = true

	// Disabled because safety?
	//j.VM.Set("eval", j.Eval)
	j.VM.Set("require", j.Require)
	j.VM.Set("print", j.print)
	j.VM.Set("header", j.header)
	j.VM.Set("die", j.die)
	j.VM.Set("http_status_code", j.http_status_code)

	// Objects start with capitol letters
	j.VM.Set("HTTP", j.HTTP)

	// Conf commands
	j.VM.Set("ConfPrintErrors", j.ConfPrintErrors)

	return j
}
