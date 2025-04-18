package autowithdraw

import (
	"math/rand"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name:      "自动撤回",
	Classify:  "内置功能",
	Usage:     `使用者撤回时，自动撤回机器人发的消息`,
	IsPassive: true,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnNotice(func(ctx *zero.Ctx) bool {
		return ctx.Event.NoticeType == "group_recall" || ctx.Event.NoticeType == "friend_recall"
	}, func(ctx *zero.Ctx) bool {
		id, ok := ctx.Event.MessageID.(int64)
		if !ok {
			return false
		}
		triggered := zero.GetTriggeredMessages(message.NewMessageIDFromInteger(id))
		if len(triggered) == 0 {
			return false
		}
		ctx.State["triggered"] = triggered
		return true
	}).SetBlock(false).Handle(withDrawMsg)
}

func withDrawMsg(ctx *zero.Ctx) {
	for _, msg := range ctx.State["triggered"].([]message.ID) {
		time.Sleep(time.Duration(rand.Intn(2000)+500) * time.Millisecond)
		ctx.DeleteMessage(msg)
	}
}
