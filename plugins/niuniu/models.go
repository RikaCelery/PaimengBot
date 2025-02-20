// Package niu 牛牛大作战
package niuniu

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
)

var (
	daJiaoProps = []string{"伟哥", "媚药"}
	jjPorps     = []string{"击剑神器", "击剑神稽"}
	query       = "WHERE UID = ?"
)

type lastLength struct {
	TimeLimit time.Time
	Count     int
	Length    float64
}

type shopItem struct {
	Name        string `json:"name"`
	Cost        int    `json:"cost"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}
type users []*userInfo

type model struct {
	sql sql.Sqlite
	sync.RWMutex
}

type userInfo struct {
	UID       int64
	Length    float64
	UserCount int
	WeiGe     int // 伟哥
	Philter   int // 媚药
	Artifact  int // 击剑神器
	ShenJi    int // 击剑神稽
	Buff1     int // 暂定
	Buff2     int // 暂定
	Buff3     int // 暂定
	Buff4     int // 暂定
	Buff5     int // 暂定
}

// AuctionInfo 拍卖信息
type AuctionInfo struct {
	ID     int     `db:"id"`
	UserID int64   `db:"user_id"`
	Length float64 `db:"length"`
	Money  float64 `db:"money"`
}

// BaseInfo ...
type BaseInfo struct {
	UID    int64
	Length float64
}

// BaseInfos ...
type BaseInfos []BaseInfo

func (m users) filter(pos bool) users {
	if pos {
		return m.positive()
	}
	return m.negative()
}

func (m users) positive() users {
	var m1 []*userInfo
	for _, i2 := range m {
		if i2.Length > 0 {
			m1 = append(m1, i2)
		}
	}
	return m1
}

func (m users) negative() users {
	var m1 []*userInfo
	for _, i2 := range m {
		if i2.Length <= 0 {
			m1 = append(m1, i2)
		}
	}
	return m1
}

func (m users) sort(isDesc bool) {
	t := func(i, j int) bool {
		return m[i].Length < m[j].Length
	}
	if isDesc {
		t = func(i, j int) bool {
			return m[i].Length > m[j].Length
		}
	}
	sort.Slice(m, t)
}

func (m users) ranking(niuniu float64, uid int64) int {
	m.sort(niuniu > 0)
	for i, user := range m {
		if user.UID == uid {
			return i + 1
		}
	}
	return -1
}

func (u *userInfo) useWeiGe() (string, float64) {
	niuniu := u.Length
	reduce := math.Abs(hitGlue(niuniu))
	niuniu += reduce
	return randomChoice([]string{
		fmt.Sprintf("哈哈，你这一用道具，牛牛就像是被激发了潜能，增加了%.2fcm！看来今天是个大日子呢！", reduce),
		fmt.Sprintf("你这是用了什么神奇的道具？牛牛竟然增加了%.2fcm，简直是牛气冲天！", reduce),
		fmt.Sprintf("使用道具后，你的牛牛就像是开启了加速模式，一下增加了%.2fcm，这成长速度让人惊叹！", reduce),
	}), niuniu
}

func (u *userInfo) usePhilter() (string, float64) {
	niuniu := u.Length
	reduce := math.Abs(hitGlue(niuniu))
	niuniu -= reduce
	return randomChoice([]string{
		fmt.Sprintf("你使用媚药,咿呀咿呀一下使当前长度发生了一些变化，当前长度%.2f", niuniu),
		fmt.Sprintf("看来你追求的是‘微观之美’，故意使用道具让牛牛凹进去了%.2fcm！", reduce),
		fmt.Sprintf("缩小奇迹’在你身上发生了，牛牛凹进去了%.2fcm，你的选择真是独特！", reduce),
	}), niuniu
}

func (u *userInfo) useArtifact(adduserniuniu float64) (string, float64, float64) {
	myLength := u.Length
	difference := myLength - adduserniuniu
	var (
		change float64
	)
	if difference > 0 {
		change = hitGlue(myLength + adduserniuniu)
	} else {
		change = hitGlue((myLength + adduserniuniu) / 2)
	}
	myLength += change
	return randomChoice([]string{
		fmt.Sprintf("凭借神秘道具的力量，你让对方在你的长度面前俯首称臣！你的长度增加了%.2fcm，当前长度达到了%.2fcm", change, myLength),
		fmt.Sprintf("神器在手，天下我有！你使用道具后，长度猛增%.2fcm，现在的总长度是%.2fcm，无人能敌！", change, myLength),
		fmt.Sprintf("这就是道具的魔力！你轻松增加了%.2fcm，让对手望尘莫及，当前长度为%.2fcm！", change, myLength),
		fmt.Sprintf("道具一出，谁与争锋！你的长度因道具而增长%.2fcm，现在的长度是%.2fcm，霸气尽显！", change, myLength),
		fmt.Sprintf("使用道具的你，如同获得神助！你的长度增长了%.2fcm，达到%.2fcm的惊人长度，胜利自然到手！", change, myLength),
	}), myLength, adduserniuniu - change/1.3
}

func (u *userInfo) useShenJi(adduserniuniu float64) (string, float64, float64) {
	myLength := u.Length
	difference := myLength - adduserniuniu
	var (
		change float64
	)
	if difference > 0 {
		change = hitGlue(myLength + adduserniuniu)
	} else {
		change = hitGlue((myLength + adduserniuniu) / 2)
	}
	myLength -= change
	var r string
	if myLength > 0 {
		r = randomChoice([]string{
			fmt.Sprintf("哦吼！？看来你的牛牛因为使用了神秘道具而缩水了呢🤣🤣🤣！缩小了%.2fcm！", change),
			fmt.Sprintf("哈哈，看来这个道具有点儿调皮，让你的长度缩水了%.2fcm！现在你的长度是%.2fcm，下次可得小心使用哦！", change, myLength),
			fmt.Sprintf("使用道具后，你的牛牛似乎有点儿害羞，缩水了%.2fcm！现在的长度是%.2fcm，希望下次它能挺直腰板！", change, myLength),
			fmt.Sprintf("哎呀，这个道具的效果有点儿意外，你的长度减少了%.2fcm，现在只有%.2fcm了！下次选道具可得睁大眼睛！", change, myLength),
		})
	} else {
		r = randomChoice([]string{
			fmt.Sprintf("哦哟，小姐姐真是玩得一手好游戏，使用道具后数值又降低了%.2fcm，小巧得更显魅力！", change),
			fmt.Sprintf("看来小姐姐喜欢更加精致的风格，使用道具后，数值减少了%.2fcm，更加迷人了！", change),
			fmt.Sprintf("小姐姐的每一次变化都让人惊喜，使用道具后，数值减少了%.2fcm，更加优雅动人！", change),
			fmt.Sprintf("小姐姐这是在展示什么是真正的精致小巧，使用道具后，数值减少了%.2fcm，美得不可方物！", change),
		})
	}
	return r, myLength, adduserniuniu + 0.7*change
}

func (u *userInfo) applyProp(props string) error {
	propsMap := map[string]struct {
		itemCount *int
		errMsg    string
	}{
		"伟哥":   {&u.WeiGe, "你还没有伟哥呢,不能使用"},
		"媚药":   {&u.Philter, "你还没有媚药呢,不能使用"},
		"击剑神器": {&u.Artifact, "你还没有击剑神器呢,不能使用"},
		"击剑神稽": {&u.ShenJi, "你还没有击剑神稽呢,不能使用"},
	}

	if propInfo, ok := propsMap[props]; ok {
		return u.useItem(propInfo.itemCount, propInfo.errMsg)
	}
	return ErrPropNotFound
}

func (u *userInfo) useItem(itemCount *int, errMsg string) error {
	if *itemCount > 0 {
		*itemCount--
		return nil
	}
	return errors.New(errMsg)
}

func (u *userInfo) checkProps(props, propSort string) error {
	validProps := map[string][]string{
		"dajiao": daJiaoProps,
		"jj":     jjPorps,
	}

	// 检查是否是有效道具类别
	validPropsList, ok := validProps[propSort]
	if !ok {
		return ErrInvalidPropType
	}

	validPropsMap := make(map[string]struct{})
	for _, prop := range validPropsList {
		validPropsMap[prop] = struct{}{}
	}

	// 如果道具属于有效道具，返回 nil
	if _, exists := validPropsMap[props]; exists {
		return nil
	}

	// 检查是否相反
	conflictingProps := daJiaoProps
	if propSort == "dajiao" {
		conflictingProps = jjPorps
	}

	// 如果道具属于冲突集合,返回
	for _, conflictProp := range conflictingProps {
		if props == conflictProp {
			return ErrInvalidPropUsageScope
		}
	}

	return ErrPropNotFound
}

func (u *userInfo) purchaseItem(n int) (int, error) {
	var (
		money int
		err   error
	)
	switch n {
	case 1:
		u.WeiGe += 5
	case 2:
		u.Philter += 5
	case 3:
		u.Artifact += 2
	case 4:
		u.ShenJi += 2
	default:
		err = errors.New("无效的选择")
	}
	return money, err
}

func (u *userInfo) processDaJiao(props string) (string, error) {
	var (
		messages string
		info     userInfo
		err      error
		f        float64
	)
	info = *u
	if props != "" {
		err := u.checkProps(props, "dajiao")
		if err != nil {
			return "", err
		}
		if err := u.applyProp(props); err != nil {
			return "", err
		}
	}
	switch {
	case u.WeiGe-info.WeiGe != 0:
		messages, f = u.useWeiGe()
		u.Length = f

	case u.Philter-info.Philter != 0:
		messages, f = u.usePhilter()
		u.Length = f

	default:
		messages, f = hitGlueNiuNiu(u.Length)
		u.Length = f
	}
	return messages, err
}

func (u *userInfo) processJJ(adduserniuniu *userInfo, props string) (string, error) {
	var (
		fencingResult string
		f             float64
		f1            float64
		info          userInfo
		err           error
	)
	info = *u
	if props != "" {
		err := u.checkProps(props, "jj")
		if err != nil {
			return "", err
		}
		if err := u.applyProp(props); err != nil {
			return "", err
		}
	}
	switch {
	case u.ShenJi-info.ShenJi != 0:
		fencingResult, f, f1 = u.useShenJi(adduserniuniu.Length)
		u.Length = f
		adduserniuniu.Length = f1

	case u.Artifact-info.Artifact != 0:
		fencingResult, f, f1 = u.useArtifact(adduserniuniu.Length)
		u.Length = f
		adduserniuniu.Length = f1

	default:
		fencingResult, f, f1 = fencing(u.Length, adduserniuniu.Length)
		u.Length = f
		adduserniuniu.Length = f1
	}
	return fencingResult, err
}

func (db *model) newLength() float64 {
	return float64(rand.Intn(9)+1) + (float64(rand.Intn(100)) / 100)
}

func (db *model) getWordNiuNiu(gid, uid int64) (*userInfo, error) {
	db.RLock()
	defer db.RUnlock()

	var u userInfo
	err := db.sql.Find(strconv.FormatInt(gid, 10), &u, query, uid)
	return &u, err
}

func (db *model) setWordNiuNiu(gid int64, u *userInfo) error {
	db.Lock()
	defer db.Unlock()
	fmt.Println(u)
	err := db.sql.Insert(strconv.FormatInt(gid, 10), u)
	if err != nil {
		err = db.sql.Create(strconv.FormatInt(gid, 10), &userInfo{})
		if err != nil {
			return err
		}
		err = db.sql.Insert(strconv.FormatInt(gid, 10), u)
	}
	return err
}

func (db *model) deleteWordNiuNiu(gid, uid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Del(strconv.FormatInt(gid, 10), query, uid)
}

func (db *model) getAllNiuNiuOfGroup(gid int64) (users, error) {
	db.Lock()
	defer db.Unlock()
	var user userInfo
	var useras users
	err := db.sql.FindFor(fmt.Sprintf("%d", gid), &user, "",
		func() error {
			newUser := user
			useras = append(useras, &newUser)
			return nil
		})
	return useras, err
}

func (db *model) setNiuNiuAuction(gid int64, u *AuctionInfo) error {
	db.Lock()
	defer db.Unlock()
	err := db.sql.Insert(fmt.Sprintf("auction_%d", gid), u)
	if err != nil {
		err = db.sql.Create(fmt.Sprintf("auction_%d", gid), &AuctionInfo{})
		if err != nil {
			return err
		}
		err = db.sql.Insert(fmt.Sprintf("auction_%d", gid), u)
	}
	return err
}

func (db *model) deleteNiuNiuAuction(gid int64, id uint) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Del(fmt.Sprintf("auction_%d", gid), "WHERE id = ?", id)
}

func (db *model) getAllNiuNiuAuction(gid int64) ([]AuctionInfo, error) {
	db.RLock()
	defer db.RUnlock()
	var user AuctionInfo
	var useras []AuctionInfo
	err := db.sql.FindFor(fmt.Sprintf("auction_%d", gid), &user, "",
		func() error {
			useras = append(useras, user)
			return nil
		})

	return useras, err
}
