// Package niu 牛牛大作战
package niuniu

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
	"github.com/RicheyJang/PaimengBot/basic/sc"
	"github.com/RicheyJang/PaimengBot/utils"
)

var (
	db         = &model{}
	globalLock sync.Mutex
	// ErrNoBoys 表示当前没有男孩子可用的错误。
	ErrNoBoys = errors.New("暂时没有男孩子哦")

	// ErrNoGirls 表示当前没有女孩子可用的错误。
	ErrNoGirls = errors.New("暂时没有女孩子哦")

	// ErrNoNiuNiu 表示用户尚未拥有牛牛的错误。
	ErrNoNiuNiu = errors.New("你还没有牛牛呢,快去注册吧！")

	// ErrNoNiuNiuINAuction 表示拍卖行当前没有牛牛可用的错误。
	ErrNoNiuNiuINAuction = errors.New("拍卖行还没有牛牛呢")

	// ErrNoMoney 表示用户资金不足的错误。
	ErrNoMoney = errors.New("你的钱不够快去赚钱吧！")

	// ErrAdduserNoNiuNiu 表示对方尚未拥有牛牛，因此无法进行某些操作的错误。
	ErrAdduserNoNiuNiu = errors.New("对方还没有牛牛呢，不能🤺")

	// ErrCannotFight 表示无法进行战斗操作的错误。
	ErrCannotFight = errors.New("你要和谁🤺？你自己吗？")

	// ErrNoNiuNiuTwo 表示用户尚未拥有牛牛，无法执行特定操作的错误。
	ErrNoNiuNiuTwo = errors.New("你还没有牛牛呢，咋的你想凭空造一个啊")

	// ErrAlreadyRegistered 表示用户已经注册过的错误。
	ErrAlreadyRegistered = errors.New("你已经注册过了")

	// ErrInvalidPropType 表示传入的道具类别错误的错误。
	ErrInvalidPropType = errors.New("道具类别传入错误")

	// ErrInvalidPropUsageScope 表示道具使用域错误的错误。
	ErrInvalidPropUsageScope = errors.New("道具使用域错误")

	// ErrPropNotFound 表示找不到指定道具的错误。
	ErrPropNotFound = errors.New("道具不存在")
)

func init() {
	if !utils.DirExists("data/niuniu") {
		err := os.MkdirAll("data/niuniu", 0775)
		if err != nil {
			panic(err)
		}
	}
	db.sql = sql.New("data/niuniu/niuniu.db")
	err := db.sql.Open(time.Hour * 24)
	if err != nil {
		panic(err)
	}
}

// DeleteWordNiuNiu ...
func DeleteWordNiuNiu(gid, uid int64) error {
	globalLock.Lock()
	defer globalLock.Unlock()
	return db.deleteWordNiuNiu(gid, uid)
}

// SetWordNiuNiu length > 0 就增加 , length < 0 就减小
func SetWordNiuNiu(gid, uid int64, length float64) error {
	globalLock.Lock()
	defer globalLock.Unlock()
	niu, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return err
	}
	niu.Length += length
	return db.setWordNiuNiu(gid, niu)
}

// GetWordNiuNiu ...
func GetWordNiuNiu(gid, uid int64) (float64, error) {
	globalLock.Lock()
	defer globalLock.Unlock()

	niu, err := db.getWordNiuNiu(gid, uid)
	return niu.Length, err
}

// GetRankingInfo 获取排行信息
func GetRankingInfo(gid int64, t bool) (BaseInfos, error) {
	globalLock.Lock()
	defer globalLock.Unlock()
	var (
		list users
		err  error
	)

	niuOfGroup, err := db.getAllNiuNiuOfGroup(gid)
	if err != nil {
		return nil, err
	}

	list = niuOfGroup.filter(t)
	list.sort(t)
	if len(list) == 0 {
		if t {
			return nil, ErrNoBoys
		}
		return nil, ErrNoGirls
	}
	f := make(BaseInfos, len(list))
	for i, info := range list {
		f[i] = BaseInfo{
			UID:    info.UID,
			Length: info.Length,
		}
	}
	return f, nil
}

