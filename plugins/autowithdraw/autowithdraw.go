package autowithdraw

import (
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
	}).SetBlock(false).Handle(withDrawMsg)
}

func withDrawMsg(ctx *zero.Ctx) {
	id, ok := ctx.Event.MessageID.(int64)
	if !ok {
		return
	}
	for _, msg := range zero.GetTriggeredMessages(message.NewMessageIDFromInteger(id)) {
		if ctx.Event.GroupID != -ctx.Event.UserID {
			ctx.DeleteMessage(msg)
		}
	}
}
