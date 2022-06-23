package poke

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/client"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strconv"
	"time"
)

var info = manager.PluginInfo{
	Name: "戳一戳",
	Usage: `用法：
	轻轻戳一下~`,
}

var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	//proxy.OnFullMatch([]string{"网易云评论"}).SetBlock(true).SecondPriority().Handle(getComment)
	proxy.OnMessage(checkEvent).SetBlock(true).SecondPriority().Handle(pokeHandler)
	proxy.AddConfig("cooldown", int64(3000))
	proxy.AddConfig("pokeMessage", []string{"ヾ(≧へ≦)〃", "ヽ（≧□≦）ノ", "再戳我你就是大变态<( ￣^￣)", "变态,不许戳！！！", "变态、变态、变态、变态、变态笨蛋大变态!!!"})
	proxy.AddConfig("pokeBack", []string{"0", "0", "0", "1", "3"})
}

func checkEvent(ctx *zero.Ctx) bool {
	log.Infof(ctx.Event.PostType)
	if /*ctx.Event.PostType == "poke" &&*/ ctx.Event.TargetID == ctx.Event.SelfID {
		return checkCall(ctx, Call{
			groupId:  ctx.Event.GroupID,
			senderId: ctx.Event.UserID,
		})
	}
	return false
}

// checkCall 判断调用间隔是否合理
func checkCall(ctx *zero.Ctx, call Call) bool {
	var now = time.Now().Unix()
	var last = pokeHistory[call.groupId]
	if now-last < proxy.GetConfigInt64("cooldown") {
		return false
	}
	if last == 0 {
		pokeHistory[call.groupId] = time.Now().Unix()
		return true
	}
	ctx.State["call"] = call
	return true
}

// newCall 获取用户戳一戳调用次数并更新
func newCall(call Call) int32 {
	count := pokeCount[call]
	pokeCount[call] = count + 1
	return count
}

// clearCall 重置用户戳一戳调用次数
func clearCall(call Call) {
	pokeCount[call] = 0
}

//
func action(ctx *zero.Ctx, call Call, count int32) {
	ctx.Send(proxy.GetConfigStrings("pokeMessage")[count])
	pokesStr := proxy.GetConfigStrings("pokeBack")[count]
	pokes, err := strconv.Atoi(pokesStr)
	if err != nil {
		pokes = 0
		log.Errorf("戳一戳配置错误, %s无法被识别为整数", pokesStr)
	}
	for i := 0; i < pokes; i++ {
		ctx.Send(message.Poke(ctx.Event.UserID))
	}

	if count == 4 {
		clearCall(call)
	}
}

func pokeHandler(ctx *zero.Ctx) {
	var call, err = ctx.State["call"].(Call)
	if !err {
		log.Errorf("无法获取调用对象")
		return
	}
	action(ctx, call, newCall(call))
}

// Call 一个调用事件
type Call struct {
	groupId  int64
	senderId int64
}

// 用户戳一戳调用次数
var pokeCount map[Call]int32

// 只在群内做调用限制
var pokeHistory map[int64]int64

const repingURL = "https://api.vvhan.com/api/reping"

func getComment(ctx *zero.Ctx) {
	var c = client.NewHttpClient(nil)
	json, err := c.GetGJson(repingURL)
	if err != nil || !json.Get("success").Bool() {
		log.Warnf("reping err: user=%v,url=%v,err=%v", ctx.Event.UserID, repingURL, err)
	}
	ctx.Send(message.Text(json.Get("data").Get("content")))
}
