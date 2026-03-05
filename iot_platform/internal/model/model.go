package model

// DeviceData 是设备上报的数据结构
type DeviceData struct {
	DeviceID string  `json:"device_id"`
	Value    float64 `json:"value"`
}
