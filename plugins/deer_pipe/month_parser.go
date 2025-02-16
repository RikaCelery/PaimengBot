package deer_pipe

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var monthCnMap = map[string]int{
	"一": 1, "二": 2, "三": 3, "四": 4, "五": 5, "六": 6,
	"七": 7, "八": 8, "九": 9, "十": 10, "十一": 11, "十二": 12,
}

func parseMonth(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	now := time.Now()
	currentYear := now.Year()

	// 清理特殊字符
	s = regexp.MustCompile(`[份\s]`).ReplaceAllString(s, "")
	// 匹配格式：2024年3月 或 2024-03 或 2024/3
	if t := tryYearMonthFormat(s, currentYear); t != nil {
		return t, nil
	}

	// 匹配纯中文月份：三月 或 十二月
	if t := tryChineseMonthOnly(s, currentYear); t != nil {
		return t, nil
	}

	// 匹配数字月份：3月 或 03月
	if t := tryDigitMonthOnly(s, currentYear); t != nil {
		return t, nil
	}

	// 匹配无年份数字格式：3 或 03（需要与纯数字年份区分）
	if t := tryMonthOnly(s, currentYear); t != nil {
		return t, nil
	}

	return nil, errors.New("无法识别的月份格式")
}

func tryYearMonthFormat(s string, defaultYear int) *time.Time {
	re := regexp.MustCompile(`^(\d{4})[年\-/]*(0?[1-9]|1[0-2]|[一二三四五六七八九十]{1,3})[月]?$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		year := parseInt(matches[1])
		month := parseMonthNumber(matches[2])
		return createMonthDate(year, month)
	}
	return nil
}

func tryChineseMonthOnly(s string, year int) *time.Time {
	re := regexp.MustCompile(`^(0?[1-9]|1[0-2]|[一二三四五六七八九十]{1,3})月?$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		month := parseMonthNumber(matches[1])
		return createMonthDate(year, month)
	}
	return nil
}

func tryDigitMonthOnly(s string, year int) *time.Time {
	re := regexp.MustCompile(`^(0?[1-9]|1[0-2])月?$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		month := parseInt(matches[1])
		return createMonthDate(year, month)
	}
	return nil
}

func tryMonthOnly(s string, year int) *time.Time {
	if month, err := strconv.Atoi(s); err == nil {
		if month >= 1 && month <= 12 {
			return createMonthDate(year, month)
		}
	}
	return nil
}

func parseMonthNumber(s string) int {
	// 优先尝试数字转换
	if num, err := strconv.Atoi(s); err == nil {
		return num
	}

	// 处理中文数字
	if strings.HasPrefix(s, "十") {
		if len(s) > 1 { // 十一、十二
			return 10 + monthCnMap[s[1:]]
		}
		return 10 // 单独"十"表示十月
	}
	return monthCnMap[s]
}

func createMonthDate(year, month int) *time.Time {
	if month < 1 || month > 12 {
		return nil
	}
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	return &t
}
