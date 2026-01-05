package utils

import "time"

func GetCurrentMinuteTime() time.Time {
	now := time.Now()
	return time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		0, // 秒清零
		0, // 纳秒清零
		now.Location(),
	)
}

func GetCurrentHourTime() time.Time {
	now := time.Now()
	return time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		0,
		0, // 秒清零
		0, // 纳秒清零
		now.Location(),
	)
}

func GetCurrentDayTime() time.Time {
	now := time.Now()
	return time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0, // 秒清零
		0, // 纳秒清零
		now.Location(),
	)
}

func GetCurrentMonthTime() time.Time {
	now := time.Now()
	return time.Date(
		now.Year(),
		now.Month(),
		0,
		0,
		0,
		0, // 秒清零
		0, // 纳秒清零
		now.Location(),
	)
}
