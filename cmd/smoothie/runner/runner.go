package runner

type Runner interface {
	Run() error
}

type config struct {
	drivers []string
}
