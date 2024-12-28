package common

type Vehicle interface {
	Id() int
	Run()
	OnTimer()
}
