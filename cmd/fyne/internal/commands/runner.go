package commands

type runner interface {
	runOutput(arg ...string) ([]byte, error)
	setDir(dir string)
	setEnv(env []string)
}
