package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"edgelink-api/internal/model"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/repo"
	"edgelink-api/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type updateDeviceRepoStub struct {
	repo.IDeviceRepo
	lookupErr error
	updated   *model.Device
}

func (r *updateDeviceRepoStub) GetDeviceById(_ context.Context, id int) (model.Device, error) {
	return model.Device{Id: id, Key: "esp32c3-dht11-001"}, r.lookupErr
}

func (r *updateDeviceRepoStub) UpdateDevice(_ context.Context, device *model.Device) error {
	r.updated = device
	return nil
}

type updateDeviceProductRepoStub struct {
	repo.IProductRepo
}

func (r *updateDeviceProductRepoStub) GetProductById(_ context.Context, id int) (model.Product, error) {
	return model.Product{Id: id}, nil
}

func TestUpdateDeviceKeyValidation(t *testing.T) {
	require.NoError(t, logger.InitLogger(logger.LogConfig{Level: "error"}))
	for _, tt := range []struct {
		name       string
		keyJSON    string
		lookupErr  error
		wantCode   int
		wantMsg    string
		wantUpdate bool
	}{
		{name: "unchanged key", keyJSON: `,"key":"esp32c3-dht11-001"`, wantMsg: "success", wantUpdate: true},
		{name: "omitted key", wantMsg: "success", wantUpdate: true},
		{name: "empty key", keyJSON: `,"key":""`, wantMsg: "success", wantUpdate: true},
		{name: "changed key", keyJSON: `,"key":"esp32c3-dht11-002"`, wantCode: 10001, wantMsg: "设备标识符不能修改"},
		{name: "missing device", lookupErr: gorm.ErrRecordNotFound, wantCode: 10001, wantMsg: "设备不存在"},
		{name: "lookup failure", lookupErr: errors.New("database unavailable"), wantCode: 500, wantMsg: "internal error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			deviceRepo := &updateDeviceRepoStub{lookupErr: tt.lookupErr}
			deviceSvc := svc.NewDeviceSvc(deviceRepo, &updateDeviceProductRepoStub{}, nil, nil)
			router := gin.New()
			NewDeviceApi(deviceSvc).RegistryRouter(router.Group("/api/edgelink"))
			body := `{"id":4,"name":"dht11温湿度传感器","product_id":4,"description":""` + tt.keyJSON + `}`
			request := httptest.NewRequest(http.MethodPost, "/api/edgelink/device/update", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			var result struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
			require.Equal(t, tt.wantCode, result.Code)
			require.Equal(t, tt.wantMsg, result.Msg)
			if tt.wantUpdate {
				require.NotNil(t, deviceRepo.updated)
				require.Equal(t, 4, deviceRepo.updated.Id)
				require.Equal(t, "dht11温湿度传感器", deviceRepo.updated.Name)
				require.Equal(t, 4, deviceRepo.updated.ProductId)
				require.Empty(t, deviceRepo.updated.Key, "updates must not write the device key")
			} else {
				require.Nil(t, deviceRepo.updated)
			}
		})
	}
}
