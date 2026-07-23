package impl

import (
	"context"
	"time"

	"error-logging/constants"
	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
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
	// INSERT ... ON DUPLICATE KEY UPDATE id=id (no-op) so a concurrent first-sighting
	// never errors; RowsAffected tells us whether we created it.
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(issue)
	if res.Error != nil {
		return nil, false, res.Error
	}
	created := res.RowsAffected > 0

	// Fetch the canonical row (ours if created, otherwise the pre-existing one).
	var out dbdto.Issue
	if err := r.db.WithContext(ctx).
		Where("service_id = ? AND fingerprint = ?", issue.ServiceID, issue.Fingerprint).
		First(&out).Error; err != nil {
		return nil, false, err
	}
	return &out, created, nil
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
