package runner

type Runner interface {
	Run() error
}

type Continue struct {
	cnf config
}

type Help struct {
	err error
}

type config struct {
	drivers []string
}
