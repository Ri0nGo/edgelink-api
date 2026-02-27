package svc

import (
	"context"
	"edgelink-api/internal/model"
	bizErr "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/repo"
	"fmt"
	"time"
)

type Step uint8

const (
	OneMinuteStep Step = iota + 1
	OneHourStep
	OneDayStep
	OneWeekStep
	OneMonthStep
	OneYearStep
)

type IMetricSvc interface {
	GetTimeSeriesHistoryData(ctx context.Context, deviceIds, propIds []int, begin, end int64) (map[string][][2]any, error)
}

type MetricSvc struct {
	hisDataRepo repo.IHistoryDataRepo
}

// GetTimeSeriesHistoryData 获取设备下所有属性的数据
func (m *MetricSvc) GetTimeSeriesHistoryData(ctx context.Context, deviceIds, propIds []int,
	begin, end int64) (map[string][][2]any, error) {
	if end < begin {
		return nil, bizErr.NewBizError("开始时间不能大于结束时间")
	}

	datas, err := m.hisDataRepo.QueryTimeSeriesHistoryData(ctx, deviceIds, propIds, time.Unix(begin, 0), time.Unix(end, 0))
	if err != nil {
		return nil, err
	}

	var results = make(map[string][][2]any)
	for _, data := range datas {
		key := fmt.Sprintf("%d:%d", data.DeviceId, data.PropertyId)
		results[key] = append(results[key], [2]any{data.Ts.Unix(), data.Value})
	}
	return results, nil
}

func NewMetricSvc(hisDataRepo repo.IHistoryDataRepo) IMetricSvc {
	return &MetricSvc{hisDataRepo: hisDataRepo}

}

// FilledValueByNil 使用空值填充数据
func FilledValueByNil(data []model.HistoryData, begin, end time.Time, step Step) []model.HistoryData {
	var (
		result         = make([]model.HistoryData, 0)
		hasValueTime   = make(map[string]struct{})
		originDataTime = make(map[string]struct{})
	)
	lastTime := begin
	for _, d := range data {
		// 排掉 data 含有的重复数据
		uniqueKey := getUniqueKeyWithHistoryData(d.DeviceId, d.PropertyId, d.Ts.Unix())
		_, ok := originDataTime[uniqueKey]
		if !ok {
			originDataTime[uniqueKey] = struct{}{}
		} else {
			continue
		}
		// 补充当前时间之前的数据
		lastUniqueKey := getUniqueKeyWithHistoryData(d.DeviceId, d.PropertyId, lastTime.Unix())
		for lastTime.Before(d.Ts) {
			result = append(result, model.HistoryData{
				Ts: lastTime,
			})
			hasValueTime[lastUniqueKey] = struct{}{}
			lastTime = getNextTimeByStep(step, lastTime)
		}

		// 添加当前值
		if _, ok := hasValueTime[lastUniqueKey]; !ok {
			result = append(result, model.HistoryData{
				DeviceId:   d.DeviceId,
				PropertyId: d.PropertyId,
				Ts:         lastTime,
				Value:      d.Value,
			})
			hasValueTime[lastUniqueKey] = struct{}{}
			lastTime = getNextTimeByStep(step, lastTime)
		}
	}
	// 补充末尾时间的数据
	for lastTime.Before(end) || lastTime.Equal(end) {
		result = append(result, model.HistoryData{
			Ts: lastTime,
		})
		lastTime = getNextTimeByStep(step, lastTime)
	}
	return result
}

func getUniqueKeyWithHistoryData(deviceId, PropertyId int, ts int64) string {
	return fmt.Sprintf("%d_%d_%d", deviceId, PropertyId, ts)
}

func getNextTimeByStep(step Step, lastTime time.Time) time.Time {
	switch step {
	case OneMinuteStep:
		lastTime = lastTime.Add(time.Minute)
	case OneHourStep:
		lastTime = lastTime.Add(time.Hour)
	case OneDayStep:
		lastTime = lastTime.AddDate(0, 0, 1)
	case OneWeekStep:
		lastTime = lastTime.AddDate(0, 0, 7)
	case OneMonthStep:
		lastTime = lastTime.AddDate(0, 1, 0)
	case OneYearStep:
		lastTime = lastTime.AddDate(1, 0, 0)
	default:
		logger.Error("step is not support, default use hour interval", "step", step)
		lastTime = lastTime.Add(time.Hour)
	}
	return lastTime
}

