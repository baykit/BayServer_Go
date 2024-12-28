package bcf

import "strings"

type BcfElement struct {
	BcfObjectStruct
	Name        string
	Arg         string
	ContentList []interface{} // []BcfObject
}

func NewBcfElement(name string, arg string, fileName string, lineNo int) *BcfElement {
	return &BcfElement{
		BcfObjectStruct{fileName, lineNo},
		name,
		arg,
		[]interface{}{}}
}

func (b *BcfElement) GetValue(key string) string {
	for _, o := range b.ContentList {

		if kv, ok := o.(*BcfKeyVal); ok {
			if strings.EqualFold(kv.Key, key) {
				return kv.Value
			}
		}
	}
	return ""
}
