package HiOSU

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func username2uid(u string) (int64, error) {
	c := client.NewHttpClient(nil)
	i, err := strconv.ParseInt(u, 10, 64)
	if err == nil {
		return i, nil
	}
	token := refreshToken(proxy.GetConfigInt64("appid"), proxy.GetConfigString("secret"))
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://osu.ppy.sh/api/v2/users/@%s", u), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	json, err := c.Do(req)
	if err != nil {
		log.Errorf("username2uid err: %v", err)
		return 0, err
	}
	all, err := io.ReadAll(json.Body)
	if err != nil {
		log.Errorf("username2uid err: %v", err)
		return 0, err
	}
	defer json.Body.Close()
	g := gjson.ParseBytes(all)
	if !g.Get("id").Exists() {
		return 0, fmt.Errorf("未找到该用户")
	}
	return g.Get("id").Int(), nil
}
func BindOSUidHandler(ctx *zero.Ctx) {
	OSUid := strings.TrimSpace(utils.GetArgs(ctx))
	uid, err := username2uid(OSUid)
	if err != nil {
		ctx.Send(err.Error())
		return
	}
	if err := PutOsuID(ctx.Event.UserID, uid); err != nil { // 有报错返回写入日志
		log.Errorf("PutUserUid err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send("绑定成功ˋ( ° ▽、° ) ")
}

func ReferOSUidHandler(ctx *zero.Ctx) { // 查询绑定的OSU Id
	ID := GetOsuid(ctx.Event.UserID) // 查询数据表中用户绑定信息
	if ID == "" {
		ctx.Send("未绑定任何OSU账号")
	} else {
		ctx.Send("当前绑定的账号为:" + ID)
	}
}

func GetOsuid(id int64) (u string) {
	key := fmt.Sprintf("hiosu.UserId.U%v", id) // 查询的key
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		return
	}
	return string(v)
}

func PutOsuID(id int64, u int64) error { // 将OSU id 和用户的ID写入数据表
	key := fmt.Sprintf("hiosu.UserId.U%v", id)                                        // 将表的键值定位 hiosu.UserId.U + Userid
	return proxy.GetLevelDB().Put([]byte(key), []byte(strconv.FormatInt(u, 10)), nil) // 写入Key和aValue的值
}
