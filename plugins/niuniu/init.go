// Package niuniu 牛牛大作战
package niuniu

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/basic/sc"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RomiChan/syncx"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	configCostPerUnregister = "cost_per_unregister"
	configDajiaoLimit       = "dajiao_per_hour"
	configJJLimit           = "jj_per_hour"
	configProfit            = "profit"
)

var (
	info = manager.PluginInfo{
		Name: "牛牛大作战",
		Usage: `
用法：
	{cmd}注册牛牛

	（无需{cmd}）打胶
	（无需{cmd}）使用[道具名称]打胶
	（无需{cmd}）jj@xxx
	（无需{cmd}）使用[道具名称]jj@xxx
	{cmd}赎牛牛(cd:60分钟)
	{cmd}出售牛牛
	{cmd}牛牛拍卖行
	{cmd}牛牛商店
	{cmd}牛牛背包
	{cmd}注销牛牛
	{cmd}查看我的牛牛
	{cmd}牛子长度排行
	{cmd}牛子深度排行
ps : 出售后的牛牛都会进入牛牛拍卖行哦`,
		SuperUsage: `
config-plugin配置项：
	niuniu.profit：卖牛牛每厘米获得的真实金钱
	niuniu.redeem_cost：赎牛牛所需真实金钱
	niuniu.unregister_rate：注销牛牛的真实金钱倍率（次数*倍率）
	niuniu.dajiao_per_hour：注销牛牛的真实金钱倍率（次数*倍率）
	niuniu.jj_per_hour：注销牛牛的真实金钱倍率（次数*倍率）
`,
		Classify:    "小游戏",
		IsPassive:   false,
		IsSuperOnly: false,
		AdminLevel:  0,
	}
	proxy         = manager.RegisterPlugin(info)
	dajiaoLimiter = rate.NewManager[string](time.Second*300, 1)
	jjLimiter     = rate.NewManager[string](time.Second*600, 1)
	jjCount       = syncx.Map[string, *lastLength]{}
	register      = syncx.Map[string, *lastLength]{}
)

