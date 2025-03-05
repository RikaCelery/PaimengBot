package images

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/png"
	"sync"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	"github.com/disintegration/imaging"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var defaultFont *truetype.Font

func init() {
	defaultFont = GetDefaultFont()
	if defaultFont != nil {
		log.Infof("成功加载默认字体")
	}
}

// ImageCtx 图片上下文
type ImageCtx struct {
	*gg.Context
}

// NewImageCtxWithBG 创建带有背景图片的图片上下文，通过opacity设置不透明度
func NewImageCtxWithBG(w, h int, bg image.Image, opacity float64) *ImageCtx {
	if opacity > 0 && opacity < 1 {
		bg = AdjustOpacity(bg, opacity)
	}
	res := NewImageCtx(w, h)
	sx := float64(w) / float64(bg.Bounds().Size().X)
	sy := float64(h) / float64(bg.Bounds().Size().Y)
	// 记录原始状态
	res.Push()
	// 设置背景
	res.Scale(sx, sy)
	res.DrawImage(bg, 0, 0)
	// 恢复原始状态
	res.Pop()
	return res
}

// NewImageCtxWithBGRGBA255 以RGBA255形式创建纯色背景的图片上下文
func NewImageCtxWithBGRGBA255(w, h int, r, g, b, a int) *ImageCtx {
	res := NewImageCtx(w, h)
	// 记录原始状态
	res.Push()
	// 设置背景
	res.SetRGBA255(r, g, b, a)
	res.Clear()
	// 恢复原始状态
	res.Pop()
	return res
}

// NewImageCtxWithBGColor 创建纯色背景的图片上下文
func NewImageCtxWithBGColor(w, h int, c string) *ImageCtx {
	res := NewImageCtx(w, h)
	// 记录原始状态
	res.Push()
	// 设置背景
	res.SetColorAuto(c)
	res.Clear()
	// 恢复原始状态
	res.Pop()
	return res
}

// NewImageCtx 创建空图片上下文
func NewImageCtx(w, h int) *ImageCtx {
	dc := gg.NewContext(w, h)
	return &ImageCtx{
		Context: dc,
	}
}

// SetFont 通过truetype.Font设置字体与字体大小
func (img *ImageCtx) SetFont(font *truetype.Font, size float64) error {
	if font == nil {
		return errors.New("point font is nil")
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: size,
	})
	img.SetFontFace(face)
	return nil
}

// UseDefaultFont 使用默认字体并设置字体大小
func (img *ImageCtx) UseDefaultFont(size float64) error {
	return img.SetFont(GetDefaultFont(), size)
}

// UseFont 使用字体并设置字体大小
func (img *ImageCtx) UseFont(path string, size float64) error {
	font, err := ParseFont(path)
	if err != nil {
		return err
	}
	return img.SetFont(font, size)
}

var tempCountMutex sync.Mutex
var tempCount int64 = 0

// GetNewTempSavePath 获取新的临时图片可保存路径名（绝对路径）
func GetNewTempSavePath(prefix string) (string, error) {
	// 获取临时序号
	tempCountMutex.Lock()
	tempCount = (tempCount + 1) % (viper.GetInt64("tmp.maxcount") + 1)
	fileName := fmt.Sprintf("%s_%v.png", prefix, tempCount)
	tempCountMutex.Unlock()

	// 尝试创建临时文件夹
	fullDir, err := utils.MakeDir(consts.TempImageDir)
	if err != nil {
		return "", err
	}
	return utils.PathJoin(fullDir, fileName), nil
}

// SaveTemp 以前缀prefix保存至临时图片文件夹
func (img *ImageCtx) SaveTemp(prefix string) (string, error) {
	path, err := GetNewTempSavePath(prefix)
	if err != nil {
		log.Errorf("创建临时目录或获取绝对路径失败 err：%v", err)
		return path, err
	}
	// 保存图片
	err = img.SavePNG(path)
	if err != nil {
		log.Errorf("保存临时图片失败 err：%v", err)
		return "", err
	}
	return path, nil
}

// SaveTempDefault 以默认前缀(tempimg)保存至临时图片文件夹
func (img *ImageCtx) SaveTempDefault() (string, error) {
	return img.SaveTemp("tempimg")
}

func (img *ImageCtx) GenPNG() ([]byte, error) {
	resultBuff := bytes.NewBuffer(nil) // 结果缓冲区
	// 新建Base64编码器（Base64结果写入结果缓冲区resultBuff）
	// 将图片PNG格式写入Base64编码器
	err := png.Encode(resultBuff, img.Image())
	if err != nil {
		return nil, err
	}
	return resultBuff.Bytes(), nil
}
func (img *ImageCtx) GenMessageBase64() (message.Segment, error) {
	resultBuff := bytes.NewBuffer(nil) // 结果缓冲区
	// 新建Base64编码器（Base64结果写入结果缓冲区resultBuff）
	encoder := base64.NewEncoder(base64.StdEncoding, resultBuff)
	// 将图片PNG格式写入Base64编码器
	err := png.Encode(encoder, img.Image())
	if err != nil {
		_ = encoder.Close()
		return message.Text("图片生成失败"), err
	}
	// 结束Base64编码
	err = encoder.Close()
	if err != nil {
		return message.Text("图片Base64生成失败"), err
	}
	return message.Image("base64://" + resultBuff.String()), nil
}

