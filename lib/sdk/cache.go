package sdk

type KeyInfo struct {
	Expires int64
	Key     string
}

type Record struct {
	Expires int64
	Value   []byte
}

type Cache interface {
	Insert(key KeyInfo, rec Record)
	Lookup(key KeyInfo) (Record, bool)
	Delete(key KeyInfo)
}
