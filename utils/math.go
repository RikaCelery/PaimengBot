package utils

type num interface {
	int | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64
}

func Clamp[T num](x, min, max T) T {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func Remap[T num](x, inMin, inMax, outMin, outMax T) T {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}