// GenMessageAuto 自动生成ZeroBot图片消息
func (img *ImageCtx) GenMessageAuto() (message.Segment, error) {
	// 消息收发端不在本地
	if !utils.IsOneBotLocal() {
		return img.GenMessageBase64()
	}
	// 消息收发端位于本地
	file, err := img.SaveTempDefault()
	if err != nil { // 生成文件出错，尝试Base64
		return img.GenMessageBase64()
	}
	return message.Image("file:///" + file), nil
}

// 扩展绘画方法

func (img *ImageCtx) DrawRoundedRectangle4(x, y, w, h, lt, rt, lb, rb float64) {
	img.NewSubPath()
	img.MoveTo(x+lt, y)
	// 上面
	img.LineTo(x+w-lt-rt, y)
	// 右上角
	img.DrawArc(x+w-lt, y+rt, rt, gg.Radians(270), gg.Radians(360))
	// 右面
	img.LineTo(x+w, y+h-rb)
	// 右下角
	img.DrawArc(x+w-rb, y+h-rb, rb, gg.Radians(0), gg.Radians(90))
	// 下面
	img.LineTo(x+lb, y+h)
	// 左下角
	img.DrawArc(x+lb, y+h-lb, lb, gg.Radians(90), gg.Radians(180))
	// 左面
	img.LineTo(x, y+lt)
	img.DrawArc(x+lt, y+lt, lt, gg.Radians(180), gg.Radians(270))
	img.ClosePath()
}

func (img *ImageCtx) InnerShadow(c string, x, y int, w, h, size float64) {
	extend := 40.0
	innerStroke := NewImageCtx(int(w+extend), int(h+extend))
	innerStroke.SetColorAuto(c)
	innerStroke.DrawRectangle(extend/2, extend/2, w, h)
	innerStroke.Clip()
	innerStroke.InvertMask()
	innerStroke.DrawRectangle(0, 0, w+extend, h+extend)
	innerStroke.Fill()
	blurred := imaging.Blur(innerStroke.Image(), size)
	innerShadow := NewImageCtx(int(w), int(h))
	innerShadow.DrawRectangle(0, 0, w, h)
	innerShadow.Clip()
	innerShadow.DrawImage(blurred, int(-extend/2), int(-extend/2))
	img.DrawImage(innerShadow.Image(), x, y)
}
func (img *ImageCtx) InnerShadowRounded(c string, x, y int, w, h, r, size float64) {
	extend := 40.0
	innerStroke := NewImageCtx(int(w+extend), int(h+extend))
	innerStroke.SetColorAuto(c)
	innerStroke.DrawRoundedRectangle(extend/2, extend/2, w, h, r)
	innerStroke.Clip()
	innerStroke.InvertMask()
	innerStroke.DrawRectangle(0, 0, w+extend, h+extend)
	innerStroke.Fill()
	blurred := imaging.Blur(innerStroke.Image(), size)
	innerShadow := NewImageCtx(int(w), int(h))
	innerShadow.DrawRoundedRectangle(0, 0, w, h, r)
	innerShadow.Clip()
	innerShadow.DrawImage(blurred, int(-extend/2), int(-extend/2))
	img.DrawImage(innerShadow.Image(), x, y)
}

func (img *ImageCtx) InnerRoundedShadow4(c string, x, y int, w, h, size, lt, rt, lb, rb float64) {
	extend := 40.0
	innerStroke := NewImageCtx(int(w+extend), int(h+extend))
	innerStroke.SetColorAuto(c)
	innerStroke.DrawRoundedRectangle4(extend/2, extend/2, w, h, lt, rt, lb, rb)
	innerStroke.Clip()
	innerStroke.InvertMask()
	innerStroke.DrawRectangle(0, 0, w+extend, h+extend)
	innerStroke.Fill()
	blurred := imaging.Blur(innerStroke.Image(), size)
	innerShadow := NewImageCtx(int(w), int(h))
	innerShadow.DrawRoundedRectangle4(0, 0, w, h, lt, rt, lb, rb)
	innerShadow.Clip()
	innerShadow.DrawImage(blurred, int(-extend/2), int(-extend/2))
	img.DrawImage(innerShadow.Image(), x, y)
}

func (img *ImageCtx) ShadowRounded(c string, x, y int, w, h, r, size, shirnk, offX, offY float64) {
	extend := 40.0
	innerStroke := NewImageCtx(int(w+extend), int(h+extend))
	innerStroke.SetColorAuto(c)
	innerStroke.DrawRoundedRectangle(extend/2+shirnk, extend/2+shirnk, w-shirnk*2, h-shirnk*2, r)
	innerStroke.Fill()
	blurred := imaging.Blur(innerStroke.Image(), size)
	img.DrawImage(blurred, x-int(extend/2-offX-shirnk/2), y-int(extend/2-offY-shirnk/2))

}
