package sc

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"sync"
	"time"

	"github.com/FloatTech/rendercard"
	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/disintegration/imaging"

	"github.com/fogleman/gg"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type signInfo struct {
	id       int64
	name     string
	double   bool
	orgFavor float64
	addFavor float64
	bg       string
	orgCoin  float64
	addCoin  float64
	signDays int
	lastSign time.Time // 最近一次签到时间
}

func initPic(uid int64) (avatar []byte, err error) {
	avatar, err = client.GetBytesRetry("https://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(uid, 10)+"&s=640", 3)
	if err != nil {
		return
	}
	return avatar, nil
}
func drawScore17b2(a *signInfo) (img *images.ImageCtx, err error) {
	getAvatar, err := initPic(a.id)
	if err != nil {
		return
	}
	back, err := images.NetImage(proxy.GetConfigString("background"))
	if err != nil {
		return
	}

	bx, by := float64(back.Bounds().Dx()), float64(back.Bounds().Dy())

	sc := 1280 / bx
	var colors []color.RGBA

	canvas := images.NewImageCtx(1280, 1280*int(by)/int(bx))
	cw, ch := float64(canvas.Width()), float64(canvas.Height())

	sch := ch * 6 / 10

	var blurback, scbackimg, backshadowimg, avatarimg, avatarbackimg, avatarshadowimg, whitetext, blacktext image.Image
	wg := &sync.WaitGroup{}
	wg.Add(7)
	scback := images.NewImageCtx(canvas.Width(), canvas.Height())

	scback.ScaleAbout(sc, sc, cw/2, ch/2)
	scback.DrawImageAnchored(back, canvas.Width()/2, canvas.Height()/2, 0.5, 0.5)
	scback.Identity()

	colors = images.TakeColor(scback.Image(), 3)
	go func() {
		defer wg.Done()

		blurback = imaging.Blur(scback.Image(), 20)

		scbackimg = rendercard.Fillet(scback.Image(), 12)
	}()

	go func() {
		defer wg.Done()
		pureblack := images.NewImageCtx(canvas.Width(), canvas.Height())
		pureblack.SetRGBA255(0, 0, 0, 255)
		pureblack.Clear()

		shadow := images.NewImageCtx(canvas.Width(), canvas.Height())
		shadow.ScaleAbout(0.6, 0.6, cw-cw/3, ch/2)
		shadow.DrawImageAnchored(pureblack.Image(), canvas.Width()-canvas.Width()/3, canvas.Height()/2, 0.5, 0.5)
		shadow.Identity()

		backshadowimg = imaging.Blur(shadow.Image(), 12)
	}()

	aw, ah := (ch-sch)/2/2/2*3, (ch-sch)/2/2/2*3

	go func() {
		defer wg.Done()
		avatar, _, err := image.Decode(bytes.NewReader(getAvatar))
		if err != nil {
			return
		}

		isc := (ch - sch) / 2 / 2 / 2 * 3 / float64(avatar.Bounds().Dy())

		scavatar := gg.NewContext(int(aw), int(ah))

		scavatar.ScaleAbout(isc, isc, aw/2, ah/2)
		scavatar.DrawImageAnchored(avatar, scavatar.Width()/2, scavatar.Height()/2, 0.5, 0.5)
		scavatar.Identity()

		avatarimg = rendercard.Fillet(scavatar.Image(), 8)
	}()

	err = canvas.UseDefaultFont((ch - sch) / 2 / 2 / 2)
	if err != nil {
		return
	}
	namew, _ := canvas.MeasureString(a.name)

	go func() {
		defer wg.Done()
		avatarshadowimg = imaging.Blur(customrectangle(cw, ch, aw, ah, namew, color.Black), 8)
	}()

	go func() {
		defer wg.Done()
		avatarbackimg = customrectangle(cw, ch, aw, ah, namew, colors[0])
	}()

	go func() {
		defer wg.Done()
		whitetext, err = customtext(a, cw, ch, aw, color.White)
		if err != nil {
			return
		}
	}()

	go func() {
		defer wg.Done()
		blacktext, err = customtext(a, cw, ch, aw, color.Black)
		if err != nil {
			return
		}
	}()

	wg.Wait()
	if scbackimg == nil || backshadowimg == nil || avatarimg == nil || avatarbackimg == nil || avatarshadowimg == nil || whitetext == nil || blacktext == nil {
		err = errors.New("图片渲染失败")
		return
	}

	canvas.DrawImageAnchored(blurback, canvas.Width()/2, canvas.Height()/2, 0.5, 0.5)

	canvas.DrawImage(backshadowimg, 0, 0)

	canvas.ScaleAbout(0.6, 0.6, cw-cw/3, ch/2)
	canvas.DrawImageAnchored(scbackimg, canvas.Width()-canvas.Width()/3, canvas.Height()/2, 0.5, 0.5)
	canvas.Identity()

	canvas.DrawImage(avatarshadowimg, 0, 0)
	canvas.DrawImage(avatarbackimg, 0, 0)
	canvas.DrawImageAnchored(avatarimg, int((ch-sch)/2/2), int((ch-sch)/2/2), 0.5, 0.5)

	canvas.DrawImage(blacktext, 2, 2)
	canvas.DrawImage(whitetext, 0, 0)

	img = canvas
	return
}

func customrectangle(cw, ch, aw, ah, namew float64, rtgcolor color.Color) (img image.Image) {
	canvas := gg.NewContext(int(cw), int(ch))
	sch := ch * 6 / 10
	canvas.DrawRoundedRectangle((ch-sch)/2/2-aw/2-aw/40, (ch-sch)/2/2-aw/2-ah/40, aw+aw/40*2, ah+ah/40*2, 8)
	canvas.SetColor(rtgcolor)
	canvas.Fill()
	canvas.DrawRoundedRectangle((ch-sch)/2/2, (ch-sch)/2/2-ah/4, aw/2+aw/40*5+namew, ah/2, 8)
	canvas.Fill()

	img = canvas.Image()
	return
}

func customtext(a *signInfo, cw, ch, aw float64, textcolor color.Color) (img image.Image, err error) {
	canvas := images.NewImageCtx(int(cw), int(ch))
	canvas.SetColor(textcolor)
	scw, sch := cw*6/10, ch*6/10
	err = canvas.UseDefaultFont((ch - sch) / 2 / 2 / 2)
	if err != nil {
		return
	}
	canvas.DrawStringAnchored(a.name, (ch-sch)/2/2+aw/2+aw/40*2, (ch-sch)/2/2-5, 0, 0.5)
	err = canvas.UseDefaultFont((ch - sch) / 2 / 2 / 3 * 2)
	if err != nil {
		return
	}
	canvas.DrawStringAnchored(time.Now().Format("2006/01/02"), cw-cw/6, ch/2-sch/2-canvas.FontHeight(), 0.5, 0.5)

	err = canvas.UseDefaultFont((ch - sch) / 2 / 2 / 2)
	if err != nil {
		return
	}

	level, up := LevelAt(a.orgFavor + a.addFavor)
	nextLevelStyle := fmt.Sprintf("还需%.2f", up)

	canvas.DrawStringAnchored("Level "+strconv.Itoa(level), cw/3*2-scw/2, ch/2+sch/2+canvas.FontHeight(), 0, 0.5)
	canvas.DrawStringAnchored(nextLevelStyle, cw/3*2+scw/2, ch/2+sch/2+canvas.FontHeight(), 1, 0.5)

	err = canvas.UseDefaultFont((ch - sch) / 2 / 2 / 3)
	if err != nil {
		return
	}

	canvas.DrawStringAnchored("Create By ZeroBot-Plugin ", 0+4, ch, 0, -0.5)

	err = canvas.UseDefaultFont((ch - sch) / 2 / 5 * 3)
	if err != nil {
		return
	}

	tempfh := canvas.FontHeight()

	canvas.DrawStringAnchored(getHourWord(time.Now()), ((cw-scw)-(cw/3-scw/2))/8, (ch-sch)/2+sch/4, 0, 0.5)

	err = canvas.UseDefaultFont((ch - sch) / 2 / 5)
	if err != nil {
		return
	}

	canvas.DrawStringAnchored("+ "+strconv.FormatFloat(RealCoin(a.addCoin), 'f', 2, 64)+Unit(), ((cw-scw)-(cw/3-scw/2))/8, (ch-sch)/2+sch/4+tempfh, 0, 0.5)
	canvas.DrawStringAnchored("+ "+strconv.FormatFloat(RealCoin(a.addFavor), 'f', 2, 64)+"好感", ((cw-scw)-(cw/3-scw/2))/8, (ch-sch)/2+sch/4+tempfh+canvas.FontHeight(), 0, 1)

	err = canvas.UseDefaultFont((ch - sch) / 2 / 4)
	if err != nil {
		return
	}

	canvas.DrawStringAnchored("你有 "+strconv.FormatFloat(RealCoin(a.orgCoin+a.addCoin), 'f', 2, 64)+" 枚"+Unit(), ((cw-scw)-(cw/3-scw/2))/8, (ch-sch)/2+sch/4*3, 0, 0.5)

	img = canvas.Image()
	return
}

func getHourWord(t time.Time) string {
	h := t.Hour()
	switch {
	case 6 <= h && h < 12:
		return "早上好"
	case 12 <= h && h < 14:
		return "中午好"
	case 14 <= h && h < 19:
		return "下午好"
	case 19 <= h && h < 24:
		return "晚上好"
	case 0 <= h && h < 6:
		return "凌晨好"
	default:
		return ""
	}
}
func (s signInfo) genMessageZbp() (message.Message, *images.ImageCtx, error) {
	img, err := drawScore17b2(&s)
	if err != nil {
		return nil, nil, err
	}
	// 生成消息
	imgMsg, err := img.GenMessageAuto()
	if err != nil {
		return nil, nil, err
	}
	return message.Message{imgMsg}, img, nil
}

func (s signInfo) String() string {
	var doubleStr string
	if s.double {
		doubleStr = "✪ ω ✪ 双倍！\n"
	}
	return fmt.Sprintf("%s已连续签到%v天\n好感度：%.2f(+%.2f)\n财富：%.0f(+%.0f)%s",
		doubleStr,
		s.signDays,
		s.orgFavor+s.addFavor, s.addFavor,
		RealCoin(s.orgCoin+s.addCoin), RealCoin(s.addCoin), Unit())
}

func genRankMessage(ctx *zero.Ctx, users []dao.UserOwn, key string) (msg message.Segment, err error) {
	var values []images.UserValue
	if key == "favor" { // 好感度
		for _, user := range users {
			values = append(values, images.UserValue{
				ID:      user.ID,
				Value:   user.Favor,
				FmtPrec: 2,
			})
		}
		return images.GenQQRankMsgWithValue("好感度排行榜", values, "")
	} else { // 财富
		for _, user := range users {
			values = append(values, images.UserValue{
				ID:    user.ID,
				Value: RealCoin(user.Wealth),
			})
		}
		return images.GenQQRankMsgWithValue("财富排行榜", values, Unit())
	}
}
