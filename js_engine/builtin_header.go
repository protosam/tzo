package js_engine

import "net/http"

func (j *JS_ENGINE) header(n, v string) {
	w, ok := j.IO_OUT.(http.ResponseWriter)
	if !ok {
		j.Errors = append(j.Errors, "header(): is not available in CLI mode.")
		j.VM.Interrupt("halt")
		return
	}
	w.Header().Set(n, v)
}

func (j *JS_ENGINE) http_status_code(status_code int) {
	if r, ok := j.IO_OUT.(http.ResponseWriter); ok {
		r.WriteHeader(status_code)
	}
}
