package gif

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"strconv"
	"strings"

	math2 "github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/imgfactory"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/saucenao"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name:  "GIF",
	Brief: "gif加速减速等",
	Usage: `用法：
	机器人可以从以下方式获取图片：附带一张图，回复一张图，下一条消息发一张图
	{cmd}gif 效果1 效果2 效果3 效果4....
效果：
	上滑，下滑，左滑，右滑：如其名
	滚，反滚：图片转一圈
	滚1，滚2，反滚1，反滚2：同上但是制定圈数
	间隔2，间隔3，间隔4：GIF播放间隔，单位10毫秒，最小为2
	重绘：如果GIF有重叠请加上这个
备注：
	图片太大会被裁切（每个边最大300像素）
`,
	SuperUsage: ``,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"gif"}).Handle(gifHandle)
}

func gifHandle(ctx *zero.Ctx) {
	if !saucenao.MustProvidePicture(ctx) {
		ctx.SendChain(message.Text("无法获取图片"))
		utils.SetNotStatistic(ctx)
		return
	}
	args := strings.Split(utils.GetArgs(ctx), " ")
	imageURL := ctx.State["image_url"].([]string)[0]
	g, err := readGIF(imageURL)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		utils.SetNotStatistic(ctx)
		return
	}
	if len(args) == 0 {
		ctx.SendChain(message.Text("ERROR: 参数为空"))
		utils.SetNotStatistic(ctx)
		return
	}
	frames := SplitAnimatedGIF(g, strings.Contains(utils.GetArgs(ctx), "重绘"))
	var delay = 4
	for _, arg := range args {
		println(arg)
		if strings.HasPrefix(arg, "间隔") {
			var err error
			delay, err = strconv.Atoi(arg[len("间隔"):])
			if err != nil {
				ctx.Send("参数不对哦，可以看看帮助,ERROR:" + err.Error())
				utils.SetNotStatistic(ctx)
				return
			}
			delay = math2.Max(2, delay)
		} else if arg == "上滑" {
			frames = slide(frames, "up")
		} else if arg == "下滑" {
			frames = slide(frames, "down")
		} else if arg == "左滑" {
			frames = slide(frames, "left")
		} else if arg == "右滑" {
			frames = slide(frames, "right")
		} else if strings.HasPrefix(arg, "滚") {
			loops, err := strconv.Atoi(arg[len("滚"):])
			if err != nil {
				loops = 1
			}
			frames = roll(frames, loops)
		} else if strings.HasPrefix(arg, "反滚") {
			loops, err := strconv.Atoi(arg[len("反滚"):])
			if err != nil {
				loops = 1
			}
			frames = roll(frames, -loops)
		}

	}
	buf := bytes.NewBuffer(nil)
	g = imgfactory.MergeGif(delay, frames)
	err = gif.EncodeAll(buf, g)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		utils.SetNotStatistic(ctx)
	}
	ctx.SendChain(message.ImageBytes(buf.Bytes()))

}

func readGIF(URL string) (*gif.GIF, error) {
	retry, err := client.GetBytesRetry(URL, 3)
	if err != nil {
		return nil, err
	}
	mime := mimetype.Detect(retry)
	if strings.HasPrefix(mime.String(), "image/gif") {
		return gif.DecodeAll(bytes.NewReader(retry))
	} else if strings.HasPrefix(mime.String(), "image") {
		decode, _, err := image.Decode(bytes.NewReader(retry))
		if err != nil {
			return nil, err
		}
		var frames []*image.NRGBA
		for i := 0; i < 20; i++ {
			nim := image.NewNRGBA(decode.Bounds())
			draw.Draw(nim, nim.Bounds(), decode, image.ZP, draw.Src)
			frames = append(frames, nim)
		}
		return imgfactory.MergeGif(5, frames), nil
	} else {
		return nil, fmt.Errorf("unknown type: %s", mime.String())
	}
}

func SplitAnimatedGIF(g *gif.GIF, alwaysReset bool) (frames []*image.NRGBA) {
	imgWidth, imgHeight := GetGifDimensions(g)
	overpaintImage := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), g.Image[0], image.ZP, draw.Src)
	for _, srcImg := range g.Image {
		if alwaysReset {
			draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Src)
		} else {
			draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)
		}
		nim := image.NewNRGBA(srcImg.Bounds())
		draw.Draw(nim, nim.Bounds(), overpaintImage, image.ZP, draw.Src)
		frames = append(frames, nim)
	}
	for i := range frames {
		limit := 300
		if frames[i].Bounds().Dx() > limit || frames[i].Bounds().Dy() > limit {
			if frames[i].Bounds().Dx() > frames[i].Bounds().Dy() {
				frames[i] = imaging.Resize(frames[i], limit, 0, imaging.Lanczos)
			} else {
				frames[i] = imaging.Resize(frames[i], 0, limit, imaging.Lanczos)
			}
		}
	}
	return
}

func GetGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}
