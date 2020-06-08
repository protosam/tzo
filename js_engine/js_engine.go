package js_engine

import (
	"io"

	"github.com/dop251/goja"
)

type JS_ENGINE struct {
	VM          *goja.Runtime
	Errors      []string
	IO_OUT      io.Writer
	JailDir     string
	PrintErrors bool
}

/* SetIOWriter allows developer to change the IO out for use with
CLI or for websites. */
func (j *JS_ENGINE) SetIOWriter(io_out io.Writer) {
	j.IO_OUT = io_out
}
