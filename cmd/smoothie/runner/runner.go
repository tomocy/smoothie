package runner

type Runner interface {
	Run() error
}

type Continue struct {
	cnf config
}

type config struct {
	drivers []string
}
