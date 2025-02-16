package deer_pipe

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var cnNumMap = map[string]int{
	"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
	"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
	"十一": 11, "十二": 12, "十三": 13, "十四": 14, "十五": 15,
	"十六": 16, "十七": 17, "十八": 18, "十九": 19, "二十": 20,
	"二十一": 21, "二十二": 22, "二十三": 23, "二十四": 24, "二十五": 25,
	"二十六": 26, "二十七": 27, "二十八": 28, "二十九": 29, "三十": 30,
	"三十一": 31,
}

func parseDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	now := time.Now()

	// 尝试完整格式：2024年3月5日 或 2024-3-5 或 2024/3/5
	if t := tryFullFormat(s); t != nil {
		return t, nil
	}

	// 尝试中文月日格式：一月三日 或 1月8号
	if t := tryChineseMonthDay(s, now); t != nil {
		return t, nil
	}

	// 尝试纯日格式：三号 或 3日
	if t := tryDayOnly(s, now); t != nil {
		return t, nil
	}

	// 尝试数字简写格式：2024-3-5 或 2024/3/5
	if t := tryDigitShortFormat(s); t != nil {
		return t, nil
	}

	return nil, errors.New("无法识别的日期")
}

func tryFullFormat(s string) *time.Time {
	re := regexp.MustCompile(`^(\d{4})\s*年\s*(\d{1,2}|[一二三四五六七八九十]{1,3})\s*月\s*(\d{1,2}|[一二三四五六七八九十]{1,3})\s*[日号]?$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		year := parseInt(matches[1])
		month := parseCnNumber(matches[2])
		day := parseCnNumber(matches[3])
		return createDate(year, month, day)
	}
	return nil
}

func tryChineseMonthDay(s string, now time.Time) *time.Time {
	re := regexp.MustCompile(`^(\d{1,2}|[一二三四五六七八九十]{1,3})\s*月\s*(\d{1,2}|[一二三四五六七八九十]{1,3})\s*[日号]$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		year := now.Year()
		month := parseCnNumber(matches[1])
		day := parseCnNumber(matches[2])
		return createDate(year, month, day)
	}
	return nil
}

func tryDayOnly(s string, now time.Time) *time.Time {
	re := regexp.MustCompile(`^(\d{1,2}|[一二三四五六七八九十]{1,3})\s*[日号]$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		year := now.Year()
		month := int(now.Month())
		day := parseCnNumber(matches[1])
		return createDate(year, month, day)
	}
	return nil
}

func tryDigitShortFormat(s string) *time.Time {
	re := regexp.MustCompile(`^(\d{4})[-/](\d{1,2})[-/](\d{1,2})$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		year := parseInt(matches[1])
		month := parseInt(matches[2])
		day := parseInt(matches[3])
		return createDate(year, month, day)
	}
	return nil
}

func parseCnNumber(s string) int {
	if num, err := strconv.Atoi(s); err == nil {
		return num
	}
	return cnNumMap[s]
}

func parseInt(s string) int {
	num, _ := strconv.Atoi(s)
	return num
}

func createDate(year, month, day int) *time.Time {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return nil
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	// 验证日期是否有效
	if t.Year() != year || t.Month() != time.Month(month) || t.Day() != day {
		return nil
	}

	return &t
}
