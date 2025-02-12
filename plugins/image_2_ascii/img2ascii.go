package image_2_ascii

import (
	"fmt"
	"image"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils/consts"
	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/fogleman/gg"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func image2ascii(img image.Image, chars int, charset []string) string {
	if len(charset) == 0 {
		charset = []string{"@", "#", "S", "%", "?", "*", "+", ";", ":", ",", ".", " "}
	}
	bounds := img.Bounds()
	dx, dy := bounds.Dx(), bounds.Dy()

	// 计算输出尺寸（考虑终端字符高宽比）
	outWidth := chars
	outHeight := int(float64(outWidth) * float64(dy) / float64(dx) * 0.6)
	if outHeight < 1 {
		outHeight = 1
	}

	// 预处理步骤：生成灰度网格
	grid := make([][]int, outHeight)
	for y := 0; y < outHeight; y++ {
		grid[y] = make([]int, outWidth)
		for x := 0; x < outWidth; x++ {
			xStart := bounds.Min.X + (x*dx)/outWidth
			xEnd := bounds.Min.X + ((x+1)*dx)/outWidth
			yStart := bounds.Min.Y + (y*dy)/outHeight
			yEnd := bounds.Min.Y + ((y+1)*dy)/outHeight

			sum, count := 0, 0
			for iy := yStart; iy < yEnd; iy++ {
				for ix := xStart; ix < xEnd; ix++ {
					r, g, b, _ := img.At(ix, iy).RGBA()
					gray := int(0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))
					sum += gray
					count++
				}
			}
			if count > 0 {
				grid[y][x] = sum / count
			}
		}
	}

	// 对比度增强
	minGray, maxGray := 255, 0
	for _, row := range grid {
		for _, val := range row {
			if val < minGray {
				minGray = val
			}
			if val > maxGray {
				maxGray = val
			}
		}
	}

	// 构建ASCII字符串
	var builder strings.Builder
	for y := 0; y < outHeight; y++ {
		for x := 0; x < outWidth; x++ {
			normalized := grid[y][x]
			if maxGray != minGray {
				normalized = (normalized - minGray) * 255 / (maxGray - minGray)
			}

			index := normalized * (len(charset) - 1) / 255
			if index >= len(charset) {
				index = len(charset) - 1
			}
			builder.WriteString(charset[index])
		}
		builder.WriteByte('\n')
	}

	return builder.String()
}

func measureString(str string, fontSize, lineSpace float64) (width float64, height float64, err error) {
	img := images.NewImageCtx(1, 1)
	font, err := images.ParseFont(consts.DefaultMonoTTFPath)
	if err != nil {
		return 0, 0, err
	}
	_ = img.SetFont(font, fontSize)
	w, h := img.MeasureMultilineString(str, lineSpace)
	return w, h, nil
}

func draw(str string) (message.Segment, error) {
	// 计算图片大小并初始化
	fontSize, lineSpace := 12.0, 1.0
	w, h, err := measureString(str, fontSize, lineSpace)
	if err != nil {
		return message.Segment{}, fmt.Errorf("formSingleHelpMsg measureString err: %v", err)
	}
	w, h = w+20, h+50+10
	img := images.NewImageCtxWithBGRGBA255(int(w), int(h), 255, 255, 255, 255)
	// 名称
	img.Push()
	defer img.Pop()
	parseFont, err := images.ParseFont(consts.DefaultMonoTTFPath)
	if err != nil {
		return message.Segment{}, fmt.Errorf("formSingleHelpMsg img err: %v", err)
	}
	err = img.SetFont(parseFont, fontSize)
	if err != nil {
		return message.Segment{}, fmt.Errorf("formSingleHelpMsg img err: %v", err)
	}
	img.SetRGB(0, 0, 0) // 纯黑色
	img.DrawStringWrapped(str, 10, 10, 0, 0, w, lineSpace, gg.AlignLeft)

	// 生成回包
	msg, err := img.GenMessageAuto()
	if err != nil {
		return message.Segment{}, fmt.Errorf("formSingleHelpMsg img err: %v", err)
	}
	return msg, nil
}
