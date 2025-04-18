// Package history 用来本地记录消息
package history

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"time"

	sql "github.com/FloatTech/sqlite"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/corona10/goimagehash"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	en *zero.Engine
	db *sql.Sqlite
	c  = client.NewHttpClient(&client.HttpOptions{TryTime: 3})
)

type tMSG struct {
	ID     int64
	UserID int64
	Stamp  int64
	Msg    string
}

func init() {
	_db := sql.New("msg.db")
	db = &_db
	db.Open(time.Hour)
	en = zero.New()
	en.OnMessage(zero.OnlyGroup).SetPriority(99999).Handle(func(ctx *zero.Ctx) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorln(err)
			}
		}()
		tablse, err := db.ListTables()
		if err != nil {
			return
		}
		if !utils.StringSliceContain(tablse, "g"+fmt.Sprintf("%d", ctx.Event.GroupID)) {
			db.Create(fmt.Sprintf("g%d", ctx.Event.GroupID), &tMSG{})
		}
		db.Insert(fmt.Sprintf("g%d", ctx.Event.GroupID), &tMSG{
			ctx.Event.MessageID.(int64),
			ctx.Event.UserID,
			time.Now().UnixMilli(),
			preprocess(ctx.Event.Message).String(),
		})
		// (ctx.Event.Message)
	})
}

// [CQ:image,file=98A19CFF533CE7F0C7E69F1F4C5BF69E.jpg,sub_type=1,url=https://multimedia.nt.qq.com.cn/download?appid=1407&amp;fileid=EhTCN-KFZ_BMyE0N1WcsbMSbyPEl1hjMvR4g_woo6eKw17bUiwMyBHByb2RQgL2jAVoQRLsfiYCKWB8Tl63kbuY_hA&amp;rkey=CAMSKLgthq-6lGU_nzlDXwRhktUP2keRqms4bxIBej07lemyr2umJ1MHSCc,file_size=499404,summary=&#91;动画表情&#93;]
// [CQ:video,url=https://multimedia.nt.qq.com.cn:443/download?appid=1415&amp;format=origin&amp;orgfmt=t264&amp;spec=0&amp;client_proto=ntv2&amp;client_appid=537243441&amp;client_type=win&amp;client_ver=9.9.15-27597&amp;client_down_type=auto&amp;client_aio_type=aio&amp;rkey=CAESkAFOP3AEg5tbdRC3ZWwdGBRSS4b7HLhgpU_7f5T1Vsf03Cur5ndkbL7EhZ_1-xjG_QreFkXjIjlIz2X3qEIP5AKdmyhMfSOLAASW5K0mloyUi8BLsIZf2Sl7VA9X1-els41mHGSSATkrMtB6I4fpu7yaxi9P7snYbmdQ9TIFcNL6lidMGGmLQPh7BM36iRhiyX8,file_size=5733432,file=962231fb3fd99818cbec4968b66aec4c.mp4,path=https://multimedia.nt.qq.com.cn:443/download?appid=1415&amp;format=origin&amp;orgfmt=t264&amp;spec=0&amp;client_proto=ntv2&amp;client_appid=537243441&amp;client_type=win&amp;client_ver=9.9.15-27597&amp;client_down_type=auto&amp;client_aio_type=aio&amp;rkey=CAESkAFOP3AEg5tbdRC3ZWwdGBRSS4b7HLhgpU_7f5T1Vsf03Cur5ndkbL7EhZ_1-xjG_QreFkXjIjlIz2X3qEIP5AKdmyhMfSOLAASW5K0mloyUi8BLsIZf2Sl7VA9X1-els41mHGSSATkrMtB6I4fpu7yaxi9P7snYbmdQ9TIFcNL6lidMGGmLQPh7BM36iRhiyX8]

func storeVideo(url string) (md5 string, err error) {
	if !utils.DirExists("videos") {
		utils.MakeDir("videos")
	}
	f := utils.PathJoin("videos", fmt.Sprintf("%d.tmp", time.Now().UnixMilli()))
	return c.DownloadToFileHash(f, url)
}

func hashImageFromUrl(url string) (phash, hmd5, format string, err error) {
	r, err := c.GetReader(url)
	if err != nil {
		return "", "", "", err
	}
	defer r.Close()
	imgbytes, err := io.ReadAll(r)
	if err != nil {
		return "", "", "", err
	}
	hasher := md5.New()
	_, err = io.Copy(hasher, bytes.NewReader(imgbytes))
	if err != nil {
		return "", "", "", err
	}
	hmd5 = hex.EncodeToString(hasher.Sum(nil))
	img, f, err := image.Decode(bytes.NewReader(imgbytes))
	if err != nil {
		return "", "", "", err
	}
	format = f
	h, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", "", "", err
	}
	phash = fmt.Sprintf("%016x", h.GetHash())
	return
}
func preprocess(msg message.Message) (ret message.Message) {

	for i := range msg {
		switch msg[i].Type {
		case "text":
		case "at":
		case "video":

			url := msg[i].Data["url"]
			file_size := msg[i].Data["file_size"]
			file := msg[i].Data["file"]
			md5, err := storeVideo(url)
			if err != nil {
				logrus.Warnf("<history>gen video hashes error %v", err)
				continue
			}
			for k := range msg[i].Data {
				delete(msg[i].Data, k)
			}

			if file_size != "" {
				msg[i].Data["file_size"] = file_size
			}
			msg[i].Data["file"] = file
			msg[i].Data["md5"] = md5
		case "image":
			url := msg[i].Data["url"]
			phash, md5hash, format, err := hashImageFromUrl(url)
			if err != nil {
				logrus.Warnf("<history>gen image hashes error %v", err)
				continue
			}
			sub_type := msg[i].Data["sub_type"]
			summary := msg[i].Data["summary"]
			file_size := msg[i].Data["file_size"]
			for k := range msg[i].Data {
				delete(msg[i].Data, k)
			}
			if sub_type != "" {
				msg[i].Data["sub_type"] = sub_type
			}
			if summary != "" {
				msg[i].Data["summary"] = summary
			}
			if file_size != "" {
				msg[i].Data["file_size"] = file_size
			}
			// custom fields
			msg[i].Data["md5"] = md5hash
			msg[i].Data["phash"] = phash
			msg[i].Data["format"] = format
		}
	}
	return msg
}
