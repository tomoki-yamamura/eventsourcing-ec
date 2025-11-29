package tenant

import (
	"context"
	"database/sql"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	appErrors "github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type TenantPolicyReadModelImpl struct {
	tx repository.Transaction
}

func NewTenantPolicyReadModel(tx repository.Transaction) readmodelstore.TenantPolicyStore {
	return &TenantPolicyReadModelImpl{
		tx: tx,
	}
}

func (t *TenantPolicyReadModelImpl) Get(ctx context.Context, tenantID string) (*dto.TenantPolicyViewDTO, error) {
	var policy *dto.TenantPolicyViewDTO
	err := t.tx.RWTx(ctx, func(ctx context.Context) error {
		tx, err := transaction.GetTx(ctx)
		if err != nil {
			return err
		}

		policyQuery := `
			SELECT id, title, abandoned_minutes, quiet_time_from, quiet_time_to, created_at, updated_at, version
			FROM tenant_cart_abandoned_policies 
			WHERE id = ?
		`

		var policyView dto.TenantPolicyViewDTO

		err = tx.QueryRowContext(ctx, policyQuery, tenantID).Scan(
			&policyView.ID,
			&policyView.Title,
			&policyView.AbandonedMinutes,
			&policyView.QuietTimeFrom,
			&policyView.QuietTimeTo,
			&policyView.CreatedAt,
			&policyView.UpdatedAt,
			&policyView.Version,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return appErrors.NotFound.New("tenant policy not found")
			}
			return appErrors.QueryError.Wrap(err, "failed to get tenant policy")
		}

		policy = &policyView
		return nil
	})
	if err != nil {
		return nil, err
	}
	return policy, nil
}

func (t *TenantPolicyReadModelImpl) Upsert(ctx context.Context, tenantID string, view *dto.TenantPolicyViewDTO) error {
	return t.tx.RWTx(ctx, func(ctx context.Context) error {
		tx, err := transaction.GetTx(ctx)
		if err != nil {
			return err
		}

		policyQuery := `
			INSERT INTO tenant_cart_abandoned_policies (id, title, abandoned_minutes, quiet_time_from, quiet_time_to, created_at, updated_at, version)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				title = VALUES(title),
				abandoned_minutes = VALUES(abandoned_minutes),
				quiet_time_from = VALUES(quiet_time_from),
				quiet_time_to = VALUES(quiet_time_to),
				updated_at = VALUES(updated_at),
				version = VALUES(version)
		`

		_, err = tx.ExecContext(ctx, policyQuery,
			view.ID,
			view.Title,
			view.AbandonedMinutes,
			view.QuietTimeFrom,
			view.QuietTimeTo,
			view.CreatedAt,
			view.UpdatedAt,
			view.Version,
		)
		if err != nil {
			return appErrors.RepositoryError.Wrap(err, "failed to upsert tenant policy")
		}

		return nil
	})
}