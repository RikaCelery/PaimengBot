package history

import (
	"errors"
	"fmt"
	"time"

	sql "github.com/FloatTech/sqlite"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type SimpleEvent struct {
	GroupID int64
	UserID  int64
	ID      int64
	Time    time.Time
	Msg     message.Message
}

func GroupHistoryAll(id int64) ([]SimpleEvent, error) {
	msgs, err := sql.FindAll[tMSG](db, fmt.Sprintf("g%d", id), "")
	if err != nil {
		return nil, err
	}
	var ret []SimpleEvent
	for _, msg := range msgs {
		ret = append(ret, SimpleEvent{
			GroupID: id,
			UserID:  msg.UserID,
			ID:      msg.ID,
			Time:    time.UnixMilli(msg.Stamp),
			Msg:     message.ParseMessageFromString(msg.Msg),
		})
	}
	return ret, nil
}

func GroupHistoryToday(id int64) ([]SimpleEvent, error) {
	now := time.Now()
	// 设置为当天0点
	zeroTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	stamp := zeroTime.UnixMilli()
	msgs, err := sql.FindAll[tMSG](db, fmt.Sprintf("g%d", id), "WHERE stamp >= ?", stamp)
	if err != nil {
		if errors.Is(err, sql.ErrNullResult) {

			return []SimpleEvent{}, nil
		}
		return nil, err
	}
	var ret []SimpleEvent
	for _, msg := range msgs {
		ret = append(ret, SimpleEvent{
			GroupID: id,
			UserID:  msg.UserID,
			ID:      msg.ID,
			Time:    time.UnixMilli(msg.Stamp),
			Msg:     message.ParseMessageFromString(msg.Msg),
		})
	}
	return ret, nil
}

func GroupHistoryRecent(id int64, d time.Duration) ([]SimpleEvent, error) {
	now := time.Now()
	zeroTime := now.Add(-d)
	stamp := zeroTime.UnixMilli()
	msgs, err := sql.FindAll[tMSG](db, fmt.Sprintf("g%d", id), "WHERE stamp >= ?", stamp)
	if err != nil {
		if errors.Is(err, sql.ErrNullResult) {
			return []SimpleEvent{}, nil
		}
		return nil, err
	}
	var ret []SimpleEvent
	for _, msg := range msgs {
		ret = append(ret, SimpleEvent{
			GroupID: id,
			UserID:  msg.UserID,
			ID:      msg.ID,
			Time:    time.UnixMilli(msg.Stamp),
			Msg:     message.ParseMessageFromString(msg.Msg),
		})
	}
	return ret, nil
}

func GetId(gid int64, id int64) (SimpleEvent, error) {
	obj, err := sql.Find[tMSG](db, fmt.Sprintf("g%d", gid), "WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNullResult) {
			return SimpleEvent{}, nil
		}
		return SimpleEvent{}, err
	}
	ret := SimpleEvent{
		GroupID: gid,
		UserID:  obj.UserID,
		ID:      obj.ID,
		Time:    time.UnixMilli(obj.Stamp),
		Msg:     message.ParseMessageFromString(obj.Msg),
	}
	return ret, err
}
