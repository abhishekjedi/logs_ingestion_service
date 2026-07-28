package dto

type ReleaseHealth struct {
	Release         string `json:"release"`
	SessionsTotal   uint64 `json:"sessions_total"`
	SessionsErrored uint64 `json:"sessions_errored"`
}
