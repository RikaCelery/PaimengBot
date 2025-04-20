// Package wife 抽老婆
package wife

import (
	"encoding/json"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := manager.RegisterPlugin(manager.PluginInfo{
		Name:  "抽老婆",
		Brief: "从老婆库抽每日老婆",
		Usage: `用法：
	{cmd}抽老婆：查看今日二次元老婆`,
		Classify: "其他",
	})
	engine.DataFolder()
	var cards []string
	engine.OnCommands([]string{"抽老婆"}, ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			data, err := engine.ReadData("wife.json")
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			err = json.Unmarshal(data, &cards)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			logrus.Infof("[wife]加载%d个老婆", len(cards))
			return true
		},
	)).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			card := cards[utils.RandSenderPerDayN(ctx.Event.UserID, len(cards))]
			data, err := engine.ReadData("wives/" + card)
			card, _, _ = strings.Cut(card, ".")
			if err != nil {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("今天的二次元老婆是~【", card, "】哒\n【图片下载失败: ", err, "】"),
				)
				return
			}
			if id := ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text("今天的二次元老婆是~【", card, "】哒"),
				message.ImageBytes(data),
			); id.ID() == 0 {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("今天的二次元老婆是~【", card, "】哒\n【图片发送失败, 请联系维护者】"),
				)
			}
		})
}