// GetGroupUserRank 获取指定用户在群中的排名
func GetGroupUserRank(gid, uid int64) (int, error) {
	globalLock.Lock()
	defer globalLock.Unlock()
	niu, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return -1, err
	}
	group, err := db.getAllNiuNiuOfGroup(gid)
	if err != nil {
		return -1, err
	}
	return group.ranking(niu.Length, uid), nil
}

// View 查看牛牛
func View(gid, uid int64, name string) (string, error) {
	i, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return "", ErrNoNiuNiu
	}
	niuniu := i.Length
	var result strings.Builder
	sexLong := "长"
	sex := "♂️"
	if niuniu < 0 {
		sexLong = "深"
		sex = "♀️"
	}
	niuniuList, err := db.getAllNiuNiuOfGroup(gid)
	if err != nil {
		return "", err
	}
	result.WriteString(fmt.Sprintf("\n📛%s<%s>的牛牛信息\n⭕性别:%s\n⭕%s度:%.2fcm\n⭕排行:%d\n⭕%s ",
		name, strconv.FormatInt(uid, 10),
		sex, sexLong, niuniu, niuniuList.ranking(niuniu, uid), generateRandomString(niuniu)))
	return result.String(), nil
}

// HitGlue 打胶
func HitGlue(gid, uid int64, prop string) (string, error) {
	niuniu, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return "", ErrNoNiuNiuTwo
	}

	messages, err := niuniu.processDaJiao(prop)
	if err != nil {
		return "", err
	}
	if err = db.setWordNiuNiu(gid, niuniu); err != nil {
		return "", err
	}
	return messages, nil
}

// Register 注册牛牛
func Register(gid, uid int64) (string, error) {
	if _, err := db.getWordNiuNiu(gid, uid); err == nil {
		return "", ErrAlreadyRegistered
	}
	// 获取初始长度
	length := db.newLength()
	u := userInfo{
		UID:    uid,
		Length: length,
	}
	if err := db.setWordNiuNiu(gid, &u); err != nil {
		return "", err
	}
	return fmt.Sprintf("注册成功,你的牛牛现在有%.2fcm", u.Length), nil
}

// JJ ...
func JJ(gid, uid, adduser int64, prop string) (message string, adduserLength float64, err error) {
	myniuniu, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return "", 0, ErrNoNiuNiu
	}

	adduserniuniu, err := db.getWordNiuNiu(gid, adduser)
	if err != nil {
		return "", 0, ErrAdduserNoNiuNiu
	}

	if uid == adduser {
		return "", 0, ErrCannotFight
	}

	message, err = myniuniu.processJJ(adduserniuniu, prop)
	if err != nil {
		return "", 0, err
	}

	if err = db.setWordNiuNiu(gid, myniuniu); err != nil {
		return "", 0, err
	}

	if err = db.setWordNiuNiu(gid, adduserniuniu); err != nil {
		return "", 0, err
	}

	adduserLength = adduserniuniu.Length

	if err = db.setWordNiuNiu(gid, adduserniuniu); err != nil {
		return "", 0, err
	}

	return
}

// Cancel 注销牛牛
func Cancel(gid, uid int64) (string, error) {
	_, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return "", ErrNoNiuNiuTwo
	}
	err = db.deleteWordNiuNiu(gid, uid)
	if err != nil {
		err = errors.New("遇到不可抗力因素，注销失败！")
	}
	return "注销成功,你已经没有牛牛了", err
}