// FilledValueByNilV2 使用空值填充数据
// 前提：data 已按 Ts 升序排列（全局排序，不要求同 device+property 连续）
func FilledValueByNilV2(data []model.HistoryData, begin, end time.Time, step Step) []model.HistoryData {
	if len(data) == 0 {
		// 如果没有任何数据，直接补全时间轴（全空值）
		return fillEmptyTimeline(begin, end, step)
	}

	result := make([]model.HistoryData, 0, 1024)
	hasValue := make(map[string]struct{}) // "device_property_tsUnix" → 是否已有值

	// 用一个指针遍历已排序的原始数据
	i := 0 // data 的游标
	current := begin

	for !current.After(end) {
		appended := false

		// 处理所有在这个 current 时间点的数据（可能有多个 device+property）
		for i < len(data) && data[i].Ts.Equal(current) {
			d := data[i]
			key := getUniqueKeyWithHistoryData(d.DeviceId, d.PropertyId, current.Unix())

			// 避免同一个时间点重复添加（虽然你说已排序，但可能有脏数据）
			if _, exist := hasValue[key]; !exist {
				result = append(result, model.HistoryData{
					DeviceId:   d.DeviceId,
					PropertyId: d.PropertyId,
					Ts:         current,
					Value:      d.Value, // 这里才是真正的值
				})
				hasValue[key] = struct{}{}
			}
			i++ // 消费这条数据
			appended = true
		}
		
		if !appended {
			// 如果你业务允许“无 device/property 的空行”，可以这样：
			result = append(result, model.HistoryData{
				Ts: current,
			})
		}

		current = getNextTimeByStep(step, current)
	}

	return result
}

func fillEmptyTimeline(begin, end time.Time, step Step) []model.HistoryData {
	var result []model.HistoryData
	current := begin

	for !current.After(end) {
		result = append(result, model.HistoryData{
			Ts: current,
			// DeviceId:   0,          // 可选：显式设为 0 或保持默认
			// PropertyId: 0,
			// Value:      0.0,        // 或保持零值，代表 null
		})
		current = getNextTimeByStep(step, current)
	}
	return result
}

// Group 表示一个监控项的唯一标识
type Group struct {
	DeviceId   int
	PropertyId int
}

// FilledValueByNilGrouped 按 (deviceId, propertyId) 分组补齐完整时间序列
//   - 每个已出现的组合都会输出从 begin 到 end 的完整时间点
//   - 有原始值的时间点填充 Value，无值的时间点 Value 保持零值（代表 null）
//   - 前提：data 已按 Ts 升序（全局排序）
//   - 输出按 Ts 排序，然后按 Group 顺序（稳定）
func FilledValueByNilV3(data []model.HistoryData, begin, end time.Time, step Step) []model.HistoryData {
	if begin.After(end) {
		return nil
	}

	// 1. 收集所有出现的 group 和每个 group 的值（map[tsUnix]value，后值覆盖前值）
	groups := make(map[Group]map[int64]float64)
	for _, d := range data {
		if d.Ts.Before(begin) || d.Ts.After(end) {
			continue // 过滤区间外数据
		}
		g := Group{DeviceId: d.DeviceId, PropertyId: d.PropertyId}
		if _, ok := groups[g]; !ok {
			groups[g] = make(map[int64]float64, 128) // 预估容量，可根据业务调整
		}
		groups[g][d.Ts.Unix()] = d.Value
	}

	if len(groups) == 0 {
		// 没有任何数据，直接返回空（或按需调用 fillEmptyTimeline）
		return nil
	}

	// 2. 计算大致的时间点数量，用于预分配 result 容量
	approxPointCount := int(end.Sub(begin)/getStepDuration(step)) + 2
	result := make([]model.HistoryData, 0, len(groups)*approxPointCount)

	// 3. 遍历时间轴，为每个 group 生成一行
	current := begin
	for !current.After(end) {
		tsUnix := current.Unix()

		for g, valueMap := range groups {
			item := model.HistoryData{
				DeviceId:   g.DeviceId,
				PropertyId: g.PropertyId,
				Ts:         current,
			}
			if val, ok := valueMap[tsUnix]; ok {
				item.Value = val
			} // else Value 保持 0.0，表示 null

			result = append(result, item)
		}

		current = getNextTimeByStep(step, current)
	}

	return result
}

// getStepDuration 返回 step 对应的 time.Duration（用于估算容量）
func getStepDuration(step Step) time.Duration {
	switch step {
	case OneMinuteStep:
		return time.Minute
	case OneHourStep:
		return time.Hour
	case OneDayStep:
		return 24 * time.Hour
	case OneWeekStep:
		return 7 * 24 * time.Hour
	case OneMonthStep:
		return 30 * 24 * time.Hour // 粗略
	case OneYearStep:
		return 365 * 24 * time.Hour // 粗略
	default:
		return time.Hour
	}
}
