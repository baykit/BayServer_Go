package huffman

type HNode struct {
	Value int //  if vlaue > 0 leaf node else inter node
	One   *HNode
	Zero  *HNode
}

func NewHNode() *HNode {
	return &HNode{
		Value: -1,
	}
}
