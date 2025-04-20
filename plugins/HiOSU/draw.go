package HiOSU

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	math2 "github.com/FloatTech/floatbox/math"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/nfnt/resize"
)

func netImage(URL string) (image.Image, error) {
	c := client.NewHttpClient(&client.HttpOptions{TryTime: 3})
	reader, err := c.GetReader(URL)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	decode, _, err := image.Decode(reader)
	return decode, err
}
func drawUserInfo(user ApiUser, recent []Score, best []Score, mode string) (*images.ImageCtx, error) {
	c := images.NewImageCtxWithBGColor(1260, 1267, "#ffffff")
	regular, err := images.ParseFont("ttf/Torus Regular.ttf")
	if err != nil {
		return nil, err
	}
	semibold, err := images.ParseFont("ttf/Torus Semi Bold.ttf")
	if err != nil {
		return nil, err
	}
	// fbg, _ := os.Open("o.png")
	// bg, _, _ := image.Decode(fbg)
	// bg = imaging.AdjustFunc(bg, func(c color.NRGBA) color.NRGBA {
	// 	return color.NRGBA{
	// 		R: c.R,
	// 		G: c.G,
	// 		B: c.B,
	// 		A: uint8(float64(c.A) * 0.7),
	// 	}
	// })
	// shadows
	xStart := 70.0
	// xStart := 735.0
	{
		size := 8.0
		// profile
		depth := 4.0
		// c.ShadowRounded("#00000038", 44, 34, 663, 1180, 10, size, 3, depth, depth)
		c.ShadowRounded("#00000038", int(xStart), 34, 96, 96, 10, size, 3, depth, depth)
		c.ShadowRounded("#00000038", int(xStart), 143, 385, 47, 8, size, 3, depth, depth)
		c.ShadowRounded("#00000038", int(xStart), 200, 385, 290, 8, size, 3, depth, depth)
		c.ShadowRounded("#00000038", int(xStart+398), 143, 692, 87, 8, size, 3, depth, depth)
		c.ShadowRounded("#00000038", int(xStart+398), 240, 692, 87, 8, size, 3, depth, depth)
		c.ShadowRounded("#00000038", int(xStart+398), 337, 692, 153, 8, size, 3, depth, depth)
		c.ShadowRounded("#00000038", int(xStart), 509, 1090.0, 131, 8, 6, 6, depth, depth)
		c.ShadowRounded("#00000038", int(xStart), 660, 1090.0, 195, 8, 6, 6, depth, depth)
		for i := 0; i < len(recent[math2.Min(1, len(recent)):]); i++ {
			c.ShadowRounded("#00000038", int(xStart), 877+i*(69+20), 1090.0, 69, 8, 6, 6, depth, depth)
		}
	}
	// return c.GenPNG()

	// profile
	// c.SetRGBA255(0, 247, 247, 255)
	// c.DrawRoundedRectangle(44, 34, 663, 1180, 10)
	// c.Clip()
	// {
	// 	i, err := netImage("https://pic.re/images?min_size=600max_size=2000")
	// 	if err == nil {
	// 		i = resize.Resize(0, 1180, i, resize.Lanczos3)
	// 		c.DrawImageAnchored(i, 44+663/2, 34+1180/2, 0.5, 0.5)
	// 	} else {
	// 		return nil, err
	// 		c.Fill()
	// 	}
	// }
	// c.ResetClip()
	// c.InnerShadowRounded("#00000088", 44, 34, 663, 1180, 10, 7)
	// profile inner shadow

	// 头像
	c.SetRGBA255(0, 247, 247, 255)
	c.DrawRoundedRectangle(xStart, 34, 96, 96, 10)
	c.Clip()
	{
		i, err := netImage(user.AvatarURL)
		if err == nil {
			i = resize.Resize(96, 96, i, resize.Lanczos3)
			c.DrawImage(i, int(xStart), 34)
		} else {
			c.Fill()
		}
	}
	c.ResetClip()
	// 昵称
	_ = c.SetFont(semibold, 59)
	c.SetColorAuto("#4D4D4D")
	c.DrawString(user.Username, xStart+120, 105)
	// osu!mania
	c.SetColorAuto("#F3F5FB")
	c.DrawRoundedRectangle(xStart, 143, 385, 47, 8)
	c.Fill()
	c.SetColorAuto("#373A66")
	_ = c.SetFont(semibold, 29)
	c.DrawStringAnchored(mode, xStart+385/2, 143+47/2-5, 0.5, 0.5)
	// rank
	_ = c.SetFont(semibold, 59)
	c.SetColorAuto("#F3F5FB")
	c.DrawRoundedRectangle4(xStart, 200, 385, 67, 8, 8, 0, 0)
	c.Fill()
	// rank text
	_ = c.SetFont(regular, 40)
	c.SetColorAuto("#373A66")
	c.DrawStringAnchored("rank", xStart+385/2-16, 200+28, 1, 0.5)
	_ = c.SetFont(regular, 29)
	c.SetColorAuto("#4D5AA2")
	c.DrawRoundedRectangle(xStart+385/2-3, 200+24, 6, 25, 3)
	c.Fill()
	c.SetColorAuto("#7A84C9")
	c.DrawStringAnchored(strconv.Itoa(user.Statistics.CountryRank), xStart+385/2+16, 200+30, 0, 0.5)
	c.SetColorAuto("#E4E9F8")
	c.DrawRoundedRectangle4(xStart, 267, 385, 223, 0, 0, 8, 8)
	c.Fill()
	// pp
	c.SetColorAuto("#FFF5F5")
	c.DrawRoundedRectangle4(xStart+398, 143, 692, 54, 8, 8, 0, 0)
	c.Fill()
	_ = c.SetFont(regular, 40)
	c.SetColorAuto("#D589A1")
	c.DrawStringAnchored("pp", xStart+398+692/2-16, 143+15, 1, 0.5)
	_ = c.SetFont(regular, 29)
	c.SetColorAuto("#CC777F")
	c.DrawRoundedRectangle(xStart+398+692/2-3, 143+13, 6, 25, 3)
	c.Fill()
	c.SetColorAuto("#B17A82")
	c.DrawStringAnchored(fmt.Sprintf("%.1f", user.Statistics.Pp), xStart+398+692/2+16, 143+20, 0, 0.5)
	c.SetColorAuto("#EDDCD7")
	c.DrawRoundedRectangle4(xStart+398, 197, 692, 33, 0, 0, 8, 8)
	c.FillPreserve()
	c.Clip()
	c.SetColorAuto("#EDC0BE")
	r := utils.Remap(0.17, 0.0, 1.0, 33.0, 8.0)
	c.DrawRoundedRectangle4(xStart+398, 197, 692*0.17, 33, 0, 0, 8, r)
	c.Fill()
	c.ResetClip()
	// acc
	c.SetColorAuto("#DAEEDE")
	c.DrawRoundedRectangle4(xStart+398, 240, 692, 54, 8, 8, 0, 0)
	c.Fill()
	_ = c.SetFont(regular, 40)
	c.SetColorAuto("#2BB33A")
	c.DrawStringAnchored("acc", xStart+398+692/2-16, 240+15, 1, 0.5)
	_ = c.SetFont(regular, 29)
	c.SetColorAuto("#2BB33A")
	c.DrawRoundedRectangle(xStart+398+692/2-3, 240+13, 6, 25, 3)
	c.Fill()
	c.SetColorAuto("#2BB33A")
	c.DrawStringAnchored(fmt.Sprintf("%.1f%%", user.Statistics.HitAccuracy), xStart+398+692/2+16, 240+20, 0, 0.5)
	c.SetColorAuto("#CAE8CC")
	c.DrawRoundedRectangle4(xStart+398, 294, 692, 33, 0, 0, 8, 8)
	c.FillPreserve()
	c.Clip()
	c.SetColorAuto("#AFD7B4")
	r = utils.Remap(user.Statistics.HitAccuracy/100, 0, 1, 33, 8)
	c.DrawRoundedRectangle4(xStart+398, 294, 692*user.Statistics.HitAccuracy/100, 33, 0, 0, 8, r)
	c.Fill()
	c.ResetClip()
	_ = c.SetFont(regular, 23)
	c.SetColorAuto("#1D7B28")
	// c.DrawStringAnchored("300", xStart+398+692-20, 294+12, 1, 0.5)

	// grade statistics
	c.SetColorAuto("#F2F4FA")
	c.DrawRectangle(xStart+398, 337, 692, 73)
	c.Fill()
	c.SetColorAuto("#E7EAFA")
	c.DrawRoundedRectangle4(xStart+398, 410, 692, 80, 0, 0, 8, 8)
	c.Fill()
	_ = c.SetFont(regular, 40)
	c.SetColorAuto("#587EBC")
	c.DrawStringAnchored("grade statistics", xStart+398+692/2, 337+73/2, 0.5, 0.2)
	//
	width := 80.0
	_ = c.SetFont(regular, 26)
	names := []string{
		"SS",
		"SS",
		"S",
		"S",
		"A",
	}
	colors := []string{
		"#BD2E91",
		"#B73295",
		"#4DA8B9",
		"#4EABB7",
		"#97D446",
	}
	fontColors := []string{
		"#FCF5FE",
		"#F6DE51",
		"#FCF5FE",
		"#F6DE51",
		"#345C23",
	}
	for i, x := range images.Grid(xStart+398, 692, 5, width, 30) {
		c.SetColorAuto(colors[i])
		c.DrawRoundedRectangle(x, 410+10, width, 40, 20)
		c.Fill()
		_ = c.SetFont(semibold, 30)
		c.SetColor(brightness(hexcolor(fontColors[i]), 0.3))
		c.DrawStringAnchored(names[i], x+width/2+1, 410+10+15+2, 0.5, 0.5)
		c.SetColorAuto(fontColors[i])
		c.DrawStringAnchored(names[i], x+width/2, 410+10+15, 0.5, 0.5)
		c.SetColorAuto("#7D8083")
		_ = c.SetFont(regular, 20)
		switch i {
		case 0:
			c.DrawStringAnchored(strconv.Itoa(user.Statistics.GradeCounts.SSH), x+width/2, 410+40+8, 0.5, 1)
		case 1:
			c.DrawStringAnchored(strconv.Itoa(user.Statistics.GradeCounts.Ss), x+width/2, 410+40+8, 0.5, 1)
		case 2:
			c.DrawStringAnchored(strconv.Itoa(user.Statistics.GradeCounts.Sh), x+width/2, 410+40+8, 0.5, 1)
		case 3:
			c.DrawStringAnchored(strconv.Itoa(user.Statistics.GradeCounts.S), x+width/2, 410+40+8, 0.5, 1)
		case 4:
			c.DrawStringAnchored(strconv.Itoa(user.Statistics.GradeCounts.A), x+width/2, 410+40+8, 0.5, 1)

		}

	}

	// play time/total hists/play count/ranked score
	c.SetColorAuto("#E3EAF6")
	c.DrawRoundedRectangle(xStart, 509, 1090.0, 131, 8)
	c.Fill()

	// best performance
	c.SetColorAuto("#F2F4F9")
	c.DrawRoundedRectangle(xStart, 660, 1090.0, 195, 8)
	c.Fill()
	c.SetColorAuto("#E4E9F4")
	c.DrawRoundedRectangle4(xStart, 720, 1090.0, 135, 0, 0, 8, 8)
	c.Fill()
	_ = c.SetFont(regular, 36)
	c.SetColorAuto("#7D8083")
	c.DrawStringAnchored("best performance", xStart+13, 660+10, 0, 1)
	w, _ := c.MeasureString("best performance")
	c.DrawRoundedRectangle(xStart+25+w, 660+23, 5, 27, 2.5)
	c.Fill()
	{
		var bb Score
		if len(best) > 0 {
			bb = best[0]
		} else {
			bb = recent[0]
		}
		// c.ShadowRounded("#00000033", 752, 744, 168.0, 93, 4, 2, -2, 2, 2)
		x := xStart + 20
		c.DrawRoundedRectangle(x, 744, 168.0, 93, 4)
		c.Clip()
		i, err := netImage(bb.Beatmapset.Covers.Card)
		if err == nil {
			i = resize.Resize(0, 93, i, resize.Lanczos3)
			c.DrawImage(i, int(x), 744)
		} else {
			c.Fill()
		}
		c.ResetClip()
		c.InnerShadowRounded("#00000033", int(x), 744, 168.0, 93, 4, 4)
		// title
		_ = c.SetFont(regular, 43)
		c.DrawStringAnchored(utils.StringLimit(bb.Beatmapset.Title, 30), x+168+13, 744+13, 0, 0.5)
		// artist
		_ = c.SetFont(regular, 20)
		{
			offset := 0.0 // 1+5
			infos := map[string]string{
				"Artist": bb.Beatmapset.Artist,
				"Mapper": strconv.Itoa(bb.Beatmap.UserID),
				"BID":    strconv.Itoa(bb.Beatmapset.ID),
				"Stars":  fmt.Sprintf("%.2f*", bb.Beatmap.DifficultyRating),
				"Acc":    fmt.Sprintf("%.2f%%", bb.Accuracy*100),
				"Rank":   bb.Rank,
			}
			for k, v := range infos {
				x2 := xStart + 205
				c.DrawRectangle(x2+offset, 793, 1, 34)
				c.Fill()
				c.DrawStringAnchored(v, offset+x2+5, 798, 0, 0.5)
				w1, _ := c.MeasureString(v)
				_ = c.SetFont(regular, 16)
				c.DrawStringAnchored(k, offset+x2+5, 818, 0, 0.5)
				w2, _ := c.MeasureString(k)
				offset += 6 + math.Max(w1, w2)
				offset += 30
			}
		}
		_ = c.SetFont(regular, 19)
		c.DrawStringAnchored("pp", xStart+1090-128, 770, 0, 0.5)
		_ = c.SetFont(regular, 42)
		c.SetColorAuto("#29304E")
		c.DrawStringAnchored(fmt.Sprintf("%.1f", bb.Pp), xStart+1090-128, 798, 0, 0.5)
	}

	// recent
	for i, b := range recent[math2.Min(1, len(recent)):] {
		c.SetColorAuto("#E4E9F4")
		c.DrawRoundedRectangle(xStart, 877+float64(i)*(69+20), 1090, 69, 8)
		c.Fill()

		{
			x := xStart + 15
			y := 877 + (i)*(69+20) + 15
			c.DrawRoundedRectangle(x, float64(y), 39, 39, 4)
			c.Clip()
			img, err := netImage(b.Beatmapset.Covers.List)
			if err == nil {
				img = resize.Resize(0, 39, img, resize.Lanczos3)
				c.DrawImage(img, int(x), y)
			} else {
				panic(err)
				c.Fill()
			}
			c.ResetClip()
			c.InnerShadowRounded("#00000033", int(x), y, 39, 39, 4, 2)
		}
		_ = c.SetFont(regular, 24)
		c.SetColorAuto("#7D8083")
		c.DrawStringAnchored(b.Beatmapset.Title, xStart+68, 877+float64(i)*(69+20)+17, 0, 0.5)
		_ = c.SetFont(regular, 19.7)
		detail := fmt.Sprintf("[4K] V2 | %d | %.2f* | ", b.Beatmap.ID, b.Beatmap.DifficultyRating)
		c.DrawStringAnchored(detail, xStart+68, 877+float64(i)*(69+20)+45, 0, 0.5)
		w, _ := c.MeasureString(detail)
		detail = fmt.Sprintf("%.1f%%", b.Accuracy*100)
		{ // text shadow
			c.SetColorAuto("#83740F")
			c.DrawStringAnchored(detail, xStart+68+w+1, 877+float64(i)*(69+20)+45, 0, 0.5)
		}
		c.SetColorAuto("#E6CD19")
		c.DrawStringAnchored(detail, xStart+68+w, 877+float64(i)*(69+20)+45, 0, 0.5)

		w2, _ := c.MeasureString(detail)
		c.SetColorAuto("#7D8083")
		c.DrawStringAnchored(" | ", xStart+68+w+w2, 877+float64(i)*(69+20)+45, 0, 0.5)
		w3, _ := c.MeasureString(" | ")
		c.DrawStringAnchored(b.Rank, xStart+68+w+w2+w3, 877+float64(i)*(69+20)+45, 0, 0.5)

		_ = c.SetFont(regular, 33.5)
		c.SetHexColor("#EC83B4")
		c.DrawStringAnchored(fmt.Sprintf("%.1fpp", b.Pp), xStart+1090-28, 877+float64(i)*(69+20)+69/2, 1, 0.33)
	}
	// level bar
	c.SetRGBA255(217, 217, 217, 255)
	c.DrawRoundedRectangle(xStart+1140-2.5, 34, 5, 1110, 2.5)
	c.Fill()
	c.SetColorAuto("#EC83B4")
	c.DrawRoundedRectangle(xStart+1140-2.5, 34+(1.0-float64(user.Statistics.Level.Progress)/100.0)*1110, 5, 1110*float64(user.Statistics.Level.Progress)/100.0, 2.5)
	c.Fill()

	// level
	c.SetColorAuto("#F3E7EC")
	c.DrawCircle(xStart+1140, 1186, 24)
	c.SetLineWidth(2)
	c.Stroke()
	_ = c.SetFont(regular, 22)
	c.SetColorAuto("#7D8083")
	c.DrawStringAnchored(fmt.Sprintf("%d", user.Statistics.Level.Current), xStart+1140, 1183, 0.5, 0.5)
	c.UseDefaultFont(22)
	c.DrawString(fmt.Sprintf("Paimeng-Bot %s", time.Now().Format("2006年01月02日 15:04:05")), 20, float64(c.Height()-20))
	// c.DrawImage(bg, 0, 0)
	return c, nil
}

func hexcolor(s string) color.Color {
	s = strings.TrimPrefix(s, "#")
	var r, g, b uint8
	format := "%02x%02x%02x"
	_, _ = fmt.Sscanf(s, format, &r, &g, &b)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func brightness(s color.Color, f float64) color.Color {
	return color.RGBA{
		R: uint8(float64(color.RGBAModel.Convert(s).(color.RGBA).R) * f),
		G: uint8(float64(color.RGBAModel.Convert(s).(color.RGBA).G) * f),
		B: uint8(float64(color.RGBAModel.Convert(s).(color.RGBA).B) * f),
		A: 255,
	}
}
