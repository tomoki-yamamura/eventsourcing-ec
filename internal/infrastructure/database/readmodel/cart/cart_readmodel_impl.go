package cart

import (
	"context"
	"database/sql"
	"strings"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	appErrors "github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type CartReadModelImpl struct {
	tx repository.Transaction
}

func NewCartReadModel(tx repository.Transaction) readmodelstore.CartStore {
	return &CartReadModelImpl{
		tx: tx,
	}
}

func (c *CartReadModelImpl) Get(ctx context.Context, aggregateID string) (*dto.CartViewDTO, error) {
	var cart *dto.CartViewDTO
	err := c.tx.RWTx(ctx, func(ctx context.Context) error {
		tx, err := transaction.GetTx(ctx)
		if err != nil {
			return err
		}

		// Get cart basic info
		cartQuery := `
			SELECT id, user_id, tenant_id, status, total_amount, item_count, created_at, updated_at, purchased_at, version
			FROM carts 
			WHERE id = ?
		`

		var cartView dto.CartViewDTO
		var purchasedAt sql.NullTime

		err = tx.QueryRowContext(ctx, cartQuery, aggregateID).Scan(
			&cartView.ID,
			&cartView.UserID,
			&cartView.TenantID,
			&cartView.Status,
			&cartView.TotalAmount,
			&cartView.ItemCount,
			&cartView.CreatedAt,
			&cartView.UpdatedAt,
			&purchasedAt,
			&cartView.Version,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return appErrors.NotFound.New("cart not found")
			}
			return appErrors.QueryError.Wrap(err, "failed to get cart")
		}

		if purchasedAt.Valid {
			cartView.PurchasedAt = &purchasedAt.Time
		}

		// Get cart items
		itemsQuery := `
			SELECT id, cart_id, name, price
			FROM cart_items 
			WHERE cart_id = ?
		`

		rows, err := tx.QueryContext(ctx, itemsQuery, aggregateID)
		if err != nil {
			return appErrors.QueryError.Wrap(err, "failed to get cart items")
		}
		defer rows.Close()

		var items []dto.CartItemViewDTO
		itemCount := 0
		for rows.Next() {
			var item dto.CartItemViewDTO
			err := rows.Scan(
				&item.ID,
				&item.CartID,
				&item.Name,
				&item.Price,
			)
			if err != nil {
				return appErrors.QueryError.Wrap(err, "failed to scan cart item")
			}
			items = append(items, item)
			itemCount++
		}

		if err := rows.Err(); err != nil {
			return appErrors.QueryError.Wrap(err, "rows iteration error")
		}

		cartView.Items = items
		cart = &cartView
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (c *CartReadModelImpl) Upsert(ctx context.Context, aggregateID string, view *dto.CartViewDTO) error {
	return c.tx.RWTx(ctx, func(ctx context.Context) error {
		tx, err := transaction.GetTx(ctx)
		if err != nil {
			return err
		}

		// Upsert cart
		cartQuery := `
			INSERT INTO carts (id, user_id, tenant_id, status, total_amount, item_count, created_at, updated_at, purchased_at, version)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				user_id = VALUES(user_id),
				tenant_id = VALUES(tenant_id),
				status = VALUES(status),
				total_amount = VALUES(total_amount),
				item_count = VALUES(item_count),
				updated_at = VALUES(updated_at),
				purchased_at = VALUES(purchased_at),
				version = VALUES(version)
		`

		_, err = tx.ExecContext(ctx, cartQuery,
			view.ID,
			view.UserID,
			view.TenantID,
			view.Status,
			view.TotalAmount,
			view.ItemCount,
			view.CreatedAt,
			view.UpdatedAt,
			view.PurchasedAt,
			view.Version,
		)
		if err != nil {
			return appErrors.RepositoryError.Wrap(err, "failed to upsert cart")
		}

		deleteItemsQuery := `DELETE FROM cart_items WHERE cart_id = ?`
		_, err = tx.ExecContext(ctx, deleteItemsQuery, aggregateID)
		if err != nil {
			return appErrors.RepositoryError.Wrap(err, "failed to delete existing cart items")
		}

		if len(view.Items) > 0 {
			values := make([]interface{}, 0, len(view.Items)*4)
			placeholders := make([]string, 0, len(view.Items))

			for _, item := range view.Items {
				placeholders = append(placeholders, "(?, ?, ?, ?)")
				values = append(values, item.ID, item.CartID, item.Name, item.Price)
			}

			itemQuery := "INSERT INTO cart_items (id, cart_id, name, price) VALUES " +
				strings.Join(placeholders, ", ")

			_, err = tx.ExecContext(ctx, itemQuery, values...)
			if err != nil {
				return appErrors.RepositoryError.Wrap(err, "failed to bulk insert cart items")
			}
		}

		return nil
	})
}
