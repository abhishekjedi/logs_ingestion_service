package dto

import dbdto "error-logging/db/dto"

type IssueListResult struct {
	Issues []dbdto.Issue
	Total  int64
}
