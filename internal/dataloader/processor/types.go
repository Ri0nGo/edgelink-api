package processor

type DataProcessor interface {
	Start() error
	Close()
}

type DeviceInfo struct {
	DeviceId          int
	DeviceKey         string
	ProductIdentifier string
}
