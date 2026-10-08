package req

// UpdateMapConfigReq 更新地图服务配置请求参数（全量更新）。
type UpdateMapConfigReq struct {
	AmapKey          string `json:"amap_key"`
	AmapSecurityCode string `json:"amap_security_code"`
	GoogleKey        string `json:"google_key"`
}
