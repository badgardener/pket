package core

type Callback interface {
	Log(msg string)
	Info(msg string)

	Success(msg string)

	Warn(msg string)
	Error(msg string)

	Prompt(msg string, def bool) bool
}
