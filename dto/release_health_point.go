package dto

type ReleaseHealthPoint struct {
	Release         string  `json:"release"`
	SessionsTotal   uint64  `json:"sessions_total"`
	SessionsErrored uint64  `json:"sessions_errored"`
	CrashFreeRate   float64 `json:"crash_free_rate"`
}
