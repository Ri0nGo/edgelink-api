package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/model"
	"edgelink-api/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type redisReply struct {
	value string
	err   error
}

type persistenceRedisStub struct {
	redis.Cmdable
	replies map[string]redisReply
	batches int
}

func (r *persistenceRedisStub) Pipeline() redis.Pipeliner {
	r.batches++
	return &persistencePipelineStub{replies: r.replies}
}

type persistencePipelineStub struct {
	redis.Pipeliner
	replies map[string]redisReply
	cmds    []redis.Cmder
}

func (p *persistencePipelineStub) HGet(_ context.Context, key, field string) *redis.StringCmd {
	reply, ok := p.replies[key+":"+field]
	if !ok {
		reply.err = redis.Nil
	}
	cmd := redis.NewStringResult(reply.value, reply.err)
	p.cmds = append(p.cmds, cmd)
	return cmd
}

func (p *persistencePipelineStub) Exec(context.Context) ([]redis.Cmder, error) {
	for _, cmd := range p.cmds {
		if cmd.Err() != nil {
			return p.cmds, cmd.Err()
		}
	}
	return p.cmds, nil
}

func initTestLogger(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "persistence.log")
	file, err := os.Create(path)
	require.NoError(t, err)
	// InitLogger captures Stdout as its writer. Own the file here so Windows can
	// close it before TempDir cleanup; production FilePath has no Close API.
	stdout := os.Stdout
	os.Stdout = file
	err = logger.InitLogger(logger.LogConfig{Level: "info"})
	os.Stdout = stdout
	t.Cleanup(func() {
		require.NoError(t, logger.InitLogger(logger.LogConfig{Level: "info"}))
		require.NoError(t, file.Close())
	})
	require.NoError(t, err)
	return path
}

func TestGetDatasKeepsOtherDevicesWhenPropertyIsMissing(t *testing.T) {
	logPath := initTestLogger(t)
	for _, missingIndex := range []int{0, 1, 2} {
		items := []DevicePropItem{{DeviceId: 4, PropertyId: 1, PropertyKey: "t"}, {DeviceId: 5, PropertyId: 1, PropertyKey: "t"}, {DeviceId: 7, PropertyId: 1, PropertyKey: "t"}}
		replies := map[string]redisReply{"device:4:data:t": {value: "28.6"}, "device:5:data:t": {value: "29.7"}, "device:7:data:t": {value: "29.5"}}
		items[missingIndex].PropertyKey = "missing"
		client := &persistenceRedisStub{replies: replies}
		p := NewMySQLPersistence(client, nil, 200, 100)
		data, err := p.GetDatas(context.Background(), items)
		require.NoError(t, err)
		require.Len(t, data, 2)
		for _, row := range data {
			require.NotEqual(t, items[missingIndex].DeviceId, row.DeviceId)
		}
	}
	logs, err := os.ReadFile(logPath)
	require.NoError(t, err)
	require.Contains(t, string(logs), "persistence redis field missing")
	require.Contains(t, string(logs), "device_id=4")
	require.Contains(t, string(logs), "property_id=1")
	require.Contains(t, string(logs), "property_key=missing")
	require.Contains(t, string(logs), "redis_key=device:4:data")
}

func TestGetDatasPreservesValidValuesAcrossFailedBatches(t *testing.T) {
	initTestLogger(t)
	readErr := errors.New("WRONGTYPE test failure")
	client := &persistenceRedisStub{replies: map[string]redisReply{
		"device:4:data:t": {err: readErr},
		"device:5:data:t": {value: "29.7"},
		"device:7:data:t": {value: "29.5"},
	}}
	p := NewMySQLPersistence(client, nil, 200, 2)
	data, err := p.GetDatas(context.Background(), []DevicePropItem{
		{DeviceId: 4, PropertyId: 1, PropertyKey: "t"},
		{DeviceId: 5, PropertyId: 1, PropertyKey: "t"},
		{DeviceId: 7, PropertyId: 1, PropertyKey: "t"},
	})
	require.ErrorIs(t, err, readErr)
	require.Len(t, data, 2)
	require.Equal(t, 5, data[0].DeviceId)
	require.Equal(t, 7, data[1].DeviceId)
	require.Equal(t, 2, client.batches)
}

