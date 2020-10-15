package common

type Scheduler interface {
	Add(key KeyInfo)
	GetChan() *chan KeyInfo
}
