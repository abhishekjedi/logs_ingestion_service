package impl

import (
	"context"
	"time"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	"error-logging/dto"
	mysqlclient "error-logging/pkg/client/mysql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type issueRepository struct {
	db *gorm.DB
}

func NewIssueRepository(c *mysqlclient.Client) repository.IssueRepository {
	return &issueRepository{db: c.DB}
}

func (r *issueRepository) ResolveOrCreate(ctx context.Context, issue *dbdto.Issue) (*dbdto.Issue, bool, error) {

	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(issue)
	if res.Error != nil {
		return nil, false, res.Error
	}
	created := res.RowsAffected > 0

	var out dbdto.Issue
	if err := r.db.WithContext(ctx).
		Where("service_id = ? AND fingerprint = ?", issue.ServiceID, issue.Fingerprint).
		First(&out).Error; err != nil {
		return nil, false, err
	}
	return &out, created, nil
}

var issueSortColumns = map[string]string{
	"event_count": "event_count",
	"last_seen":   "last_seen",
	"first_seen":  "first_seen",
}

func (r *issueRepository) List(ctx context.Context, f dto.IssueListFilter) ([]dbdto.Issue, int64, error) {
	q := r.db.WithContext(ctx).Model(&dbdto.Issue{}).Where("service_id = ?", f.ServiceID)
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol, ok := issueSortColumns[f.Sort]
	if !ok {
		sortCol = "last_seen"
	}
	order := "DESC"
	if f.Order == "asc" {
		order = "ASC"
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var issues []dbdto.Issue
	err := q.Order(sortCol + " " + order).Limit(limit).Offset(f.Offset).Find(&issues).Error
	return issues, total, err
}

func (r *issueRepository) GetByID(ctx context.Context, id uint64) (*dbdto.Issue, error) {
	var issue dbdto.Issue
	if err := r.db.WithContext(ctx).First(&issue, id).Error; err != nil {
		return nil, err
	}
	return &issue, nil
}

func (r *issueRepository) UpdateStatsBatch(ctx context.Context, updates []dto.IssueStatsUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			if err := tx.Model(&dbdto.Issue{}).
				Where("id = ?", u.IssueID).
				Updates(map[string]any{
					"event_count":                u.EventCount,
					"affected_users_estimate":    u.AffectedUsers,
					"affected_sessions_estimate": u.AffectedSessions,
					"last_seen":                  u.LastSeen,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *issueRepository) MarkRegressed(ctx context.Context, id uint64) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&dbdto.Issue{}).
		Where("id = ? AND status = ?", id, constants.StatusResolved).
		Updates(map[string]any{
			"status":       constants.StatusRegressed,
			"regressed_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
