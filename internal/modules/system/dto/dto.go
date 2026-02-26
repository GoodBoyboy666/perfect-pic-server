package dto

type InitRequest struct {
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required"`
	SiteName        string `json:"site_name" binding:"required"`
	SiteDescription string `json:"site_description" binding:"required"`
}

type SystemInfoResponse struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
}

type ServerStatsResponse struct {
	ImageCount   int64              `json:"image_count"`
	StorageUsage int64              `json:"storage_usage"`
	UserCount    int64              `json:"user_count"`
	SystemInfo   SystemInfoResponse `json:"system_info"`
}
