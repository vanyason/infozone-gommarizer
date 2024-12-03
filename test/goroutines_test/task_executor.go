package main

type TaskExecuter interface {
	Execute() ([]TaskExecuter, error)
}

type TaskExecuterFuncWrap func() ([]TaskExecuter, error)

func (f TaskExecuterFuncWrap) Execute() ([]TaskExecuter, error) {
	return f()
}
