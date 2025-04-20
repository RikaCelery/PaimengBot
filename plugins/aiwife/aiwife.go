// Package aiwife 随机老婆
package aiwife

import (
	"fmt"
	"math/rand"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/client"
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
		Brief:    "随机ai生成图",
		Classify: "其他",
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
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		defer proxy.UnlockUser(ctx.Event.UserID)
		miku := rand.Intn(100000) + 1
		bytes, err := client.GetBytesRetry(fmt.Sprintf(bed, miku), 3)
		if err != nil {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.At(ctx.Event.UserID), message.ImageBytes(bytes))
	})
}
