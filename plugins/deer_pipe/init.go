package deer_pipe

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "鹿管签到",
	Usage: `
一个鹿管签到小插件
用法：
	{cmd}鹿：今日签到和日历
	{cmd}鹿历 [月份参数]?：日历，如果有参数就是指定月份的日历
	{cmd}补鹿 [日期]：补签
		关于[日期]：格式要求非常宽松，以下格式都可以支持
			1. 昨天、前天、大前天、大大前天
			2. 周一/星期一/上周一/上星期一、周二/星期二/上周二/上星期二、...、周日/星期日/上周日/上星期日
			3. 4号/4日/二十号/二十日
			4. 3月3号/3月3日/五月3日/6月十三号
			5. 2024-1-12、2024年5月20日
把 鹿 替换成 鹿的emoji表情也可以触发
`,
	Classify: "怪东西",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"🦌历", "鹿历"}, zero.OnlyToMe).SetBlock(true).SetPriority(8).Handle(calenderHandle)
	proxy.OnCommands([]string{"🦌", "鹿"}, zero.OnlyToMe).SetBlock(true).SetPriority(7).Handle(deerHandle)
	proxy.OnCommands([]string{"补🦌", "补鹿"}, zero.OnlyToMe).SetBlock(true).SetPriority(3).Handle(deerReSignHandle)
}

func calenderHandle(ctx *zero.Ctx) {
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	date, err := parseMonth(arg)
	if arg != "" && err != nil {
		log.Errorf("<deer_pipe>parseMonth err: %v", err)
		ctx.Send("格式错误，可以看看帮助: " + err.Error())
		return
	}
	if date == nil {
		t := time.Now()
		date = &t
	}

	calenderThisMonth, err := getCount(proxy.GetDB(), ctx.Event.UserID, date)
	if err != nil {
		log.Errorf("<deer_pipe>addCount err: %v", err)
		ctx.Send("出错惹")
		return
	}
	body, err := utils.GetQQAvatar(ctx.Event.UserID, 100)
	if err != nil {
		log.Errorf("<deer_pipe>get user avatar err: %v", err)
		ctx.Send("出错了...")
		return
	}
	headerImg, _, err := image.Decode(body)
	if err != nil {
		log.Errorf("<deer_pipe>get user avatar err: %v", err)
		ctx.Send("出错了...")
	}
	defer body.Close()

	calenderImage, err := createCalenderImage(calenderThisMonth, headerImg, *date)
	if err != nil {
		log.Errorf("<deer_pipe>create calender image err: %v", err)
		ctx.Send("出错惹")
		return
	}
	var buf = bytes.Buffer{}
	err = jpeg.Encode(&buf, calenderImage, nil)
	if err != nil {
		log.Errorf("<deer_pipe>jpeg encode err: %v", err)
		ctx.Send("出错惹")
		return
	}
	ctx.SendChain(message.ImageBytes(buf.Bytes()))
}

func deerHandle(ctx *zero.Ctx) {
	calenderThisMonth, err := addCount(proxy.GetDB(), ctx.Event.UserID, nil)
	if err != nil {
		log.Errorf("<deer_pipe>addCount err: %v", err)
		ctx.Send("出错惹")
		return
	}
	body, err := utils.GetQQAvatar(ctx.Event.UserID, 100)
	if err != nil {
		log.Errorf("<deer_pipe>get user avatar err: %v", err)
		ctx.Send("出错了...")
		return
	}
	headerImg, _, err := image.Decode(body)
	if err != nil {
		log.Errorf("<deer_pipe>get user avatar err: %v", err)
		ctx.Send("出错了...")
	}
	defer body.Close()

	calenderImage, err := createCalenderImage(calenderThisMonth, headerImg, time.Now())
	if err != nil {
		log.Errorf("<deer_pipe>create calender image err: %v", err)
		ctx.Send("出错惹")
		return
	}
	var buf = bytes.Buffer{}
	err = jpeg.Encode(&buf, calenderImage, nil)
	if err != nil {
		log.Errorf("<deer_pipe>jpeg encode err: %v", err)
		ctx.Send("出错惹")
		return
	}
	ctx.SendChain(message.ImageBytes(buf.Bytes()))
}
func deerReSignHandle(ctx *zero.Ctx) {
	date, err := getDeerArg(ctx)
	if err != nil {
		log.Errorf("<deer_pipe>getDeerArg err: %v", err)
		ctx.Send(err.Error())
		return
	}
	calenderThisMonth, err := addCount(proxy.GetDB(), ctx.Event.UserID, date)
	if err != nil {
		log.Errorf("<deer_pipe>addCount err: %v", err)
		ctx.Send("出错惹")
		return
	}
	body, err := utils.GetQQAvatar(ctx.Event.UserID, 100)
	if err != nil {
		log.Errorf("<deer_pipe>get user avatar err: %v", err)
		ctx.Send("出错了...")
		return
	}
	headerImg, _, err := image.Decode(body)
	if err != nil {
		log.Errorf("<deer_pipe>get user avatar err: %v", err)
		ctx.Send("出错了...")
	}
	defer body.Close()

	calenderImage, err := createCalenderImage(calenderThisMonth, headerImg, *date)
	if err != nil {
		log.Errorf("<deer_pipe>create calender image err: %v", err)
		ctx.Send("出错惹")
		return
	}
	var buf = bytes.Buffer{}
	err = jpeg.Encode(&buf, calenderImage, nil)
	if err != nil {
		log.Errorf("<deer_pipe>jpeg encode err: %v", err)
		ctx.Send("出错惹")
		return
	}
	ctx.SendChain(message.ImageBytes(buf.Bytes()))
}

