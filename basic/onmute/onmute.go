package onmute

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name:        "禁言提醒",
	Usage:       `在禁言时通知`,
	IsPassive:   true,
	IsSuperOnly: true,
	Classify:    "内置功能",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.On("notice/group_ban/ban").FirstPriority().Handle(handleBotMute)
}

func handleBotMute(ctx *zero.Ctx) {
	if ctx.Event.UserID == ctx.Event.SelfID {
		ctx.SendPrivateMessage(ctx.Event.OperatorID, fmt.Sprintf("您好不需要的功能可以关闭（ %s关闭 功能名）或（mikaku/禁用 某个功能），烦请不要禁言", zero.BotConfig.CommandPrefix))
	}
}
