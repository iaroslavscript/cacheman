package sdk

type LogInfo struct {
	Id   int64
	Time int64
}

type ReplItem struct {
	Action int8
	Key    KeyInfo
	Value  Record
}

type ReplLog struct {
	Info LogInfo
	Data []ReplItem
}

type Replication interface {
	Add(item ReplItem)
}

// TODO remove unnessasery copy of []bytes here
func NewReplItem(action int8, key KeyInfo, value Record) *ReplItem {
	return &ReplItem{
		Action: action,
		Key:    key,
		Value:  value,
	}
}