func getDeerArg(ctx *zero.Ctx) (*time.Time, error) {
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	switch arg {
	case "昨天":
		date := time.Now().AddDate(0, 0, -1)
		return &date, nil
	case "前天":
		date := time.Now().AddDate(0, 0, -2)
		return &date, nil
	case "大前天":
		date := time.Now().AddDate(0, 0, -3)
		return &date, nil
	case "大大前天":
		date := time.Now().AddDate(0, 0, -4)
		return &date, nil
	case "星期一":
		fallthrough
	case "周一":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+1)
		if date.Sub(time.Now()).Seconds() > 0 {
			return nil, errors.New("还没到周一")
		}
		return &date, nil
	case "星期二":
		fallthrough
	case "周二":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+2)
		if date.Sub(time.Now()).Seconds() > 0 {
			return nil, errors.New("还没到周二")
		}
		return &date, nil
	case "星期三":
		fallthrough
	case "周三":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+3)
		if date.Sub(time.Now()).Seconds() > 0 {
			return nil, errors.New("还没到周三")
		}
		return &date, nil
	case "星期四":
		fallthrough
	case "周四":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+4)
		if date.Sub(time.Now()).Seconds() > 0 {
			return nil, errors.New("还没到周四")
		}
		return &date, nil
	case "星期五":
		fallthrough
	case "周五":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+5)
		if date.Sub(time.Now()).Seconds() > 0 {
			return nil, errors.New("还没到周五")
		}
		return &date, nil
	case "星期六":
		fallthrough
	case "周六":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+6)
		if date.Sub(time.Now()).Seconds() > 0 {
			return nil, errors.New("还没到周六")
		}
		return &date, nil
	case "星期日":
		fallthrough
	case "周日":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+7)
		if date.Sub(time.Now()).Seconds() > 0 {
			return nil, errors.New("还没到周日")
		}
	case "上星期一":
		fallthrough
	case "上周一":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())-7)
		return &date, nil
	case "上星期二":
		fallthrough
	case "上周二":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())-6)
		return &date, nil
	case "上星期三":
		fallthrough
	case "上周三":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())-5)
		return &date, nil
	case "上星期四":
		fallthrough
	case "上周四":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())-4)
		return &date, nil
	case "上星期五":
		fallthrough
	case "上周五":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())-3)
		return &date, nil
	case "上星期六":
		fallthrough
	case "上周六":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())-2)
		return &date, nil
	case "上星期日":
		fallthrough
	case "上周日":
		date := time.Now().AddDate(0, 0, -int(time.Now().Weekday())-1)
		return &date, nil
	default:
		// 解析日期 （需要允许可能出现的各种空格）
		// 一月三日 / 1月8号
		// 三号 / 三日 / 3号 5日
		// 2024年3月5日
		// 2024-3-5 / 2024/3/5
		date, _ := parseDate(arg)
		if date == nil {
			return nil, errors.New("无法识别的日期")
		} else {
			if date.Sub(time.Now()).Seconds() > 0 {
				return nil, errors.New("日子还没到呢")
			}
			return date, nil
		}

	}
	panic("unreachable")
}

