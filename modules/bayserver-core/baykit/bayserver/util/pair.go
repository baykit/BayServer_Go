package util

type Pair[TA any, TB any] struct {
	A TA
	B TB
}

func NewPair[TA any, TB any](a TA, b TB) *Pair[TA, TB] {
	return &Pair[TA, TB]{A: a, B: b}
}
