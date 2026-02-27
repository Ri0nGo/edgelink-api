package dto

type ReqTimeSeriesData struct {
	DeviceIds   []int `json:"device_ids"`
	PropertyIds []int `json:"property_ids"`
	BeginAndEnd
}
