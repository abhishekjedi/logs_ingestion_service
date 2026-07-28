package dto

type IssueListFilter struct {
	ServiceID uint64
	Status    string
	Sort      string
	Order     string
	Limit     int
	Offset    int
}
