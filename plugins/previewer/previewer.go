// Package previewer a plugin to generate preview images
package previewer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/FloatTech/floatbox/binary"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/playwright-community/playwright-go"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type generator struct {
	name  string
	check zero.Rule
	send  func(matched []string, ctx *zero.Ctx) error
}

var mappers map[*regexp.Regexp]generator

var ErrBlocked = errors.New("该链接已被黑名单正则屏蔽")

type ShotConfig struct {
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	DPI     float64 `json:"dpi"`
	Wait    int     `json:"wait"`
	WaitJs  string  `json:"wait_js"`
	JS      string  `json:"js"`
	Css     string  `json:"css"`
	Quality int     `json:"quality"`
}
type PreviewType string

var (
	TypeScreenShot PreviewType = "SCREEN_SHOT"
	TypeMusic      PreviewType = "CQ_MUSIC"
	TypeCQMessage  PreviewType = "CQ_MESSAGE"
)

type Config struct {
	Type                PreviewType  `json:"type"`
	Name                string       `json:"name"`
	Regex               string       `json:"regex"`
	BlacklistRegex      string       `json:"blacklist_regex"`
	MatchedGroup        int          `json:"matched_group"`
	UrlReplacementRegex string       `json:"url_replacement_regex"`
	UrlReplacement      string       `json:"url_replacement"`
	ErrorTemplate       string       `json:"error_template"`
	ScreenShotConfig    ShotConfig   `json:"config"`
	AllowGroup          []int64      `json:"allow_group"`
	Extra               gjson.Result `json:"-"`
}

func (c *Config) UnmarshalJSON(bytes []byte) error {
	// log.Infoln("<previewer> 原始 JSON 数据:", string(bytes))
	g := gjson.ParseBytes(bytes)
	c.Type = PreviewType(g.Get("type").Str)
	c.Name = g.Get("name").Str
	c.Regex = g.Get("regex").Str
	c.BlacklistRegex = g.Get("blacklist_regex").Str
	c.MatchedGroup = int(g.Get("matched_group").Int())
	c.UrlReplacementRegex = g.Get("url_replacement_regex").Str
	c.UrlReplacement = g.Get("url_replacement").Str
	c.ErrorTemplate = g.Get("error_template").Str
	// log.Infoln("<previewer> 解析后的部分字段:", c.Type, c.Name, c.Regex)
	if g.Get("config").Exists() {
		err := json.Unmarshal(binary.StringToBytes(g.Get("config").Raw), &c.ScreenShotConfig)
		if err != nil {
			log.Errorf("<previewer> 解析 ScreenShotConfig 错误: %v", err)
			return err
		}
	}
	for _, result := range g.Get("allow_group").Array() {
		c.AllowGroup = append(c.AllowGroup, result.Int())
	}
	c.Extra = g.Get("extra")
	// log.Infoln("<previewer> 解析后的 Extra 字段:", c.Extra)
	return nil
}

func checkAllow(group int64, allowed []int64) bool {
	return slices.Contains(allowed, group)
}

type triggered struct {
	gen     *generator
	matched []string
}

