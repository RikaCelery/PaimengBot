package dice

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "骰子",
	Usage: `
快速随机数生成
用法：
	{cmd}[个数]d[最大值]：扔[个数]次骰子，骰子数字范围1-[最大值]
备注：
	可以同时输入多组参数，例如：{cmd}3d3 4d4 9d3
`,
	Classify: "实用工具",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnMessage(zero.NewPattern(nil).Command("(?i)(?:\\d+d\\d+\\s*)+").AsRule(), zero.OnlyToMe).SetBlock(true).ThirdPriority().Handle(diceHandler)
	proxy.AddConfig("max", 100)
}

func diceHandler(ctx *zero.Ctx) {
	type Dice struct {
		Max   int
		Count int
	}
	parsed := extension.PatternModel{}
	_ = ctx.Parse(&parsed)
	diceReg := regexp.MustCompile(`(?i)(\d+)d(\d+)`)
	var dices []Dice
	for _, submatch := range diceReg.FindAllStringSubmatch(parsed.Matched[0].Text()[0], -1) {
		if len(submatch) != 3 {
			ctx.SendChain(message.Text("骰子参数错误"))
			return
		}
		if len(submatch[1]) == 0 || len(submatch[2]) == 0 {
			ctx.SendChain(message.Text("骰子参数错误"))
			return
		}
		max, err := strconv.ParseInt(submatch[2], 10, 32)
		if err != nil {
			ctx.SendChain(message.Text("骰子参数错误"))
			return
		}
		count, err := strconv.ParseInt(submatch[1], 10, 32)
		if err != nil {
			ctx.SendChain(message.Text("骰子参数错误"))
			return
		}
		if count < 1 || max < 1 {
			ctx.SendChain(message.Text("骰子参数错误"))
			return
		}
		if max > proxy.GetConfigInt64("max") {
			ctx.SendChain(message.Text("你投太多次了"))
			return
		}
		dices = append(dices, Dice{
			Max:   int(max),
			Count: int(count),
		})
	}
	sb := &strings.Builder{}
	for i, dice := range dices {
		if i > 0 {
			sb.WriteString("\n-------\n")
		}
		for i := 0; i < dice.Count; i++ {
			sb.WriteString(strconv.Itoa(rand.Intn(dice.Max) + 1))
			if i != dice.Count-1 {
				sb.WriteString(", ")
			}
		}
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(sb.String()))
}
