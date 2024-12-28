package bcf

type BcfKeyVal struct {
	BcfObjectStruct
	Key   string
	Value string
}

func NewBcfKeyVal(key string, val string, fileName string, lineNo int) *BcfKeyVal {
	return &BcfKeyVal{BcfObjectStruct{fileName, lineNo}, key, val}
}
