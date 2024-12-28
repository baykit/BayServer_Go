package maputil

func CopyMap[T comparable, U any](srcMap map[T]U) map[T]U {
	dstMap := make(map[T]U)

	for key, value := range srcMap {
		dstMap[key] = value
	}

	return dstMap
}