func initMapper(e *manager.PluginProxy) {
	mappers = make(map[*regexp.Regexp]generator)
	jb, err := e.ReadData("config.json")
	if err != nil {
		return
	}
	{
		configs := make([]Config, 0)
		err := json.Unmarshal(jb, &configs)
		if err != nil {
			return
		}
		for _, v0 := range configs {
			v := v0
			log.Infoln("<previewer> 加载预览模板:", v.Name, v.Type, v.AllowGroup)

			var check zero.Rule
			if len(v.AllowGroup) > 0 {
				var tmp = make([]int64, 0, len(v.AllowGroup))
				tmp = append(tmp, v.AllowGroup...)
				check = func(ctx *zero.Ctx) bool {
					log.Infoln("<previewer> 预览模板", v.Name, "check:", tmp)
					if zero.SuperUserPermission(ctx) {
						log.Warnln("<previewer> 允许超级用户调用", ctx.Event.GroupID, ctx.Event.UserID)
						return true
					}
					if checkAllow(ctx.Event.GroupID, tmp) {
						log.Warnln("<previewer> 允许群组调用", ctx.Event.GroupID, ctx.Event.UserID)
						return true
					}
					if checkAllow(ctx.Event.UserID, tmp) {
						log.Warnln("<previewer> 允许用户调用", ctx.Event.GroupID, ctx.Event.UserID)
						return true
					}
					log.Warnln("<previewer> 不允许调用", ctx.Event.GroupID, ctx.Event.UserID)
					return false
				}
			}
			switch v.Type {
			case TypeScreenShot:
				var replacer func(string) string
				var blacklist *regexp.Regexp
				if v.UrlReplacementRegex != "" {
					re, err := regexp.Compile(v.UrlReplacementRegex)
					if err != nil {
						log.Warnln("<previewer> 预览模板", v.Name, "url_replacement_regex 配置错误:", err)
						continue
					}
					replacer = func(s string) string {
						return re.ReplaceAllString(s, v.UrlReplacement)
					}
				}
				if v.BlacklistRegex != "" {
					blacklist, err = regexp.Compile(v.BlacklistRegex)
					if err != nil {
						log.Warnln("<previewer> 预览模板", v.Name, "blacklist_regex 配置错误:", err)
						continue
					}
				}
				mappers[regexp.MustCompile(v.Regex)] = generator{
					name:  v.Name,
					check: check,
					send: func(matched []string, ctx *zero.Ctx) error {
						var u = matched[v.MatchedGroup]
						if replacer != nil {
							u = replacer(u)
						}
						if blacklist != nil && blacklist.MatchString(u) {
							return ErrBlocked
						}
						opt := utils.DefaultPageOptions
						opt.Quality = playwright.Int(v.ScreenShotConfig.Quality)
						opt.Style = playwright.String(fmt.Sprintf("%s\n%s", *opt.Style, v.ScreenShotConfig.Css))
						bytes, err := utils.ScreenShotPageURL(u, utils.ScreenShotPageOption{
							Width:  v.ScreenShotConfig.Width,
							Height: v.ScreenShotConfig.Height,
							DPI:    v.ScreenShotConfig.DPI,
							Sleep:  time.Duration(v.ScreenShotConfig.Wait) * time.Second,
							Before: func(page playwright.Page) {
								if v.ScreenShotConfig.WaitJs != "" {
									_, err := page.WaitForFunction(v.ScreenShotConfig.WaitJs, nil, playwright.PageWaitForFunctionOptions{
										Polling: 500,
										Timeout: utils.DefaultPageOptions.Timeout,
									})
									if err != nil {
										log.Warningf("[previewer] WaitJS error %v", err)
									}
								}
								if v.ScreenShotConfig.JS != "" {
									_, err := page.Evaluate(v.ScreenShotConfig.JS, v)
									if err != nil {
										log.Warningf("[previewer] JS error %v", err)
									}
								}
							},
							PwOption: opt,
						})
						if err != nil {
							return err
						}
						ctx.Send(message.ImageBytes(bytes))
						return nil
					},
				}
			case TypeCQMessage:
				re := regexp.MustCompile(v.Regex)
				var replacer func(string) string
				var blacklist *regexp.Regexp
				if v.UrlReplacementRegex != "" {
					re, err := regexp.Compile(v.UrlReplacementRegex)
					if err != nil {
						log.Warnln("<previewer> 预览模板", v.Name, "url_replacement_regex 配置错误:", err)
						continue
					}
					replacer = func(s string) string {
						return re.ReplaceAllString(s, v.UrlReplacement)
					}
				}
				if v.BlacklistRegex != "" {
					blacklist, err = regexp.Compile(v.BlacklistRegex)
					if err != nil {
						log.Warnln("<previewer> 预览模板", v.Name, "blacklist_regex 配置错误:", err)
						continue
					}
				}
				mappers[re] = generator{
					name:  v.Name,
					check: check,
					send: func(matched []string, ctx *zero.Ctx) error {
						var msg = matched[v.MatchedGroup]
						if replacer != nil {
							msg = replacer(msg)
						}
						if blacklist != nil && blacklist.MatchString(msg) {
							return ErrBlocked
						}
						if v.Extra.IsObject() && v.Extra.Get("http").Bool() {
							c := client.NewHttpClient(nil)
							resp, err := c.Get(msg)
							if err != nil {
								return err
							}
							defer resp.Body.Close()
							data, err := io.ReadAll(resp.Body)
							if err != nil {
								return err
							}
							if gjson.ValidBytes(data) && gjson.ParseBytes(data).Get("type").Str == "json" {
								ctx.Send(message.JSON(string(data)))
							} else {
								ctx.Send(string(data))
							}
						} else {
							ctx.Send(msg)
						}
						return nil
					},
				}
			case TypeMusic:
				{
					re := regexp.MustCompile(v.Regex)
					var replacer func(string) string
					var blacklist *regexp.Regexp
					if v.UrlReplacementRegex != "" {
						re, err := regexp.Compile(v.UrlReplacementRegex)
						if err != nil {
							log.Warnln("<previewer> 预览模板", v.Name, "url_replacement_regex 配置错误:", err)
							continue
						}
						replacer = func(s string) string {
							return re.ReplaceAllString(s, v.UrlReplacement)
						}
					}
					if v.BlacklistRegex != "" {
						blacklist, err = regexp.Compile(v.BlacklistRegex)
						if err != nil {
							log.Warnln("<previewer> 预览模板", v.Name, "blacklist_regex 配置错误:", err)
							continue
						}
					}
					mappers[re] = generator{
						name:  v.Name,
						check: check,
						send: func(matched []string, ctx *zero.Ctx) error {
							var msg = matched[v.MatchedGroup]
							if replacer != nil {
								msg = replacer(msg)
							}
							if blacklist != nil && blacklist.MatchString(msg) {
								return ErrBlocked
							}
							d := ctx.CallAction("get_music_ark", zero.Params{
								"musicUrl": fmt.Sprintf("http://music.163.com/song/media/outer/url?id=%s", msg),
							})
							println(d.Data.Str)
							ctx.Send(message.JSON(d.Data.Str))
							return nil
						},
					}
				}
			}
		}
	}
}
func extractValues(json gjson.Result, result *strings.Builder, depth int) {
	if depth > 20 {
		return
	}
	switch json.Type {
	case gjson.String:
		result.WriteString(json.Str)
	case gjson.JSON:
		json.ForEach(func(key, value gjson.Result) bool {
			extractValues(value, result, depth+1)
			return true // 继续遍历
		})
	default: // 对于其他类型（如数组），直接处理其元素
		json.ForEach(func(key, value gjson.Result) bool {
			extractValues(value, result, depth+1)
			return true // 继续遍历
		})
	}
}
func init() {
	e := manager.RegisterPlugin(manager.PluginInfo{
		Name: "图片预览",
		Usage: `自动识别某些消息并生成预览图片
注：某些预览需要向Bot妈申请权限, 你可以使用 /report 或者 #联系管理员 功能来向Bot妈申请
`,
		IsPassive: true,
		Classify:  "实用工具",
	})
	initMapper(e)
	e.OnCommands([]string{"previewer"}, zero.SuperUserPermission, func(ctx *zero.Ctx) bool {
		return strings.TrimSpace(ctx.State["args"].(string)) == "reload"
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		initMapper(e)
		ctx.Send(fmt.Sprintf("重载预览模板成功,共%d", len(mappers)))
	})
	e.OnMessage(func(ctx *zero.Ctx) bool {
		var triggers = make([]triggered, 0, 4)
		rawMessage := ctx.Event.RawMessage
		if len(ctx.Event.Message) > 0 && ctx.Event.Message[0].Type == "json" {
			sb := &strings.Builder{}
			extractValues(gjson.Parse(ctx.Event.Message[0].Data["data"]), sb, 0)
			rawMessage = "[json_msg]" + sb.String()
		}
		for r, v := range mappers {
			v := v
			if r.MatchString(rawMessage) {
				if strings.HasPrefix(v.name, "public-") {
					goto pass
				}
				if zero.SuperUserPermission(ctx) {
					goto pass
				}
				if v.check != nil && !v.check(ctx) {
					continue
				}
			pass:
				log.Infoln("<previewer>", v.name, "触发")
				triggers = append(triggers, triggered{
					gen:     &v,
					matched: r.FindStringSubmatch(rawMessage),
				})
			}
		}
		ctx.State["triggers"] = triggers
		return len(triggers) != 0
	}).Handle(func(ctx *zero.Ctx) {
		for _, t := range ctx.State["triggers"].([]triggered) {
			go func() {
				previewerName := t.gen.name
				matched := t.matched
				sender := t.gen.send
				err := sender(matched, ctx)
				if err != nil {
					log.Errorf("<previewer> error [%s]: %v", previewerName, err)
					return
				}
			}()
			time.Sleep(1 * time.Second)
		}
	})
}
