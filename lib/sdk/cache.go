package sdk

import "sync/atomic"

type KeyInfo struct {
	Expires int64
	Key     string
}

type Record struct {
	recId   uint64
	Expires int64
	Value   []byte
}

type Cache interface {
	Insert(key KeyInfo, rec Record)
	Lookup(key KeyInfo) (Record, bool)
	Delete(key KeyInfo)
}

var (
	currRecId uint64 = 0
)

// TODO remove unnessasery copy of []bytes here
func NewRecord(expires int64, value []byte) *Record {

	return &Record{
		recId:   atomic.AddUint64(&currRecId, 1),
		Expires: expires,
		Value:   value,
	}
}

func (rec *Record) GetRecId() uint64 {
	return rec.recId
}

func LatestRecordId() uint64 {
	return atomic.LoadUint64(&currRecId)
}
