// Package aiwife 随机老婆
package aiwife

import (
	"fmt"
	"math/rand"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	bed = "https://www.thiswaifudoesnotexist.net/example-%d.jpg"
)

var (
	info = manager.PluginInfo{
		Name: "ai老婆",
		Usage: `随机一个ai生成的二次元老婆
用法：
	{cmd}aiwife
	{cmd}ai老婆
备注：
	等上一条发出去之后才可以再发`,
		Brief: "ai随机生成老婆",
	}
	proxy *manager.PluginProxy
)

func init() { // 插件主体
	proxy = manager.RegisterPlugin(info)
	proxy.OnCommands([]string{"aiwife", "ai老婆"}, func(ctx *zero.Ctx) bool {
		locked := proxy.LockUser(ctx.Event.UserID)
		if locked {
			logrus.Infoln("<aiwife> user locked")
		}
		return !locked
	}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			miku := rand.Intn(100000) + 1
			ctx.SendChain(message.At(ctx.Event.UserID), message.Image(fmt.Sprintf(bed, miku)))
			proxy.UnlockUser(ctx.Event.UserID)
		})
}
