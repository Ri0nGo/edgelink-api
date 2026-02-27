package svc

import (
	"edgelink-api/internal/model"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestFilledValueByNil(t *testing.T) {
	loc := time.UTC
	parse := func(s string) time.Time {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
		return t
	}

	tests := []struct {
		name    string
		begin   time.Time
		end     time.Time
		step    Step
		input   []model.HistoryData
		wantLen int
	}{
		{
			name:  "normal_single_device",
			begin: parse("2025-01-01 10:00:00"),
			end:   parse("2025-01-01 10:03:00"),
			step:  OneMinuteStep,
			input: []model.HistoryData{
				{1, 100, parse("2025-01-01 10:01:30"), 23.5},
				{1, 100, parse("2025-01-01 10:02:45"), 24.0},
			},
			wantLen: 4, // 10:00,10:01,10:02,10:03
		},
		{
			name:    "empty_data",
			begin:   parse("2025-06-01 00:00:00"),
			end:     parse("2025-06-01 01:00:00"),
			step:    OneHourStep,
			input:   []model.HistoryData{},
			wantLen: 2,
		},
		{
			name:  "one day",
			begin: parse("2025-06-01 00:00:00"),
			end:   parse("2025-06-02 00:00:00"),
			step:  OneMinuteStep,
			input: []model.HistoryData{
				{1, 100, parse("2025-06-01 10:01:30"), 23.5},
				{1, 100, parse("2025-06-01 11:02:45"), 24.0},
			},
			wantLen: 1441,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilledValueByNil(tt.input, tt.begin, tt.end, tt.step)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestFilledValueByNilV2(t *testing.T) {
	loc := time.UTC
	parse := func(s string) time.Time {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
		return t
	}

	tests := []struct {
		name    string
		begin   time.Time
		end     time.Time
		step    Step
		input   []model.HistoryData
		wantLen int
	}{
		{
			name:  "normal_single_device",
			begin: parse("2025-01-01 10:00:00"),
			end:   parse("2025-01-01 10:03:00"),
			step:  OneMinuteStep,
			input: []model.HistoryData{
				{1, 100, parse("2025-01-01 10:01:30"), 23.5},
				{1, 100, parse("2025-01-01 10:02:45"), 24.0},
			},
			wantLen: 4, // 10:00,10:01,10:02,10:03
		},
		{
			name:    "empty_data",
			begin:   parse("2025-06-01 00:00:00"),
			end:     parse("2025-06-01 01:00:00"),
			step:    OneHourStep,
			input:   []model.HistoryData{},
			wantLen: 2,
		},
		{
			name:  "one day",
			begin: parse("2025-06-01 00:00:00"),
			end:   parse("2025-06-02 00:00:00"),
			step:  OneMinuteStep,
			input: []model.HistoryData{
				{1, 100, parse("2025-06-01 10:01:30"), 23.5},
				{1, 100, parse("2025-06-01 11:02:45"), 24.0},
			},
			wantLen: 1441,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilledValueByNilV2(tt.input, tt.begin, tt.end, tt.step)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// --------------------- benchmark ---------------------
func BenchmarkFilledValueByNil(b *testing.B) {
	// 小数据集
	b.Run("Small-10dev-5prop-200pts", func(b *testing.B) {
		data := generateTestData(10, 5, 200)
		begin := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		end := begin.Add(24 * time.Hour)
		step := OneMinuteStep

		b.Run("V1", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNil(data, begin, end, step)
			}
		})

		b.Run("V2", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNilV2(data, begin, end, step)
			}
		})
		b.Run("V3", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNilV3(data, begin, end, step)
			}
		})
	})

	// 中等数据集
	b.Run("Medium-50dev-10prop-1000pts", func(b *testing.B) {
		data := generateTestData(50, 10, 1000)
		begin := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		end := begin.Add(7 * 24 * time.Hour)
		step := OneHourStep

		b.Run("V1", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNil(data, begin, end, step)
			}
		})

		b.Run("V2", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNilV2(data, begin, end, step)
			}
		})

		b.Run("V3", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNilV3(data, begin, end, step)
			}
		})
	})

	// 大数据集（模拟真实压力）
	b.Run("Large-200dev-20prop-5000pts", func(b *testing.B) {
		data := generateTestData(200, 20, 5000)
		begin := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		end := begin.Add(30 * 24 * time.Hour)
		step := OneHourStep

		b.Run("V1", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNil(data, begin, end, step)
			}
		})

		b.Run("V2", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNilV2(data, begin, end, step)
			}
		})

		b.Run("V3", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNilV3(data, begin, end, step)
			}
		})
	})

	// 大数据集（模拟真实压力）
	b.Run("Large-10dev-10prop-one-minute-year", func(b *testing.B) {
		data := generateTestData(10, 10, 5000)
		begin := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		end := begin.AddDate(1, 0, 0)
		step := OneMinuteStep

		b.Run("V1", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNil(data, begin, end, step)
			}
		})

		b.Run("V2", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNilV2(data, begin, end, step)
			}
		})

		b.Run("V3", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = FilledValueByNilV3(data, begin, end, step)
			}
		})
	})
}

// --------------------- 生成测试数据 ---------------------
func generateTestData(nDevices, nPropsPerDev, nPointsPerSeries int) []model.HistoryData {
	var data []model.HistoryData
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	for dev := 1; dev <= nDevices; dev++ {
		for prop := 1; prop <= nPropsPerDev; prop++ {
			for i := 0; i < nPointsPerSeries; i++ {
				// 稀疏：随机跳过一些点
				if rand.Intn(3) == 0 {
					continue
				}
				ts := baseTime.Add(time.Duration(i*5+rand.Intn(3)) * time.Minute)
				data = append(data, model.HistoryData{
					DeviceId:   dev,
					PropertyId: prop + 1000*dev,
					Ts:         ts,
					Value:      rand.Float64() * 100,
				})
			}
		}
	}

	// 排序（生产中假设已排序，这里强制一下）
	sort.Slice(data, func(i, j int) bool {
		return data[i].Ts.Before(data[j].Ts)
	})

	return data
}

func TestName(t *testing.T) {
	var s int
	s = 3025
	fmt.Println(byte(s))
}
