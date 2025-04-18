package ban

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/clause"

	"github.com/RicheyJang/PaimengBot/basic/auth"
	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "功能开关",
	Usage: `
用法：
	{cmd}开启[功能] [时长]?：将开启本群的指定功能，时长为可选项，形式参照示例
	{cmd}关闭[功能] [时长]?：将关闭本群的指定功能，时长为可选项，形式参照示例
	{cmd}全部开启[时长]?：将开启本群的全部功能，时长为可选项，形式参照示例
	{cmd}全部关闭[时长]?：将关闭本群的全部功能，时长为可选项，形式参照示例
	{cmd}重置功能状态：恢复为拉群时的初始状态（黑名单模式，无禁用功能（被主人全局关闭的功能不算））
	{cmd}封禁[qq号] [功能]? [时长]?：封禁指定用户使用指定功能（当指定功能时）或全部功能，时长为可选项，形式参照示例
	{cmd}解封[qq号] [功能]?：解封指定用户使用指定功能，时长为可选项，形式参照示例
	{cmd}黑名单：获取所有被封禁用户的被封禁功能列表
	{cmd}白名单模式：只能运行开启的功能
	{cmd}黑名单模式：无法运行关闭的功能
示例：
	{cmd}封禁123456：封禁用户ID为123456的所有功能
	{cmd}封禁123456 25m：封禁用户ID为123456的所有功能25分钟
	{cmd}封禁123456 翻译 1h30m：封禁用户ID123456的翻译功能1小时零30分钟`,

	SuperUsage: `
用法：
	{cmd}白名单：获取白名单运行的群/用户
	{cmd}设置插件白名单[功能名] [群号]+：讲这个功能的运行模式改为白名单，只允许特定群使用
	{cmd}取消/删除/移除插件白名单[功能名] [群号]+：删除若干群号，若群号为all则全部删除，为空时自动恢复默认运行模式
		如果需要设置私聊的白名单，群号需要为负数qq号 -[qq号]
	在私聊中：
		使用{cmd}开启\关闭[功能] [时长]?命令，将针对所有用户和群开启\关闭该功能（全局Ban）
		还可通过 {cmd}开启\关闭[群ID] [功能] [时长]? 来开启\关闭指定群的指定功能
		{cmd}黑名单：获取所有被封禁用户、群的被封禁功能列表
	在群聊中，等同于最高权限群管理员执行命令
config-plugin配置项：
	ban.tip: 调用某项被禁用的功能时，是(true)否(false)提示"该功能已被禁用"，但不会提示个人封禁`,
	AdminLevel: 1,
	Classify:   "内置功能",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"开启"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(openPlugin)
	proxy.OnCommands([]string{"关闭"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(closePlugin)
	proxy.OnCommands([]string{"全部开启"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(func(ctx *zero.Ctx) {
		switchAllPlugins(ctx, true)
	})
	proxy.OnCommands([]string{"全部关闭"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(func(ctx *zero.Ctx) {
		switchAllPlugins(ctx, false)
	})
	proxy.OnCommands([]string{"重置功能状态"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(resetPluginStatus)
	proxy.OnCommands([]string{"白名单模式"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(setModeWhite)
	proxy.OnCommands([]string{"黑名单模式"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(setModeBlack)
	proxy.OnCommands([]string{"设置插件白名单"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).FirstPriority().Handle(addPluginWhite)
	proxy.OnCommands([]string{"取消插件白名单", "删除插件白名单", "移除插件白名单"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).FirstPriority().Handle(removePluginWhite)
	proxy.OnCommands([]string{"封禁", "ban", "Ban"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(banUser)
	proxy.OnCommands([]string{"解封", "unban", "Unban"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(unbanUser)
	proxy.OnCommands([]string{"黑名单"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(showBlack)
	// 仅超级用户可以使用白名单查看某个白名单模式群的启用功能
	proxy.OnCommands([]string{"白名单"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).FirstPriority().Handle(showWhite)
	proxy.AddConfig("tip", false)
	manager.AddPreHook(checkPluginStatus)
}

func removePluginWhite(ctx *zero.Ctx) {

	args := strings.Split(strings.TrimSpace(utils.GetArgs(ctx)), " ")
	if len(args) < 2 {
		ctx.Send("参数不够哦，可以参考一下帮助")
		return
	}
	plugin := findPluginByName(args[0])
	if plugin == nil {
		ctx.Send("未找到该功能")
		return
	}
	args = args[1:]
	var groups []string
	for _, group := range args {
		if group == "all" {
			groups = []string{}
			break
		}
		id, err := strconv.ParseInt(group, 10, 64)
		if err != nil {
			ctx.Send("参数错误" + err.Error())
			return
		}
		groups = append(groups, strconv.FormatInt(id, 10))
	}
	var list = dao.PluginWhiteList{
		PluginKey: plugin.Key,
	}
	if err := proxy.GetDB().Find(&list).Error; err != nil {
		log.Errorln("<ban> query error", err)
		ctx.Send("数据库错误" + err.Error())
		return
	}
	list.GroupID = strings.Join(utils.MergeStringSlices(groups), "|")
	if len(list.GroupID) == 0 {
		if err := proxy.GetDB().Delete(&list).Error; err != nil {
			log.Errorf("set plugin(%v) white list error(sql): %v", plugin.Key, err)
			// return err
		}
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plugin_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_id"}), // Upsert
	}).Create(&list).Error; err != nil {
		log.Errorf("set plugin(%v) white list error(sql): %v", plugin.Key, err)
	}
	if list.GroupID == "" {
		ctx.Send("好哒，取消了白名单模式")
	} else {
		ctx.Send("好哒")
	}
}

func addPluginWhite(ctx *zero.Ctx) {
	args := strings.Split(strings.TrimSpace(utils.GetArgs(ctx)), " ")
	if len(args) < 2 {
		ctx.Send("参数不够哦，可以参考一下帮助")
		return
	}
	plugin := findPluginByName(args[0])
	if plugin == nil {
		ctx.Send("未找到该功能")
		return
	}
	args = args[1:]
	var groups []string
	for _, group := range args {
		id, err := strconv.ParseInt(group, 10, 64)
		if err != nil {
			ctx.Send("参数错误" + err.Error())
			return
		}
		groups = append(groups, strconv.FormatInt(id, 10))
	}
	var list = dao.PluginWhiteList{
		PluginKey: plugin.Key,
	}
	if err := proxy.GetDB().FirstOrCreate(&list).Error; err != nil {
		log.Errorln("<ban> query error", err)
		ctx.Send("数据库错误" + err.Error())
		return
	}
	list.GroupID = strings.Join(utils.MergeStringSlices(groups, strings.Split(list.GroupID, "|")), "|")
	log.Infoln("<ban> Saving PluginWhiteList with PluginKey:", list.PluginKey, "Groups:", list.GroupID)
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plugin_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_id"}), // Upsert
	}).Create(&list).Error; err != nil {
		log.Errorf("set plugin(%v) white list error(sql): %v", plugin.Key, err)
	}
	ctx.Send("好哒")
}

const AllPluginKey = "all"
const TipContent = "该功能已被禁用"

func checkPluginStatus(condition *manager.PluginCondition, ctx *zero.Ctx) error {
	// 群ban
	if ctx.Event.GroupID != 0 && !GetGroupPluginStatus(ctx.Event.GroupID, condition) {
		if proxy.GetConfigBool("tip") {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(TipContent))
		}
		return fmt.Errorf("此插件<%v>在此群(%v)已被关闭", condition.Key, ctx.Event.GroupID)
	}
	// 个人ban
	if ctx.Event.UserID != 0 && !GetUserPluginStatus(ctx.Event.UserID, condition) {
		// 不去理会被封禁的个人，因此不提示已被封禁
		return fmt.Errorf("此插件<%v>对此用户(%v)已被禁用", condition.Key, ctx.Event.UserID)
	}
	// 全局ban
	if !GetUserPluginStatus(0, condition) {
		if proxy.GetConfigBool("tip") {
			ctx.Send(message.Text(TipContent))
		}
		return fmt.Errorf("此插件<%v>已全局禁用", condition.Key)
	}
	// 白名单插件
	id := ctx.Event.GroupID
	if id == 0 {
		id = -ctx.Event.UserID
	}
	if !CheckPluginWhiteList(condition, id) {
		return fmt.Errorf("群(%d)不在此插件<%v>白名单内", ctx.Event.GroupID, condition.Key)
	}
	return nil
}

func resetPluginStatus(ctx *zero.Ctx) {
	if utils.IsMessageGroup(ctx) {
		if !auth.CheckPriority(ctx, 5, false) {
			return
		}
		for _, plugin := range manager.GetAllPluginConditions() {
			err := SetGroupPluginStatus(true, ctx.Event.GroupID, plugin, 0)
			if err != nil {
				log.Errorf("resetPluginStatus err: %v", err)
				ctx.Send("失败了...")
				return
			}
		}
		ctx.Send("好哒")
	} else if utils.IsMessagePrimary(ctx) {
		if !utils.IsSuperUser(ctx.Event.UserID) {
			ctx.Send("仅超级用户可以执行此命令")
			return
		}
		setModeBlack(ctx)
		for _, plugin := range manager.GetAllPluginConditions() {
			err := SetUserPluginStatus(true, 0, plugin, 0)
			if err != nil {
				log.Errorf("resetPluginStatus err: %v", err)
				ctx.Send("失败了...")
				return
			}
		}
		ctx.Send("好哒")
	}
}
func dealUserPluginStatus(ctx *zero.Ctx, status bool, userID int64, plugin *manager.PluginCondition, period time.Duration) {
	if status == GetUserPluginStatus(userID, plugin) {
		ctx.Send("请不要重复开关功能哦")
		return
	}
	err := SetUserPluginStatus(status, userID, plugin, period)
	if err != nil {
		ctx.Send("失败了...")
	} else {
		ctx.Send("好哒")
	}
}
func dealUserAllPluginStatus(ctx *zero.Ctx, status bool, userID int64, period time.Duration) {
	for _, plugin := range manager.GetAllPluginConditions() {
		if status == GetUserPluginStatus(userID, plugin) {
			continue
		}
		if !status && isImportant(plugin) {
			// 避免全部关闭后无法打开
			continue
		}
		err := SetUserPluginStatus(status, userID, plugin, period)
		if err != nil {
			log.Errorf("switchPlugin err: %v", err)
			ctx.Send("失败了...")
			return
		}
	}
	ctx.Send("好哒")

}

func isImportant(plugin *manager.PluginCondition) bool {
	if plugin.IsHidden {
		return true
	}
	keys := []string{
		"auth",
		"ban",
		"event",
		"help",
		"invite",
		"autowithdraw",
		"limiter",
	}
	for _, key := range keys {
		if plugin.Key == key {
			return true
		}
	}
	return false
}

func dealGroupPluginStatus(ctx *zero.Ctx, status bool, groupID int64, plugin *manager.PluginCondition, period time.Duration) {
	if status == GetGroupPluginStatus(groupID, plugin) {
		ctx.Send("请不要重复开关功能哦")
		return
	}
	err := SetGroupPluginStatus(status, groupID, plugin, period)
	if err != nil {
		ctx.Send("失败了...")
	} else {
		ctx.Send("好哒")
	}
}
func dealGroupAllPluginStatus(ctx *zero.Ctx, status bool, groupID int64, period time.Duration) {
	for _, plugin := range manager.GetAllPluginConditions() {
		if status == GetGroupPluginStatus(groupID, plugin) {
			continue
		}
		if !status && isImportant(plugin) {
			// 避免全部关闭后无法打开
			continue
		}
		err := SetGroupPluginStatus(status, groupID, plugin, period)
		if err != nil {
			log.Errorf("switchPlugin err: %v", err)
			ctx.Send("失败了...")
			return
		}
	}
	ctx.Send("好哒")
}

func hasPluginKey(org, key string) bool {
	return strings.Contains(org, "|"+key+"|")
}

func addPluginKey(org, key string) string {
	if len(org) == 0 || !strings.HasSuffix(org, "|") {
		org += "|"
	}
	if !hasPluginKey(org, key) {
		return org + key + "|"
	}
	return org
}

func delPluginKey(org, key string) string {
	name := fmt.Sprintf("|%s|", key)
	return strings.ReplaceAll(org, name, "|")
}

// 通过插件名或Key查找插件Condition
func findPluginByName(name string) *manager.PluginCondition {
	return manager.Lookup(name)
}
