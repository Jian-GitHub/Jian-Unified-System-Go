package time

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	YearMillis   = 31556736000 // 1 年的毫秒数
	MonthMillis  = 3240000000  // 1 个月的毫秒数
	DayMillis    = 108000000   // 1 天的毫秒数
	HourMillis   = 3600000     // 1 小时的毫秒数
	MinuteMillis = 60000       // 1 分钟的毫秒数
	SecondMillis = 1000        // 1 秒的毫秒数
)

// UTCStart 公历起始时间（东八区 1999-11-06 15:10:00.000）
var UTCStart = time.Date(1999, 11, 6, 15, 10, 0, 0, time.FixedZone("CST", 8*3600))

type Time struct {
	Year          int
	Month         int
	Day           int
	Hour          int
	Minute        int
	Second        int
	Millis        int64
	TimestampNano int64
}

// Now 当前时间
func Now() Time {
	return ConvertToNewCalendar(time.Now())
}

// GetLocation 获取指定时区 - 空字符串返回上海时区 - 错误字符串抛出异常
func GetLocation(timeZone string) (*time.Location, error) {
	if timeZone == "" {
		return time.LoadLocation("Asia/Shanghai") // 默认时区
	}
	return time.LoadLocation(timeZone)
}

// NewJianCalendarDateTime 解析字符串并创建自定义时间
func NewJianCalendarDateTime(timeStr string) (Time, error) {
	parts := strings.Split(timeStr, "-")
	if len(parts) < 3 {
		return Time{}, fmt.Errorf("invalid date format")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return Time{}, err
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return Time{}, err
	}

	dayTime := strings.Split(parts[2], " ")
	if len(dayTime) < 2 {
		return Time{}, fmt.Errorf("invalid date format")
	}
	day, err := strconv.Atoi(dayTime[0])
	if err != nil {
		return Time{}, err
	}

	timeParts := strings.Split(dayTime[1], ":")
	if len(timeParts) != 3 {
		return Time{}, fmt.Errorf("invalid time format")
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return Time{}, err
	}
	minute, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return Time{}, err
	}

	secondParts := strings.Split(timeParts[2], ".")
	second, err := strconv.Atoi(secondParts[0])
	if err != nil {
		return Time{}, err
	}

	millis := int64(0)
	if len(secondParts) == 2 {
		millis, err = strconv.ParseInt(secondParts[1], 10, 64)
		if err != nil {
			return Time{}, err
		}
	}

	timestampNano := time.Now().UnixNano()

	return Time{year, month, day, hour, minute, second, millis, timestampNano}, nil
}

// GetTotalMillis 计算从 0000-01-01 00:00:00.000 到当前时间的总毫秒数
func (j Time) GetTotalMillis() int64 {
	millis := int64(j.Year)*YearMillis +
		int64(j.Month-1)*MonthMillis +
		int64(j.Day-1)*DayMillis +
		int64(j.Hour)*HourMillis +
		int64(j.Minute)*MinuteMillis +
		int64(j.Second)*SecondMillis +
		j.Millis
	return millis
}

// ConvertToUTC 将新历法时间转换为公历时间
func ConvertToUTC(jianTime Time) time.Time {
	//if timezone != nil {
	//	return ConvertToUTCWithTimezone(jianTime, *timezone)
	//}
	totalMillis := jianTime.GetTotalMillis()
	return UTCStart.Add(time.Duration(totalMillis) * time.Millisecond)
}

// ConvertToUTCWithTimezone 允许指定 IANA 时区
func ConvertToUTCWithTimezone(jianTime Time, timezone *time.Location) time.Time {
	totalMillis := jianTime.GetTotalMillis()
	utcTime := UTCStart.Add(time.Duration(totalMillis) * time.Millisecond)

	// 如果时区为 nil 重置为UTC
	//loc, err := time.LoadLocation(timezone)
	if timezone == nil {
		//	fmt.Printf("⚠️ 时区 \"%s\" 无效，使用 UTC 作为默认时区\n", timezone)
		timezone = time.UTC
	}

	return utcTime.In(timezone)
}

// ConvertToNewCalendar 从公历时间转换为新历法时间
func ConvertToNewCalendar(t time.Time) Time {
	totalMillis := t.Sub(UTCStart).Milliseconds()

	years := int(totalMillis / YearMillis)
	totalMillis %= YearMillis

	months := int(totalMillis/MonthMillis) + 1
	totalMillis %= MonthMillis

	days := int(totalMillis/DayMillis) + 1
	totalMillis %= DayMillis

	hours := int(totalMillis / HourMillis)
	totalMillis %= HourMillis

	minutes := int(totalMillis / MinuteMillis)
	totalMillis %= MinuteMillis

	seconds := int(totalMillis / SecondMillis)
	millis := totalMillis % SecondMillis

	timestampNano := time.Now().UnixNano()

	return Time{years, months, days, hours, minutes, seconds, millis, timestampNano}
}

// ToString 格式化输出
func (j Time) ToString() string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d.%03d",
		j.Year, j.Month, j.Day, j.Hour, j.Minute, j.Second, j.Millis)
}

// 测试代码
func main() {
	// 获取当前时间（公历）
	now := time.Now()

	// 公历 -> 新历法
	jianTime := ConvertToNewCalendar(now)
	fmt.Println("当前新历法时间:", jianTime.ToString())

	// 新历法 -> 公历
	asiaTokyo, _ := time.LoadLocation("Asia/Tokyo")
	convertedUTC := ConvertToUTCWithTimezone(jianTime, asiaTokyo)
	fmt.Println("转换回公历时间:", convertedUTC.Format("2006-01-02 15:04:05.000"))

	// 转换为新西兰时间（Pacific/Auckland）
	pacificAuckland, _ := time.LoadLocation("Pacific/Auckland")
	nzTime := ConvertToUTCWithTimezone(jianTime, pacificAuckland)
	fmt.Println("转换为新西兰时间:", nzTime.Format("2006-01-02 15:04:05.000 MST"))

	// 测试新历法 -> 公历
	if testTime, err := NewJianCalendarDateTime("30-1-1 0:0:0.000"); err != nil {
		println(err.Error())
	} else {
		hostname, _ := os.Hostname()
		println("30-1-1 0:0:0.000 转为 UTC 时间: ", ConvertToUTC(testTime).Format("2006-01-02 15-04-05.000"), ", data from ->", hostname)
	}
}
