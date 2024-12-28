package arrayutil

func Contains[T comparable](slice []T, target T) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}

func IndexOf[T comparable](slice []T, target T) int {
	for i, value := range slice {
		if value == target {
			return i
		}
	}
	return -1
}

func RemoveAt[T any](slice []T, pos int) []T {
	return append(slice[:pos], slice[pos+1:]...)
}

func RemoveObject[T comparable](slice []T, target T) ([]T, bool) {
	pos := IndexOf(slice, target)
	if pos >= 0 {
		return RemoveAt(slice, pos), true
	} else {
		return slice, false
	}
}

func CopyArray[T any](slice []T) []T {
	destArray := make([]T, len(slice))
	copy(destArray[:], slice[:])
	return destArray
}

func MakeSequence(start int, n int) []int {
	// Make slice which length is zero and capacity is n
	seq := make([]int, 0, n)

	for i := start; i <= n; i++ {
		seq = append(seq, i)
	}
	return seq
}
