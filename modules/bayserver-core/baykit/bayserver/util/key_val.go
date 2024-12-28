package util

type KeyVal struct {
	Name  string
	Value string
}

func NewKeyVal(name string, value string) *KeyVal {
	return &KeyVal{
		Name:  name,
		Value: value,
	}
}
