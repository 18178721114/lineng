package handler

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/mozillazg/go-pinyin"

	"lineng/internal/model"
)

func trimStr(s string) string { return strings.TrimSpace(s) }

func parseUint(s string) uint {
	var v uint
	fmt.Sscanf(s, "%d", &v)
	return v
}

func anyToInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}

func anyToStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func anyF(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}

func round2(f float64) float64 { return math.Round(f*100) / 100 }

// toPinyin 中文转全拼 + 首字母；非汉字字母数字原样保留小写
func toPinyin(name string) (full, initial string) {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	for _, r := range name {
		s := string(r)
		py := pinyin.Pinyin(s, a)
		if len(py) > 0 && len(py[0]) > 0 && py[0][0] != "" && py[0][0] != s {
			full += py[0][0]
			initial += py[0][0][:1]
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			l := unicode.ToLower(r)
			full += string(l)
			initial += string(l)
		}
	}
	return full, initial
}

func dateRange(start, end string) []string {
	var out []string
	sd, e1 := time.Parse("2006-01-02", start)
	ed, e2 := time.Parse("2006-01-02", end)
	if e1 != nil || e2 != nil || sd.After(ed) {
		return out
	}
	for d := sd; !d.After(ed); d = d.AddDate(0, 0, 1) {
		out = append(out, d.Format("2006-01-02"))
	}
	return out
}

func md(date string) string { // 2026-10-08 -> 10.8
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return date
	}
	return parts[1] + "." + strings.TrimPrefix(parts[2], "0")
}

// ---------------- 时间（TIME 列）转换 ----------------

// hhmm 把数据库返回的 HH:MM:SS 规整为 HH:MM，空值原样返回。
func hhmm(t string) string {
	t = strings.TrimSpace(t)
	if len(t) > 5 {
		t = t[:5]
	}
	return t
}

// workTime 规整工作时间为 HH:MM，未填时落到 00:00 以满足 NOT NULL。
func workTime(v any) model.Clock {
	t := hhmm(anyToStr(v))
	if t == "" {
		return model.Clock("00:00")
	}
	return model.Clock(t)
}

// optTime 把加班时间转换为 *model.Clock，未填时存 NULL。
func optTime(v any) *model.Clock {
	t := hhmm(anyToStr(v))
	if t == "" {
		return nil
	}
	c := model.Clock(t)
	return &c
}

// derefTime 把数据库可空时间转换为字符串，NULL 输出空串。
func derefTime(p *model.Clock) string {
	if p == nil {
		return ""
	}
	return hhmm(string(*p))
}

func cellCol(n int) string { // 1->A
	s := ""
	for n > 0 {
		n--
		s = string(rune('A'+n%26)) + s
		n /= 26
	}
	return s
}
