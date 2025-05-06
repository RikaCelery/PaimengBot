package mention_reminder

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/basic/history"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	info = manager.PluginInfo{
		Name: "谁@我了",
		Usage: `用法：
	{cmd}谁@我了：
	{cmd}谁回我了：`,
		SuperUsage:  "",
		Brief:       "快速查看最近提到你的消息",
		Classify:    "",
		IsPassive:   false,
		IsSuperOnly: false,
		IsHidden:    false,
		AdminLevel:  0,
	}
	proxy *manager.PluginProxy
)

func init() {
	proxy = manager.RegisterPlugin(info)
	proxy.OnCommands([]string{"谁@我了", "谁at我了", "谁@我", "谁at我"}).SetBlock(true).Handle(handleMention)
	proxy.OnCommands([]string{"谁回我", "谁回我了"}).SetBlock(true).Handle(handleMention)
}
func getReply(msgs message.Message, gid int64, user int64) (history.SimpleEvent, bool) {
	for _, msg := range msgs {
		if msg.Type == "reply" {
			id, err := strconv.ParseInt(msg.Data["id"], 10, 64)
			if err != nil || id == 0 {
				continue
			}
			getId, err := history.GetId(gid, id)
			if err != nil || getId.UserID != user {
				logrus.Warnln("<mention_reminder>获取消息失败", err)
				continue
			}
			return getId, true
		}
	}
	return history.SimpleEvent{}, false
}
func containsAt(msgs message.Message, user int64) bool {
	for _, msg := range msgs {
		if msg.Type == "at" && msg.Data["qq"] == strconv.FormatInt(user, 10) {
			return true
		}
	}
	return false
}

type msgRender struct {
	Event    history.SimpleEvent
	Nickname string
	Reply    string
}

func simpleFormatMsg(msgs message.Message) string {
	var sb strings.Builder
	for _, msg := range msgs {
		if msg.Type == "text" {
			sb.WriteString(msg.Data["text"])
		} else if msg.Type == "face" {
			sb.WriteString("[表情]")
		} else if msg.Type == "image" {
			sb.WriteString("[图片]")
		} else if msg.Type == "reply" {
			sb.WriteString("[回复]")
		} else if msg.Type == "at" {
			sb.WriteString(fmt.Sprintf("%v(%v)", msg.Data["name"], msg.Data["qq"]))
		} else if msg.Type == "face" {
			sb.WriteString("[表情]")
		} else if msg.Type == "xml" {
			sb.WriteString("[xml]")
		} else if msg.Type == "json" {
			sb.WriteString("[json]")
		} else if msg.Type == "node" {
			sb.WriteString("[node]")
		} else if msg.Type == "forward" {
			sb.WriteString("[转发]")
		} else if msg.Type == "video" {
			sb.WriteString("[视频]")
		} else if msg.Type == "music" {
			sb.WriteString("[音乐]")
		} else if msg.Type == "poke" {
			sb.WriteString("[戳一戳]")
		} else if msg.Type == "dice" {
			sb.WriteString("[骰子]")
		} else if msg.Type == "shake" {
			sb.WriteString("[窗口抖动]")
		}
	}
	return sb.String()
}
func handleMention(ctx *zero.Ctx) {
	today, err := history.GroupHistoryRecent(ctx.Event.GroupID, 8*time.Hour)
	if err != nil {
		ctx.Send("失败了...")
		logrus.Warn("<mention_reminder>获取群历史信息失败", err)
		return
	}
	var atMsg []msgRender
	for _, event := range today {
		if strings.HasPrefix(ctx.State["command"].(string), "谁回我") {
			if reply, b := getReply(event.Msg, ctx.Event.GroupID, ctx.Event.UserID); b {
				memberInfo := ctx.GetThisGroupMemberInfo(event.UserID, false)
				name := memberInfo.Get("card").String()
				atMsg = append(atMsg, msgRender{event, name, simpleFormatMsg(reply.Msg)})
			}
		} else if containsAt(event.Msg, ctx.Event.UserID) {
			memberInfo := ctx.GetThisGroupMemberInfo(event.UserID, false)
			name := memberInfo.Get("card").String()
			if name == "" {
				name = memberInfo.Get("nickname").String()
			}
			atMsg = append(atMsg, msgRender{event, name, ""})
		}
	}
	if len(atMsg) == 0 {
		if strings.HasPrefix(ctx.State["command"].(string), "谁回我") {
			ctx.Send("好像没人回你")
		} else {
			ctx.Send("好像没人@你")
		}
		return
	}
	tmpl, err := manager.GetStaticFile("mention_reminder/template.gohtml")
	if err != nil {
		ctx.Send("失败了...")
		logrus.Warn("<mention_reminder>获取模板失败", err)
		return
	}
	tmplData, err := io.ReadAll(tmpl)
	if err != nil {
		ctx.Send("失败了...")
		logrus.Warn("<mention_reminder>获取读取失败", err)
		return
	}
	bytes, err := utils.ScreenShotElementTemplateString(string(tmplData), ".wrapper", atMsg)
	if err != nil {
		ctx.Send("生成图片失败了...")
		logrus.Warn("<mention_reminder>生成图片失败", err)
		return
	}
	ctx.Send(message.ImageBytes(bytes))
}
