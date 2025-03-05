package onmute

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "禁言提醒",
	Usage: `再禁言时通知
config-plugin配置项：
	event.greeting_msg: 加好友后的欢迎消息 留空不发送
	event.join_msg: 加群的介绍消息 留空不发送
	event.notautoleave: 是(true)否(false)关闭被动拉群时自动退群
	event.autoagree: 是(true)否(false)自动同意所有好友请求`,
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
	ctx.SendPrivateMessage(ctx.Event.OperatorID, fmt.Sprintf("您好不需要的功能可以关闭（ %s关闭 功能名）或（mikaku/禁用 某个功能），烦请不要禁言", zero.BotConfig.CommandPrefix))
}
