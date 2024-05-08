package utils

func CalcSum(a []int32) int32 {
	sum := int32(0)
	for _, v := range a {
		sum += v
	}
	return sum
}
