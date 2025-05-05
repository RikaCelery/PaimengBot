package gif

import (
	"image"

	"golang.org/x/image/draw"

	"github.com/disintegration/imaging"
)

type slideDirection string

const (
	SlideUp    slideDirection = "up"
	SlideDown  slideDirection = "down"
	SlideLeft  slideDirection = "left"
	SlideRight slideDirection = "right"
)

func extend(in *image.NRGBA, expend float64, mirror bool) *image.NRGBA {
	w, h := in.Bounds().Dx(), in.Bounds().Dy()
	newW := int(float64(w) * expend)
	newH := int(float64(h) * expend)

	if newW <= w && newH <= h {
		return in // 不扩展
	}

	out := image.NewNRGBA(image.Rect(0, 0, newW, newH))
	offsetX := (newW - w) / 2
	offsetY := (newH - h) / 2

	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			// 计算相对坐标（以中心为基准）
			dx := x - offsetX
			dy := y - offsetY

			// 将坐标映射到原图内（平铺 or 镜像）
			srcX := mod(dx, w)
			srcY := mod(dy, h)

			if mirror {
				// 镜像翻转逻辑
				flipX := (dx/w)%2 != 0
				flipY := (dy/h)%2 != 0

				if flipX {
					srcX = w - 1 - srcX
				}
				if flipY {
					srcY = h - 1 - srcY
				}
			}

			out.Set(x, y, in.NRGBAAt(srcX, srcY))
		}
	}

	return out
}

// 取模，确保总是非负（支持负数输入）
func mod(a, b int) int {
	r := a % b
	if r < 0 {
		r += b
	}
	return r
}
func slide(frames []*image.NRGBA, direction slideDirection) (ret []*image.NRGBA) {
	totalFrames := len(frames) // 获取总帧数
	for i, frame := range frames {
		ori := imaging.Clone(frame)
		w, h := frame.Bounds().Dx(), frame.Bounds().Dy()
		frame = image.NewNRGBA(image.Rect(0, 0, w, h))
		// 计算当前帧的位移量（基于总帧数和图像尺寸）
		var offset int
		switch direction {
		case SlideUp, SlideDown:
			offset = (i * h) / totalFrames // 垂直方向位移量
		case SlideLeft, SlideRight:
			offset = (i * w) / totalFrames // 水平方向位移量
		}

		// 根据方向应用位移
		switch direction {
		case SlideUp:
			draw.Copy(frame, image.Point{0, -offset}, ori, ori.Bounds(), draw.Src, nil)
			draw.Copy(frame, image.Point{0, h - offset}, ori, ori.Bounds(), draw.Over, nil)
		case SlideDown:
			draw.Copy(frame, image.Point{0, offset}, ori, ori.Bounds(), draw.Src, nil)
			draw.Copy(frame, image.Point{0, -h + offset}, ori, ori.Bounds(), draw.Over, nil)
		case SlideLeft:
			draw.Copy(frame, image.Point{-offset, 0}, ori, ori.Bounds(), draw.Src, nil)
			draw.Copy(frame, image.Point{w - offset, 0}, ori, ori.Bounds(), draw.Over, nil)
		case SlideRight:
			draw.Copy(frame, image.Point{offset, 0}, ori, ori.Bounds(), draw.Src, nil)
			draw.Copy(frame, image.Point{-w + offset, 0}, ori, ori.Bounds(), draw.Over, nil)
		}
		ret = append(ret, frame)
	}
	return
}

func roll(frames []*image.NRGBA, loops int) (ret []*image.NRGBA) {
	var angle = float64(360.0*loops) / float64(len(frames))
	for i, frame := range frames {
		a := angle * float64(i)
		nrgba := extend(frame, 1.8, true)
		nrgba = imaging.Rotate(nrgba, a, image.Black)
		// center crop
		nrgba = imaging.CropCenter(nrgba, frame.Bounds().Dx(), frame.Bounds().Dy())
		ret = append(ret, nrgba)
	}
	return
}

func rotate(frames []*image.NRGBA, angle float64) (ret []*image.NRGBA) {
	for _, frame := range frames {
		nrgba := extend(frame, 1.8, true)
		nrgba = imaging.Rotate(nrgba, angle, image.Black)
		// center crop
		nrgba = imaging.CropCenter(nrgba, frame.Bounds().Dx(), frame.Bounds().Dy())
		ret = append(ret, nrgba)
	}
	return
}
