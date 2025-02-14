package image_2_ascii

import (
	"fmt"
	"image"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/extension/shell"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "字符画",
	Usage: `把图片变成字符画
用法：
	{cmd}字符画[图片] [参数]?：将一个图片变为字符画图片
	[回复某个消息]{cmd}字符画 [参数]?
参数:
	-w 40 字符分辨率（一行多少字符）为40
	-S Abcde_. 使用的字符集（黑色到白色依次使用里面的字符）
`,
	Classify: "图片相关",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	re := regexp.MustCompile("(?:字符画|ascii)(.*)")
	proxy.OnMessage(zero.NewPattern(nil).Reply().SetOptional().Add("text", false, func(msg *message.Segment) zero.PatternParsed {
		s := msg.Data["text"]
		s = strings.Trim(s, " \n\r\t")
		if strings.HasPrefix(s, zero.BotConfig.CommandPrefix) {
			s = strings.TrimPrefix(s, zero.BotConfig.CommandPrefix)
		}
		matchString := re.MatchString(s)
		if matchString {
			return zero.PatternParsed{
				Value: re.FindStringSubmatch(s),
				Msg:   msg,
			}
		}
		return zero.PatternParsed{}
	}).AsRule(), func(ctx *zero.Ctx) bool {
		var img *message.Segment
		parsed := extension.PatternModel{}
		_ = ctx.Parse(&parsed)
		if parsed.Matched[0].Raw() != nil {
			msg := ctx.GetMessage(parsed.Matched[0].Reply())
			for _, element := range msg.Elements {
				fmt.Println(element)
				if element.Type == "image" {
					img = &element
					break
				}
			}
		}
		for _, element := range ctx.Event.Message {
			if element.Type == "image" {
				img = &element
				break
			}
		}
		if img == nil {
			return false
		}

		ctx.State["img"] = img.Data["url"]
		// shell 解析
		args := shell.Parse(parsed.Matched[1].Text()[1])
		chars := ""
		width := int64(0)
		for i, arg := range args {
			if strings.HasPrefix(arg, "-w") && len(args) > i+1 {
				var err error
				width, err = strconv.ParseInt(args[i+1], 10, 64)
				if err != nil {
					return false

				}
			}
			if strings.HasPrefix(arg, "-S") && len(args) > i+1 {
				chars = args[i+1]
			}
		}
		ctx.State["chars"] = chars
		ctx.State["width"] = width

		return true
	}).Handle(handleAscii)
}

func handleAscii(ctx *zero.Ctx) {
	img := ctx.State["img"].(string)
	chars := ctx.State["chars"].(string)
	width := ctx.State["width"].(int64)
	resp, err := http.Get(img)
	if err != nil {
		log.Errorf("<image_2_ascii>无法下载图片: %v", err)
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	defer resp.Body.Close()
	decode, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Errorf("<image_2_ascii>无法加载图片: %v", err)
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	if width == 0 && decode.Bounds().Dx() > 300 {
		// resize
		decode = resize.Resize(300, 0, decode, resize.Lanczos3)
	} else if width > 0 {
		if width > 3000 {
			ctx.Send("宽度太大了！")
			return
		}
		decode = resize.Resize(uint(width), 0, decode, resize.Lanczos3)
	}
	segment, err := draw(image2ascii(decode, decode.Bounds().Dx(), strings.Split(chars, "")))
	if err != nil {
		log.Errorf("<image_2_ascii>无法生成字符画: %v", err)
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(segment)
}
