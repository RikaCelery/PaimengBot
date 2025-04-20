// Package saucenao P站ID/saucenao/ascii2d搜图
package saucenao

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/ctxext"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/jozsefsallai/gophersauce"
)

const (
	configEnablePicGlobal = "enablePic"
	configApiKey          = "apikey"
)

var (
	saucenaocli *gophersauce.Client
)

// MustProvidePicture 消息不存在图片阻塞120秒至有图片，超时返回 false
func MustProvidePicture(ctx *zero.Ctx) bool {
	if ctx.Event.Message[0].Type == "reply" {
		msg := ctx.GetMessage(ctx.Event.Message[0].Data["id"])
		var urls = []string{}
		for _, elem := range msg.Elements {
			if elem.Type == "image" {
				if elem.Data["url"] != "" {
					urls = append(urls, elem.Data["url"])
				} else if elem.Data["file"] != "" {
					urls = append(urls, elem.Data["file"])
				}
			}
		}
		if len(urls) > 0 {
			ctx.State["image_url"] = urls
			return true
		}
	}
	if zero.HasPicture(ctx) {
		return true
	}
	// 没有图片就索取
	ctx.SendChain(message.Text("请发送一张图片"))
	next := zero.NewFutureEvent("message", 0, true, ctx.CheckSession(), zero.HasPicture).Next()
	select {
	case <-time.After(time.Second * 10):
		return false
	case newCtx := <-next:
		ctx.State["image_url"] = newCtx.State["image_url"]
		ctx.Event.MessageID = newCtx.Event.MessageID
		return true
	}
}
func init() { // 插件主体
	engine := manager.RegisterPlugin(manager.PluginInfo{
		Name: "以图搜图",
		Usage: `用法：
	{cmd}以图搜图/{cmd}搜索图片/{cmd}以图识图/{cmd}source/{cmd}src?/{cmd}src？ [图片]
		回复一张图也可以自动识别`,
		SuperUsage: `	{cmd}^(开启|打开|启用|关闭|关掉|禁用)搜图显示图片$：为当前群 开启/关闭 图片显示`,
		Classify:   "实用工具",
	})
	engine.AddConfig(configEnablePicGlobal, false)
	engine.AddConfig(configApiKey, "")
	engine.AddOnConfigChange(func() {
		newcli, err := gophersauce.NewClient(&gophersauce.Settings{
			MaxResults: 1,
			APIKey:     engine.GetConfigString(configApiKey),
		})
		if err != nil {
			log.Errorln("saucenao api key error: ", err)
			return
		}
		saucenaocli = newcli
	})

	// 以图搜图
	engine.OnMessage(func(ctx *zero.Ctx) bool {
		raw := strings.TrimSpace(ctx.ExtractPlainText())
		if strings.HasPrefix(raw, zero.BotConfig.CommandPrefix) {
			raw = strings.TrimPrefix(raw, zero.BotConfig.CommandPrefix)
		} else {
			return false
		}
		commands := []string{"以图搜图", "搜索图片", "以图识图", "source", "src？", "src?"}
		for _, command := range commands {
			if strings.EqualFold(raw, command) {
				return true
			}
		}
		return false
	}, ctxext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		newcli, err := gophersauce.NewClient(&gophersauce.Settings{
			MaxResults: 1,
			APIKey:     engine.GetConfigString(configApiKey),
		})
		if err != nil {
			log.Errorln("<saucenao> api key error: ", err)
		} else {
			log.Infoln("<saucenao> init success")
			saucenaocli = newcli
			return true
		}
		return false
	})).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if !MustProvidePicture(ctx) {
				ctx.Send("那算了")
				return
			}
			// 开始搜索图片
			pics, ok := ctx.State["image_url"].([]string)
			if !ok {
				ctx.SendChain(message.Text("ERROR: 未获取到图片链接"))
				return
			}
			println(pics)
			if saucenaocli == nil {
				ctx.SendChain(message.Text("请设置saucenao api key, 详情请查看帮助图片"))
			}
			showPic := false
			if b, err := engine.GetLevelDB().Get([]byte(fmt.Sprintf("saucenao_enable_pic_%d", ctx.Event.GroupID)), nil); err == nil {
				showPic = b[0] == 1
			} else {
				showPic = engine.GetConfigBool(configEnablePicGlobal)
			}
			ctxext.ReactionLoadingAdd(ctx)
			defer ctxext.ReactionLoadingRemove(ctx)
			for _, pic := range pics {
				resp, err := saucenaocli.FromURL(pic)
				if err == nil && resp.Count() > 0 {
					result := resp.First()
					s, err := strconv.ParseFloat(result.Header.Similarity, 64)
					if err == nil {
						rr := reflect.ValueOf(&result.Data).Elem()
						b := binary.NewWriterF(func(w *binary.Writer) {
							r := rr.Type()
							for i := 0; i < r.NumField(); i++ {
								if !rr.Field(i).IsZero() {
									w.WriteString("\n")
									w.WriteString(r.Field(i).Name)
									w.WriteString(": ")
									w.WriteString(fmt.Sprint(rr.Field(i).Interface()))
								}
							}
						})
						resp, err := http.Head(result.Header.Thumbnail)
						msg := make(message.Message, 0, 3)
						if s > 80.0 {
							msg = append(msg, message.Text("我有把握是这个!"))
						} else {
							msg = append(msg, message.Text("也许是这个?"))
						}
						if showPic {
							if err == nil {
								_ = resp.Body.Close()
								if resp.StatusCode == http.StatusOK {
									msg = append(msg, message.Image(result.Header.Thumbnail))
								} else {
									msg = append(msg, message.Image(pic))
								}
							} else {
								msg = append(msg, message.Image(pic))
							}
						}
						msg = append(msg, message.Text("\n图源: ", result.Header.IndexName, binary.BytesToString(b)))
						ctx.Send(message.Message{ctxext.FakeSenderForwardNode(ctx, msg...)})
						if s > 80.0 {
							continue
						}
					}
				} else {
					log.Errorln("<saucenao>", err)
					ctx.SendChain(message.Text("ERROR: ", err))
				}

				// ascii2d 搜索
				// result, err := ascii2d.ASCII2d(pic)
				// if err != nil {
				// 	ctx.SendChain(message.Text("ERROR: ", err))
				// 	continue
				// }
				// msg := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text("ascii2d搜图结果"))}
				// for i := 0; i < len(result) && i < 5; i++ {
				// 	var resultMsgs message.Message
				// 	if showPic {
				// 		resultMsgs = append(resultMsgs, message.Image(result[i].Thumb))
				// 	}
				// 	resultMsgs = append(resultMsgs, message.Text(fmt.Sprintf(
				// 		"标题: %s\n图源: %s\n画师: %s\n画师链接: %s\n图片链接: %s",
				// 		result[i].Name,
				// 		result[i].Type,
				// 		result[i].AuthNm,
				// 		result[i].Author,
				// 		result[i].Link,
				// 	)))
				// 	msg = append(msg, ctxext.FakeSenderForwardNode(ctx, resultMsgs...))
				// }
				// if id := ctx.Send(msg).ID(); id == 0 {
				// 	ctx.SendChain(message.Text("ERROR: 可能被风控了"))
				// }
			}
		})
	engine.OnMessage(zero.NewPattern(nil).Command(`^(开启|打开|启用|关闭|关掉|禁用|恢复|重置)搜图显示图片$`).AsRule(), zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid <= 0 {
				// 个人用户设为负数
				gid = -ctx.Event.UserID
			}
			option := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			switch option {
			case "开启", "打开", "启用":
				err := engine.GetLevelDB().Put([]byte(fmt.Sprintf("saucenao_enable_pic_%d", gid)), []byte{1}, nil)
				if err != nil {
					log.Errorln("<saucenao>", err)
					return
				}
			case "关闭", "关掉", "禁用":
				err := engine.GetLevelDB().Put([]byte(fmt.Sprintf("saucenao_enable_pic_%d", gid)), []byte{0}, nil)
				if err != nil {
					log.Errorln("<saucenao>", err)
					return
				}
			case "恢复", "重置":
				err := engine.GetLevelDB().Delete([]byte(fmt.Sprintf("saucenao_enable_pic_%d", gid)), nil)
				if err != nil {
					log.Errorln("<saucenao>", err)
					return
				}
			default:
				return
			}
			ctx.SendChain(message.Text("已", option, "搜图显示图片"))
		})
}