func createCalenderImage(days []dao.DeerPipeCalender, header image.Image, now time.Time) (image.Image, error) {
	// 获取本月第一天是星期几
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	// 获取本月总天数
	totalDays := firstDay.AddDate(0, 1, -1).Day() - 1
	startGrid := int(firstDay.Weekday())
	blockSize := 100
	rows := (startGrid + totalDays + 6) / 7
	bannerHeight := 120
	ctx := images.NewImageCtx(7*blockSize, rows*blockSize+bannerHeight)
	ctx.SetColor(color.White)
	ctx.DrawRectangle(0, 0, float64(7*blockSize), float64(rows*blockSize))
	ctx.Fill()
	deerBg, err := manager.DecodeStaticImage("deer_pipe/deerpipe.jpg")
	if err != nil {
		return nil, err
	}
	size := float64(blockSize)
	ctx.SetColor(color.White)
	ctx.Clear()
	padding := 1
	paddingF := float64(padding)
	deerBg = resize.Resize(uint(int(size)-2*padding), 0, deerBg, resize.Lanczos3)
	// banner
	header = resize.Resize(100, 100, header, resize.Lanczos3)
	ctx.DrawImage(header, 10, 10)
	dateStr := now.Format("2006-01-02 签到日历")
	w, h := images.MeasureStringDefault(dateStr, 20, 1)
	ctx.UseDefaultFont(20)
	_ = ctx.PasteStringDefault(dateStr, 20, 1, 120, 10+h/2, w)
	// calender
	for i := 0; i < rows; i++ {
		for j := 0; j < 7; j++ {
			dayNum := i*7 + j - startGrid // 当天的日期（从0开始）
			if i == 0 && j < startGrid {
				continue
			} else if dayNum > totalDays {
				continue
			} else {
				ctx.DrawImage(deerBg, j*blockSize+padding, i*blockSize+padding+bannerHeight)
				// ctx.DrawRectangle(float64(j)*size+paddingF, float64(i)*size+paddingF, size-2*paddingF, size-2*paddingF)
				// ctx.Fill()
				ctx.SetColor(color.Black)
				w, h := images.MeasureStringDefault(strconv.FormatInt(int64(dayNum+1), 10), 24, 1.3)
				_ = ctx.PasteStringDefault(strconv.FormatInt(int64(dayNum+1), 10), 24, 1.2, float64(j)*size+paddingF, float64(i+1)*size-paddingF*2-h*1.3+float64(bannerHeight), w)
				if days[dayNum].Count > 0 {
					str := fmt.Sprintf("x%d", days[dayNum].Count)
					w, h := images.MeasureStringDefault(str, 24, 1.3)
					ctx.SetRGB(0, 0, 0)
					ctx.UseDefaultFont(24)
					for iter := 0; iter < 100; iter++ {
						smoothness := 4
						angle := float64(iter) / float64(smoothness) * 2 * math.Pi
						maxOffset := 2.0
						xOffset := math.Cos(angle) * maxOffset
						yOffset := math.Sin(angle) * maxOffset
						ctx.DrawStringWrapped(str, float64(j)*size+paddingF+xOffset, (float64(i)+0.5)*size-paddingF*2-h+yOffset+float64(bannerHeight), 0, 0, w, 1.3, gg.AlignLeft)
					}
					ctx.SetRGB(1, 0, 0)
					ctx.DrawStringWrapped(str, float64(j)*size+paddingF, (float64(i)+0.5)*size-paddingF*2-h+float64(bannerHeight), 0, 0, w, 1.3, gg.AlignLeft)
				}
			}
		}
	}
	return ctx.Image(), err
}

func getCount(db *gorm.DB, id int64, date *time.Time) (calenderThisMonth []dao.DeerPipeCalender, err error) {
	t := time.Now()
	if date != nil {
		t = *date
	}
	if err = db.Where("qq=? and year = ? and month = ?", id, t.Year(), t.Month()).Find(&calenderThisMonth).Error; err != nil {
		return
	}
	if len(calenderThisMonth) == 0 { // 如果没查到就创建本月所有天数
		now := t
		year, month, _ := now.Date()
		firstDay := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
		lastDay := firstDay.AddDate(0, 1, -1)

		for day := firstDay; !day.After(lastDay); day = day.AddDate(0, 0, 1) {
			calenderThisMonth = append(calenderThisMonth, dao.DeerPipeCalender{
				QQ:    id,
				Year:  day.Year(),
				Month: int(day.Month()),
				Day:   day.Day(),
				Count: 0,
			})
		}
		db.Save(&calenderThisMonth)
	}
	return
}

func addCount(db *gorm.DB, id int64, date *time.Time) (calenderThisMonth []dao.DeerPipeCalender, err error) {
	now := time.Now()
	if date != nil {
		now = *date
		log.Infoln(date)
	}
	if err = db.Where("qq = ? and year = ? and month = ?", id, now.Year(), now.Month()).Find(&calenderThisMonth).Error; err != nil {
		return
	}
	if len(calenderThisMonth) == 0 { // 如果没查到就创建本月所有天数
		year, month, _ := now.Date()
		firstDay := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
		lastDay := firstDay.AddDate(0, 1, -1)

		for day := firstDay; !day.After(lastDay); day = day.AddDate(0, 0, 1) {
			calender := dao.DeerPipeCalender{
				QQ:    id,
				Year:  day.Year(),
				Month: int(day.Month()),
				Day:   day.Day(),
				Count: 0,
			}
			calenderThisMonth = append(calenderThisMonth, calender)
			db.Create(&calender)

		}
	}
	for i := range calenderThisMonth {
		if calenderThisMonth[i].Day == now.Day() {
			calenderThisMonth[i].Count++
			db.Save(&calenderThisMonth[i])
		}
	}
	return
}
