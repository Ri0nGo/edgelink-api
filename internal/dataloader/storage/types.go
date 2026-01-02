package storage

import (
	"context"
)

type Storage interface {
	SaveData(ctx context.Context, deviceId int, data []byte) error
	SaveStatus(ctx context.Context, deviceId int, data []byte) error
	Close() error
}
