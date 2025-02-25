package wordcloud

import (
	"fmt"
	"os"
	"strconv"

	"github.com/RicheyJang/PaimengBot/basic/history"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/go-ego/gse"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	info = manager.PluginInfo{
		Name:     "词云",
		Classify: "其他",
		Usage: `
用法：
	{cmd}词云：生成本群今日词云
	`,
		SuperUsage: `
	{cmd}词云 [群号]：生成指定群今日词云
配置项：
	data/wordcloud/stopwords.txt：停用词文本文件，一行一个
	data/wordcloud/dict.txt：用户词典，一行一个
`,
	}
	proxy *manager.PluginProxy
)

func initSeg() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Error("<wordcloud>初始化失败", err)
		}
	}()
	seg = gse.Segmenter{}
	// 加载默认词典
	// seg.LoadDictEmbed()
	// 加载日文词典
	// seg.LoadDictEmbed("jp")
	seg.LoadStopEmbed()
	if p := proxy.ResolveFile("stopwords.txt"); utils.FileExists(p) {
		seg.LoadStop(p)
		logrus.Info("<wordcloud>加载", len(seg.StopWordMap), "个停用词")
	} else {
		proxy.WriteData([]byte(""), "stopwords.txt")
	}

	if p := proxy.ResolveFile("dict.txt"); utils.FileExists(p) {
		seg.LoadDict(p)
		logrus.Info("<wordcloud>加载", len(seg.Dict.Tokens), "个单词")
	} else {
		proxy.WriteData([]byte(""), "dict.txt")
	}
}
func init() {
	proxy = manager.RegisterPlugin(info)
	go initSeg()
	proxy.OnCommands([]string{"词云"}).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if args := utils.GetArgs(ctx); args != "" && zero.SuperUserPermission(ctx) {
			g, err := strconv.ParseInt(args, 10, 64)
			if err != nil {
				ctx.Send("群号格式不对哦，可以看看帮助")
				return
			}
			gid = g
		}
		msgs, err := history.GroupHistoryToday(gid)
		if err != nil {
			ctx.Send("失败了...")
			logrus.Warn("<wordloud>获取群历史信息失败", err)
			return
		}
		if len(msgs) == 0 {
			ctx.Send("还没有记录消息...")
			return
		}
		plain := extractPlain(msgs)
		img, err := GetWordCloud(plain)
		if err != nil {
			ctx.Send("失败了...")
			logrus.Warn("<wordloud>生成词云失败", err)
			return
		}
		ctx.Send(message.ImageBytes(img))
	})
	proxy.OnCommands([]string{"全部词云"}).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if args := utils.GetArgs(ctx); args != "" && zero.SuperUserPermission(ctx) {
			g, err := strconv.ParseInt(args, 10, 64)
			if err != nil {
				ctx.Send("群号格式不对哦，可以看看帮助")
				return
			}
			gid = g
		}
		msgs, err := history.GroupHistoryAll(gid)
		if err != nil {
			ctx.Send("失败了...")
			logrus.Warn("<wordloud>获取群历史信息失败", err)
			return
		}
		if len(msgs) == 0 {
			ctx.Send("还没有记录消息...")
			return
		}
		plain := extractPlain(msgs)
		os.WriteFile(fmt.Sprintf("word_%d.txt", gid), []byte(plain), 0644)
		img, err := GetWordCloud(plain)
		if err != nil {
			ctx.Send("失败了...")
			logrus.Warn("<wordloud>生成词云失败", err)
			return
		}
		ctx.Send(message.ImageBytes(img))
	})
}

func extractPlain(msgs []history.SimpleEvent) (out string) {
	for _, msg := range msgs {
		for _, seg := range msg.Msg {
			if seg.Type == "text" {
				out += seg.Data["text"]
			}
		}
	}
	return
}
