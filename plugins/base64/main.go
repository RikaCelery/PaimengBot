// Package base64gua base64卦 与 tea 加解密
package base64

import (
	"encoding/base64"

	"github.com/FloatTech/floatbox/crypto"
	"github.com/RicheyJang/PaimengBot/manager"
	base14 "github.com/fumiama/go-base16384"
	"github.com/fumiama/unibase2n"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

var info = manager.PluginInfo{
	Name: "base64",
	Usage: `
用法：
[base16384加解密]
	{cmd}加密xxx
	{cmd}解密xxx
	{cmd}用yyy加密xxx
	{cmd}用yyy解密xxx
[base64加解密]
	{cmd}b64加密xxx
	{cmd}b64解密xxx
[六十四卦加解密]
	{cmd}六十四卦加密xxx
	{cmd}六十四卦解密xxx
	{cmd}六十四卦用yyy加密xxx
	{cmd}六十四卦用yyy解密xxx
[天城文加解密]
	{cmd}天城文加密xxx
	{cmd}天城文解密xxx
	{cmd}天城文用yyy加密xxx
	{cmd}天城文用yyy解密xxx`,
	Brief:    `base*及其衍生算法`,
	Classify: "dev",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	proxy.OnMessage(zero.NewPattern(nil).Command(`^加密\s*(.+)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			es := base14.EncodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^解密\s*([一-踀]+[㴁-㴆]?)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			es := base14.DecodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^用(.+)加密\s*(.+)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1], ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[2]
			t := crypto.GetTEA(key)
			es, err := base14.UTF16BE2UTF8(base14.Encode(t.Encrypt(helper.StringToBytes(str))))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(es)))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^用(.+)解密\s*([一-踀]+[㴁-㴆]?)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1], ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[2]
			t := crypto.GetTEA(key)
			es, err := base14.UTF82UTF16BE(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(t.Decrypt(base14.Decode(es)))))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})

	proxy.OnMessage(zero.NewPattern(nil).Command(`^b64加密\s*(.+)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			es := base64.StdEncoding.EncodeToString([]byte(str))
			ctx.SendChain(message.Text(es))
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^b64解密\s*([a-zA-Z0-9=\-/+]+)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			es, err := base64.StdEncoding.DecodeString(str)
			if err == nil {
				ctx.SendChain(message.Text(es))
				return
			}
			es, err = base64.URLEncoding.DecodeString(str)
			if err == nil {
				ctx.SendChain(message.Text(es))
				return
			}
			ctx.SendChain(message.Text("解密失败!"))
		})

	proxy.OnMessage(zero.NewPattern(nil).Command(`^六十四卦加密\s*(.+)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			es := unibase2n.Base64Gua.EncodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^六十四卦解密\s*([䷀-䷿]+[☰☱]?)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			es := unibase2n.Base64Gua.DecodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^六十四卦用(.+)加密\s*(.+)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1], ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[2]
			t := crypto.GetTEA(key)
			es, err := unibase2n.UTF16BE2UTF8(unibase2n.Base64Gua.Encode(t.Encrypt(helper.StringToBytes(str))))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(es)))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^六十四卦用(.+)解密\s*([䷀-䷿]+[☰☱]?)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1], ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[2]
			t := crypto.GetTEA(key)
			es, err := unibase2n.UTF82UTF16BE(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(t.Decrypt(unibase2n.Base64Gua.Decode(es)))))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})

	proxy.OnMessage(zero.NewPattern(nil).Command(`^天城文加密\s*(.+)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			es := unibase2n.BaseDevanagari.EncodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^天城文解密\s*([ऀ-ॿ]+[০-৫]?)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1]
			es := unibase2n.BaseDevanagari.DecodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^天城文用(.+)加密\s*(.+)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1], ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[2]
			t := crypto.GetTEA(key)
			es, err := unibase2n.UTF16BE2UTF8(unibase2n.BaseDevanagari.Encode(t.Encrypt(helper.StringToBytes(str))))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(es)))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	proxy.OnMessage(zero.NewPattern(nil).Command(`^天城文用(.+)解密\s*([ऀ-ॿ]+[০-৫]?)$`).AsRule()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[1], ctx.State[zero.KeyPattern].([]zero.PatternParsed)[0].Text()[2]
			t := crypto.GetTEA(key)
			es, err := unibase2n.UTF82UTF16BE(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(t.Decrypt(unibase2n.BaseDevanagari.Decode(es)))))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
}