func TestEmptyAndInvalidRedisData(t *testing.T) {
	initTestLogger(t)
	client := &persistenceRedisStub{replies: map[string]redisReply{
		"device:4:data:text": {value: "invalid"},
		"device:4:data:nan":  {value: "NaN"},
		"device:4:data:inf":  {value: "+Inf"},
	}}
	p := NewMySQLPersistence(client, nil, 200, 100)
	data, err := p.GetDatas(context.Background(), nil)
	require.NoError(t, err)
	require.Empty(t, data)
	require.Zero(t, client.batches)
	data, err = p.GetDatas(context.Background(), []DevicePropItem{
		{DeviceId: 4, PropertyId: 1, PropertyKey: "missing"},
		{DeviceId: 4, PropertyId: 2, PropertyKey: "text"},
		{DeviceId: 4, PropertyId: 3, PropertyKey: "nan"},
		{DeviceId: 4, PropertyId: 4, PropertyKey: "inf"},
	})
	require.NoError(t, err)
	require.Empty(t, data)
	require.NoError(t, p.BatchSave(context.Background(), data))
}

func TestBatchSaveReturnsInsertFailure(t *testing.T) {
	initTestLogger(t)
	p := NewMySQLPersistence(nil, nil, 1, 100)
	err := p.BatchSave(context.Background(), []model.HistoryData{{DeviceId: 4}, {DeviceId: 5}})
	require.ErrorContains(t, err, "insert batch 1: db is nil")
	require.ErrorContains(t, err, "insert batch 2: db is nil")
}

type cyclePersisterStub struct {
	data      []model.HistoryData
	readErr   error
	saveErr   error
	readCalls int
	saveCalls int
	saved     []model.HistoryData
}

func (p *cyclePersisterStub) GetDatas(context.Context, []DevicePropItem) ([]model.HistoryData, error) {
	p.readCalls++
	return p.data, p.readErr
}

func (p *cyclePersisterStub) BatchSave(_ context.Context, data []model.HistoryData) error {
	p.saveCalls++
	p.saved = data
	return p.saveErr
}

func TestPersistenceCycleSavesPartialReadAndReportsSaveFailure(t *testing.T) {
	for _, saveErr := range []error{nil, errors.New("mysql unavailable")} {
		logPath := initTestLogger(t)
		backend := &cyclePersisterStub{data: []model.HistoryData{{DeviceId: 4}}, readErr: errors.New("redis partial failure"), saveErr: saveErr}
		p := NewGenericPersistence(context.Background(), backend)
		p.handlerDevicePropCreatedOrUpdated([]*dataloader.DevicePropInfo{{DeviceId: 4, PropertyId: 1, PropertyKey: "t"}})
		p.runOnce()
		require.Equal(t, 1, backend.saveCalls)
		require.Equal(t, backend.data, backend.saved)
		logs, err := os.ReadFile(logPath)
		require.NoError(t, err)
		if saveErr != nil {
			require.Contains(t, string(logs), "batch save data failed")
			require.NotContains(t, string(logs), "persistence cycle completed")
		} else {
			require.Contains(t, string(logs), "saved_rows=1")
			require.Contains(t, string(logs), "read_failed=true")
		}
	}
}

func TestPersistenceCycleSkipsEmptyData(t *testing.T) {
	initTestLogger(t)
	backend := &cyclePersisterStub{}
	p := NewGenericPersistence(context.Background(), backend)
	p.runOnce()
	require.Zero(t, backend.readCalls)
	p.handlerDevicePropCreatedOrUpdated([]*dataloader.DevicePropInfo{{DeviceId: 4, PropertyId: 1, PropertyKey: "t"}})
	p.runOnce()
	require.Equal(t, 1, backend.readCalls)
	require.Zero(t, backend.saveCalls)
}

func TestConfigEventLogsOldAndNewKey(t *testing.T) {
	logPath := initTestLogger(t)
	p := NewGenericPersistence(context.Background(), nil)
	for _, key := range []string{"t", "temperature"} {
		err := p.Notify(context.Background(), &notify.Event{
			NotifyType: notify.DevicePropertyNotifyType, Operation: notify.OperationTypeUpdated,
			DeviceKey: "sensor", Ts: 123,
			Payload: []*dataloader.DevicePropInfo{{DeviceId: 4, PropertyId: 1, PropertyKey: key}},
		})
		require.NoError(t, err)
	}
	require.Equal(t, []DevicePropItem{{DeviceId: 4, PropertyId: 1, PropertyKey: "temperature"}}, p.getAllDeviceProps())
	logs, err := os.ReadFile(logPath)
	require.NoError(t, err)
	require.Contains(t, string(logs), "old_property_key=t property_key=temperature")
	require.Contains(t, string(logs), "event_ts=123")
	err = p.Notify(context.Background(), &notify.Event{
		NotifyType: notify.DevicePropertyNotifyType, Operation: notify.OperationTypeDeleted,
		Payload: []*dataloader.DevicePropInfo{{DeviceId: 4, PropertyId: 1}},
	})
	require.NoError(t, err)
	require.Empty(t, p.getAllDeviceProps())
}