// Redeem 赎牛牛
func Redeem(gid, uid int64, lastLength float64) error {
	money := sc.BaseCoinOf(uid)
	if money < 150 {
		var builder strings.Builder
		walletName := sc.Unit()
		builder.WriteString("赎牛牛需要150")
		builder.WriteString(walletName)
		builder.WriteString("，快去赚钱吧，目前仅有:")
		builder.WriteString(strconv.FormatFloat(money, 'f', 2, 64))
		builder.WriteString("个")
		builder.WriteString(walletName)
		return errors.New(builder.String())
	}

	if _, ok := sc.AddBaseCoin(uid, -150); !ok {
		return errors.New("添加金钱失败")
	}

	niu, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return ErrNoNiuNiu
	}

	niu.Length = lastLength

	return db.setWordNiuNiu(gid, niu)
}

// Store 牛牛商店
func Store(gid, uid int64, money float64, n int) error {
	info, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return err
	}

	fmt.Println(info)
	_, err = info.purchaseItem(n)
	fmt.Println(info)
	if err != nil {
		return err
	}

	if sc.BaseCoinOf(uid) < money {
		return ErrNoMoney
	}

	if _, ok := sc.AddBaseCoin(uid, -money); !ok {
		return ErrNoMoney
	}

	return db.setWordNiuNiu(gid, info)
}

// Sell 出售牛牛
func Sell(gid, uid int64) (string, error) {
	niu, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return "", ErrNoNiuNiu
	}
	money, t, message := profit(niu.Length)
	if !t {
		return "", errors.New(message)
	}

	if err := db.deleteWordNiuNiu(gid, uid); err != nil {
		return "", err
	}

	_, ok := sc.AddBaseCoin(uid, float64(money))
	if !ok {
		return "", errors.New("添加金钱失败")
	}

	infos, _ := db.getAllNiuNiuAuction(gid)

	u := AuctionInfo{
		ID:     len(infos),
		UserID: niu.UID,
		Length: niu.Length,
		Money:  money * 2,
	}

	err = db.setNiuNiuAuction(gid, &u)
	return message, err
}

// ShowAuction 展示牛牛拍卖行
func ShowAuction(gid int64) ([]AuctionInfo, error) {
	globalLock.Lock()
	defer globalLock.Unlock()
	return db.getAllNiuNiuAuction(gid)
}

// Auction 购买牛牛
func Auction(gid, uid int64, index int) (string, error) {
	infos, err := db.getAllNiuNiuAuction(gid)
	if err != nil {
		return "", ErrNoNiuNiuINAuction
	}
	if _, ok := sc.AddBaseCoin(uid, float64(-infos[index].Money)); !ok {
		return "", ErrNoMoney
	}

	niu, err := db.getWordNiuNiu(gid, uid)

	if err != nil {
		niu.UID = uid
	}

	niu.Length = infos[index].Length

	if infos[index].Money >= 30 {
		niu.WeiGe += 2
		niu.Philter += 2
	}

	if err = db.deleteNiuNiuAuction(gid, uint(index)); err != nil {
		return "", err
	}

	if err = db.setWordNiuNiu(gid, niu); err != nil {
		return "", err
	}

	if infos[index].Money >= 30 {
		return fmt.Sprintf("恭喜你购买成功,当前长度为%.2fcm,此次购买将赠送你2个伟哥,2个媚药", niu.Length), nil
	}

	return fmt.Sprintf("恭喜你购买成功,当前长度为%.2fcm", niu.Length), nil
}

// Bag 牛牛背包
func Bag(gid, uid int64) (string, error) {
	niu, err := db.getWordNiuNiu(gid, uid)
	if err != nil {
		return "", ErrNoNiuNiu
	}

	var result strings.Builder
	result.Grow(100)

	result.WriteString("当前牛牛背包如下\n")
	result.WriteString(fmt.Sprintf("伟哥: %v\n", niu.WeiGe))
	result.WriteString(fmt.Sprintf("媚药: %v\n", niu.Philter))
	result.WriteString(fmt.Sprintf("击剑神器: %v\n", niu.Artifact))
	result.WriteString(fmt.Sprintf("击剑神稽: %v\n", niu.ShenJi))

	return result.String(), nil
}
