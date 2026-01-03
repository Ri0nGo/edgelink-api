package dataloader

type ProviderType string

const (
	MQTTProvider ProviderType = "mqtt"
	HTTPProvider ProviderType = "http"
)
