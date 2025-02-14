package deer_pipe

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"strconv"
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
	{cmd}鹿：签到和日历
	{cmd}鹿历：日历
把 鹿 替换成 鹿的emoji表情也可以触发
`,
	Classify: "怪东西",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"🦌历"}, zero.OnlyToMe).SetBlock(true).SetPriority(8).Handle(calenderHandle)
	proxy.OnCommands([]string{"🦌"}, zero.OnlyToMe).SetBlock(true).SetPriority(7).Handle(deerHandle)
}

func calenderHandle(ctx *zero.Ctx) {
	calenderThisMonth, err := getCount(proxy.GetDB(), ctx.Event.UserID)
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

	calenderImage, err := createCalenderImage(calenderThisMonth, headerImg)
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
	calenderThisMonth, err := addCount(proxy.GetDB(), ctx.Event.UserID)
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

	calenderImage, err := createCalenderImage(calenderThisMonth, headerImg)
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

func createCalenderImage(days []dao.DeerPipeCalender, header image.Image) (image.Image, error) {
	// 获取本月第一天是星期几
	now := time.Now()
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
	dateStr := time.Now().Format("2006-01-02 签到日历")
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

func getCount(db *gorm.DB, id int64) (calenderThisMonth []dao.DeerPipeCalender, err error) {
	if err = db.Where("qq=? and year = ? and month = ?", id, time.Now().Year(), time.Now().Month()).Find(&calenderThisMonth).Error; err != nil {
		return
	}
	if len(calenderThisMonth) == 0 { // 如果没查到就创建本月所有天数
		now := time.Now()
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
	}
	for i := range calenderThisMonth {
		db.Save(&calenderThisMonth[i])
	}
	return
}

func addCount(db *gorm.DB, id int64) (calenderThisMonth []dao.DeerPipeCalender, err error) {
	if err = db.Where("qq=? and year = ? and month = ?", id, time.Now().Year(), time.Now().Month()).Find(&calenderThisMonth).Error; err != nil {
		return
	}
	if len(calenderThisMonth) == 0 { // 如果没查到就创建本月所有天数
		now := time.Now()
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
	}
	for i := range calenderThisMonth {
		if calenderThisMonth[i].Day == time.Now().Day() {
			calenderThisMonth[i].Count++
		}
		db.Save(&calenderThisMonth[i])
	}
	return
}
