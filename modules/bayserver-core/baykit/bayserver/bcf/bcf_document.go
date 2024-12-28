package bcf

import (
	"fmt"
)

type BcfDocument struct {
	ContentList []interface{} // []BcfObject
}

func NewBcfDocument() BcfDocument {
	return BcfDocument{[]interface{}{}}
}

func (doc *BcfDocument) Print() {
	doc.printContentList(doc.ContentList, 0)
}

func (doc *BcfDocument) printContentList(list []interface{}, indent int) {
	for _, o := range list {
		doc.printIndent(indent)
		if elm, ok := o.(*BcfElement); ok {
			fmt.Printf("Element(%s, %s){", elm.Name, elm.Arg)
			doc.printContentList(elm.ContentList, indent+1)
			doc.printIndent(indent)
			fmt.Printf("}\n")

		} else {
			kv, _ := o.(*BcfKeyVal)
			doc.printKeyVal(kv)
			fmt.Printf("\n")
		}
	}
}

func (doc *BcfDocument) printKeyVal(kv *BcfKeyVal) {
	fmt.Printf("KeyVal(%s=%s)", kv.Key, kv.Value)
}

func (doc *BcfDocument) printIndent(indent int) {
	for i := 0; i < indent; i++ {
		fmt.Printf(" ")
	}
}
