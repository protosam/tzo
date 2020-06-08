package js_engine

func (j *JS_ENGINE) die() {
	j.VM.Interrupt("die()")
}
