// Package browser playwright浏览器相关
package browser

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/ctxext"
	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/shell"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := manager.RegisterPlugin(manager.PluginInfo{
		Name: "网页截图",
		Usage: `注意：
只允许白名单网址截图
用法：
	{cmd}截图 <网址>`,
	})
	type cmd struct {
		Width   int      `arg:"-W" default:"1280"`
		Height  int      `arg:"-H" help:"截图高度，0表示全屏" default:"0"`
		DPI     float64  `arg:"--dpi" default:"1.5"`
		Quality int      `arg:"-q" help:"截图质量，最高100" default:"70"`
		URL     []string `arg:"positional"`
	}

	engine.OnCommands([]string{"截图", "截屏"}, func(ctx *zero.Ctx) bool {
		var screenShotCmd = cmd{}
		browserArgsParser, err := arg.NewParser(arg.Config{Program: zero.BotConfig.CommandPrefix + "截图", IgnoreEnv: true}, &screenShotCmd)
		if err != nil {
			panic(err)
		}
		err = browserArgsParser.Parse(shell.Parse(ctx.State["args"].(string)))
		if err != nil || len(screenShotCmd.URL) == 0 {
			buf := strings.Builder{}
			browserArgsParser.WriteHelp(&buf)
			ctx.Send(buf.String())
			ctx.Break()
			return false
		}
		ctx.State["flag"] = screenShotCmd
		if zero.SuperUserPermission(ctx) {
			return true
		}
		json, err := engine.ReadJson("config.json")
		if err != nil {
			log.Errorln("<browser>", err)
			return false
		}
		for _, u := range screenShotCmd.URL {
			if !strings.HasPrefix(u, "http") {
				u = fmt.Sprintf("https://%s", u)
			}
			parse, err := url.Parse(u)
			if err != nil {
				log.Errorln("<browser>", err)
				return false
			}
			for _, host := range json.Get(fmt.Sprintf("blacklist.all")).Array() {
				if parse.Host == host.String() {
					ctx.Send("不允许截图该网址")
					goto notok
				}
			}
			for _, host := range json.Get(fmt.Sprintf("blacklist.group.%d", ctx.Event.GroupID)).Array() {
				if parse.Host == host.String() {
					ctx.Send("不允许截图该网址")
					goto notok
				}
			}
			for _, host := range json.Get(fmt.Sprintf("blacklist.user.%d", ctx.Event.UserID)).Array() {
				if parse.Host == host.String() {
					ctx.Send("不允许截图该网址")
					goto notok
				}
			}
			for _, host := range json.Get(fmt.Sprintf("whitelist.group.%d", ctx.Event.GroupID)).Array() {
				if parse.Host == host.String() {
					goto ok
				}
			}
			for _, host := range json.Get(fmt.Sprintf("whitelist.all")).Array() {
				if parse.Host == host.String() {
					goto ok
				}
			}
			// 不在全局白名单
			for _, fullmatch := range json.Get(fmt.Sprintf("exact")).Array() {
				if strings.Replace(parse.String(), "https://", "", 1) == fullmatch.String() {
					goto ok
				}
			}
		notok:
			// No
			ctx.CallAction("set_group_reaction", zero.Params{
				"group_id":   ctx.Event.GroupID,
				"message_id": ctx.Event.MessageID,
				"code":       "123",
				"is_add":     true,
			})
			log.Warnln("<previewer>", "截图请求被拒绝")
			return false
		ok:
			continue

		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		model := ctx.State["flag"].(cmd)
		c := http.Client{}
		// 按钮
		ctxext.ReactionLoadingAdd(ctx)
		defer ctxext.ReactionLoadingRemove(ctx)
		for _, u := range model.URL {
			option := url.Values{}
			option.Set("url", u)
			option.Set("width", strconv.Itoa(model.Width))
			option.Set("height", strconv.Itoa(model.Height))
			option.Set("factor", strconv.FormatFloat(model.DPI, 'f', 2, 64))
			option.Set("quality", strconv.Itoa(model.Quality))
			parse, _ := url.Parse("http://localhost:5544/preview")
			parse.RawQuery = option.Encode()
			response, err := c.Get(parse.String())
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			defer response.Body.Close()
			img, err := io.ReadAll(response.Body)
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			ctx.Send(message.ImageBytes(img))
		}
	})
}
