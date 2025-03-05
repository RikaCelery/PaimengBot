package images

// Grid 用于在x起始，长width的区域中，居中均匀排布itemWidth个元素，元素之间间有gap像素间隔
func Grid(x, width float64, items int, itemWidth, gap float64) (xPos []float64) {
	totalUsed := float64(items)*itemWidth + float64(items-1)*gap
	startOffset := (width - totalUsed) / 2
	startX := x + startOffset
	xPos = make([]float64, items)
	for i := range xPos {
		xPos[i] = startX + float64(i)*(itemWidth+gap)
	}
	return xPos
}
