package sdk

type Scheduler interface {
	Add(key KeyInfo)
	GetChan() *chan KeyInfo
}
