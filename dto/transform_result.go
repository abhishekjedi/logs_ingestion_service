package dto

import dbdto "error-logging/db/dto"

type TransformResult struct {
	Logs        []dbdto.LogRow
	ErrorEvents []dbdto.ErrorEventRow
}