func init() {
	proxy.AddConfig("shop_item", []string{
		`{"name":"伟哥", "cost":2, "scope":"打胶", "description":"可以让你打胶每次都增长", "count":5}`,
		`{"name":"媚药", "cost":2, "scope":"打胶", "description":"可以让你打胶每次都减少","count": 5}`,
		`{"name":"击剑神器", "cost":10, "scope":"jj", "description":"可以让你每次击剑都立于不败之地", "count":2}`,
		`{"name":"击剑神稽", "cost":10, "scope":"jj", "description":"可以让你每次击剑都失败", "count":2}`,
	})
	proxy.OnCommands([]string{"牛牛拍卖行"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		auction, err := ShowAuction(gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		var sb = &strings.Builder{}
		sb.WriteString("牛牛拍卖行有以下牛:\n请输入对应序号进行购买\n\n")
		for _, info := range auction {
			msg := fmt.Sprintf("商品序号: [%d]\n牛牛原所属: %d\n牛牛价格: %.0f%s\n牛牛大小: %.2fcm",
				info.ID+1, info.UserID, info.Money*sc.Rate(), sc.Unit(), info.Length)
			sb.WriteString(msg + "\n\n")
		}
		segment, _ := imageCard(sb.String())
		ctx.SendChain(message.Reply(ctx.Event.MessageID), segment)
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.CheckUser(uid), zero.CheckGroup(gid), zero.RegexRule(`^(\d+)$`)).Repeat()
		defer cancel()
		timer := time.NewTimer(120 * time.Second)
		answer := ""
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				ctx.SendChain(message.At(uid), message.Text(" 超时,已自动取消"))
				return
			case r := <-recv:
				answer = r.Event.Message.String()
				n, err := strconv.Atoi(answer)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				n--
				msg, err := Auction(gid, uid, n)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				ctx.SendChain(message.Reply(ctx.Event.Message), message.Text(msg))
				return
			}
		}
	})
	proxy.OnCommands([]string{"出售牛牛"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		sell, err := Sell(gid, uid, proxy.GetConfigFloat64(configProfit))
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(sell))
	})
	proxy.OnCommands([]string{"牛牛背包"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		bag, err := Bag(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(bag))
	})
	proxy.OnCommands([]string{"牛牛商店"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		if proxy.LockUser(uid) {
			return
		}
		defer proxy.UnlockUser(uid)

		if _, err := GetWordNiuNiu(gid, uid); err != nil {
			ctx.SendChain(message.Text(ErrNoNiuNiu))
			return
		}

		propMap := map[int]shopItem{}
		for i, item := range getShopItems() {
			propMap[i+1] = item
		}
		sb := &strings.Builder{}
		sb.WriteString("牛牛商店当前售卖的物品如下\n")
		for id := range propMap {
			product := propMap[id]
			productInfo := fmt.Sprintf("商品[%d]\n商品名: %s\n商品价格: %.2f%s\n商品作用域: %s\n商品描述: %s\n使用次数:%d",
				id, product.Name, float64(product.Cost)*sc.Rate(), sc.Unit(), product.Scope, product.Description, product.Cost)
			sb.WriteString(productInfo + "\n\n")
		}
		sb.WriteString("\n输入对应序号进行购买商品")
		msg, err := imageCard(sb.String())
		if err != nil {
			ctx.Send(message.Text("发送商店失败", err.Error()))
			utils.SetNotStatistic(ctx)
			return
		}
		ctx.Send(msg)
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.CheckUser(uid), zero.CheckGroup(gid), zero.RegexRule(`^(\d+)$`)).Repeat()
		defer cancel()
		timer := time.NewTimer(120 * time.Second)
		answer := ""
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				ctx.SendChain(message.At(uid), message.Text(" 超时,已自动取消"))
				return
			case r := <-recv:
				answer = r.Event.Message.String()
				n, err := strconv.Atoi(answer)
				if err != nil {
					return
				}
				item, ok := propMap[n]
				if !ok {
					ctx.SendChain(message.Text("商品不存在!"))
				}
				if err = Store(gid, uid, float64(item.Cost), n); err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text("购买成功!"))
				return
			}
		}
	})
	proxy.OnCommands([]string{"赎牛牛"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		last, ok := jjCount.Load(fmt.Sprintf("%d_%d", gid, uid))

		if !ok {
			ctx.SendChain(message.Text("你还没有被厥呢"))
			return
		}

		if time.Since(last.TimeLimit) > time.Hour {
			ctx.SendChain(message.Text("时间已经过期了,牛牛已被收回!"))
			jjCount.Delete(fmt.Sprintf("%d_%d", gid, uid))
			return
		}

		if last.Count < 4 {
			ctx.SendChain(message.Text("你还没有被厥够4次呢,不能赎牛牛"))
			return
		}
		ctx.SendChain(message.Text("再次确认一下哦,这次赎牛牛，牛牛长度将会变成", last.Length, "cm\n还需要嘛【是|否】"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.CheckUser(uid), zero.CheckGroup(gid), zero.RegexRule(`^(是|否)$`)).Repeat()
		defer cancel()
		timer := time.NewTimer(2 * time.Minute)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				ctx.SendChain(message.Text("操作超时，已自动取消"))
				return
			case c := <-recv:
				answer := c.Event.Message.String()
				if answer == "否" {
					ctx.SendChain(message.Text("取消成功!"))
					return
				}

				if err := Redeem(gid, uid, last.Length); err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}

				jjCount.Delete(fmt.Sprintf("%d_%d", gid, uid))

				ctx.SendChain(message.At(uid), message.Text(fmt.Sprintf("恭喜你!成功赎回牛牛,当前长度为:%.2fcm", last.Length)))
				return
			}
		}
	})
	proxy.OnCommands([]string{"牛子长度排行"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		infos, err := GetRankingInfo(gid, true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		img, err := processRankingImg(infos, ctx, true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.ImageBytes(img))
	})
	proxy.OnCommands([]string{"牛子深度排行"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		infos, err := GetRankingInfo(gid, false)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		img, err := processRankingImg(infos, ctx, false)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.ImageBytes(img))
	})
	proxy.OnCommands([]string{"查看我的牛牛"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		view, err := View(gid, uid, ctx.CardOrNickName(uid))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(view))
	})
	proxy.OnRegex(`^(?:.*使?用(.*))??打胶$`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if !proxy.CheckCallLimit("dajiao", ctx.Event.UserID) {
			ctx.Send(fmt.Sprintf("每小时只能打%d次～，歇一会儿吧～", proxy.GetConfigInt64(configDajiaoLimit)))
			utils.SetNotStatistic(ctx)
			return
		}
		// 获取群号和用户ID
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		fiancee := ctx.State["regex_matched"].([]string)

		msg, err := HitGlue(gid, uid, fiancee[1])
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			dajiaoLimiter.Delete(fmt.Sprintf("%d_%d", ctx.Event.GroupID, ctx.Event.UserID))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(msg))
	})
	proxy.OnCommands([]string{"注册牛牛"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		msg, err := Register(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(msg))
	})
	proxy.OnMessage(zero.NewPattern(nil).Text(`^(?:.*使?用(.*))??jj`).At().AsRule(),
		zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if !proxy.CheckCallLimit("jj", ctx.Event.UserID) {
			ctx.Send(fmt.Sprintf("每小时只能击剑%d次～，歇一会儿吧～", proxy.GetConfigInt64(configJJLimit)))
			utils.SetNotStatistic(ctx)
			return
		}
		patternParsed := ctx.State[zero.KeyPattern].([]zero.PatternParsed)
		adduser, err := strconv.ParseInt(patternParsed[1].At(), 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			jjLimiter.Delete(fmt.Sprintf("%d_%d", ctx.Event.GroupID, ctx.Event.UserID))
			return
		}
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		msg, length, err := JJ(gid, uid, adduser, patternParsed[0].Text()[1])
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			jjLimiter.Delete(fmt.Sprintf("%d_%d", ctx.Event.GroupID, ctx.Event.UserID))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(msg))
		j := fmt.Sprintf("%d_%d", gid, adduser)
		count, ok := jjCount.Load(j)
		var c lastLength
		// 按照最后一次被jj时的时间计算，超过60分钟则重置
		if !ok {
			c = lastLength{
				TimeLimit: time.Now(),
				Count:     1,
				Length:    length,
			}
		} else {
			c = lastLength{
				TimeLimit: time.Now(),
				Count:     count.Count + 1,
				Length:    count.Length,
			}
			if time.Since(c.TimeLimit) > time.Hour {
				c = lastLength{
					TimeLimit: time.Now(),
					Count:     1,
					Length:    length,
				}
			}
		}

		jjCount.Store(j, &c)
		if c.Count > 2 {
			if c.Count <= 4 { // 不要一直发
				ctx.SendChain(message.Text(randomChoice([]string{
					fmt.Sprintf("你们太厉害了，对方已经被你们打了%d次了，你们可以继续找他🤺", c.Count),
					"你们不要再找ta🤺啦！"},
				)))
			}

			if c.Count >= 4 && c.Count <= 6 { // 不要一直发
				id := ctx.SendPrivateMessage(adduser,
					message.Text(fmt.Sprintf("你在%d群里已经被厥冒烟了，快去群里赎回你原本的牛牛!\n发送:`%s赎牛牛`即可！", gid, zero.BotConfig.CommandPrefix)))
				if id == 0 {
					ctx.SendChain(message.At(adduser), message.Text(fmt.Sprintf("快发送`%v赎牛牛`来赎回你原本的牛牛!", zero.BotConfig.CommandPrefix)))
				}
			}
		}
	})
	proxy.OnCommands([]string{"注销牛牛"}, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		key := fmt.Sprintf("%d_%d", gid, uid)
		data, ok := register.Load(key)
		switch {
		case !ok || time.Since(data.TimeLimit) > time.Hour*12:
			data = &lastLength{
				TimeLimit: time.Now(),
				Count:     1,
			}
		default:
			cost := float64(data.Count) * proxy.GetConfigFloat64(configCostPerUnregister)
			if _, ok := sc.AddBaseCoin(uid, cost); !ok {
				ctx.SendChain(message.Text("你的钱不够你注销牛牛了，这次注销需要", strconv.FormatFloat(cost*sc.Rate(), 'f', 2, 64), sc.Unit()))
				return
			}
		}
		register.Store(key, data)
		msg, err := Cancel(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(msg))
	})
	proxy.AddNormalCommands("打胶", "jj")
	proxy.AddConfig(configProfit, 5)
	proxy.AddConfig(configCostPerUnregister, 10)
	proxy.AddConfig(configDajiaoLimit, 2)
	proxy.SetCallLimiter("dajiao", time.Hour*1, 2).BindTimesConfig(configDajiaoLimit)
	proxy.AddConfig(configJJLimit, 2)
	proxy.SetCallLimiter("jj", time.Hour*1, 2).BindTimesConfig(configJJLimit)
}
