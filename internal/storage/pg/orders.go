package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

type ordersStmts struct {
	stmtAddOrder            *sql.Stmt
	stmtGetOrder            *sql.Stmt
	stmtGetUserOrders       *sql.Stmt
	stmtGetOrdersByStatuses *sql.Stmt
	stmtUpdateOrderStatus   *sql.Stmt
}

func prepareOrdersStmts(ctx context.Context, p *Pg) error {

	newOrdersStmts := ordersStmts{}

	var err error

	if newOrdersStmts.stmtAddOrder, err = p.db.PrepareContext(ctx, queryAddOrder); err != nil {
		return err
	}

	if newOrdersStmts.stmtGetOrder, err = p.db.PrepareContext(ctx, queryGetOrder); err != nil {
		return err
	}

	if newOrdersStmts.stmtGetUserOrders, err = p.db.PrepareContext(ctx, queryGetOrdersByUser); err != nil {
		return err
	}

	if newOrdersStmts.stmtGetOrdersByStatuses, err = p.db.PrepareContext(ctx, queryGetOrdersByStatuses); err != nil {
		return err
	}

	if newOrdersStmts.stmtUpdateOrderStatus, err = p.db.PrepareContext(ctx, queryUpdateOrderStatus); err != nil {
		return err
	}

	p.ordersStmts = &newOrdersStmts

	return nil
}

func (p *Pg) AddOrder(c context.Context, order *model.Order) error {
	log.Debug().Msg("Pg.AddOrder START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.AddOrder END")
		} else {
			log.Debug().Msg("Pg.AddOrder END")
		}
	}()

	_, err = p.ordersStmts.stmtAddOrder.ExecContext(c, order.UserID, order.Number, order.Status, order.Accrual, order.UploadedAt)
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok &&
			pgerrcode.IsIntegrityConstraintViolation(pgError.Code) &&
			pgError.ConstraintName == "orders_number_key" {

			existingOrder, errGetOrder := p.GetOrder(c, order.Number)
			if errGetOrder != nil {
				return fmt.Errorf(`pg: %w`, errGetOrder)
			}
			if order.UserID == existingOrder.UserID {
				return fmt.Errorf(`pg: %w: %s`, dberr.ErrOrderWasUploadedByCurrentUser, err)
			} else {
				return fmt.Errorf(`pg: %w: %s`, dberr.ErrOrderWasUploadedByAnotherUser, err)
			}
		}
		return fmt.Errorf(`pg: %w`, err)
	}

	return nil
}

func (p *Pg) GetOrdersByUser(c context.Context, userID int64) ([]model.Order, error) {
	log.Debug().Msg("Pg.GetOrdersByUser START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.GetOrdersByUser END")
		} else {
			log.Debug().Msg("Pg.GetOrdersByUser END")
		}
	}()

	rows, err := p.ordersStmts.stmtGetUserOrders.QueryContext(c, userID)
	if err != nil {
		return nil, fmt.Errorf(`pg: %w`, err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		currOrder := model.Order{}
		rows.Scan(&currOrder.UserID, &currOrder.Number, &currOrder.Status, &currOrder.Accrual, &currOrder.UploadedAt)
		orders = append(orders, currOrder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf(`pg: %w`, err)
	}

	return orders, nil
}

func (p *Pg) GetOrder(c context.Context, number string) (*model.Order, error) {
	log.Debug().Msg("Pg.GetOrder START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.GetOrder END")
		} else {
			log.Debug().Msg("Pg.GetOrder END")
		}
	}()

	var order model.Order
	err = p.ordersStmts.stmtGetOrder.QueryRowContext(c, number).
		Scan(&order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(`pg: %w: %s`, dberr.ErrOrderIsNotExists, err)
		}
		return nil, fmt.Errorf(`pg: %w`, err)
	}

	return &order, nil
}

func (p *Pg) GetOrdersByStatuses(statuses []string) ([]model.Order, error) {
	log.Debug().Msg("Pg.GetOrdersByStatuses START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.GetOrdersByStatuses END")
		} else {
			log.Debug().Msg("Pg.GetOrdersByStatuses END")
		}
	}()

	rows, err := p.ordersStmts.stmtGetOrdersByStatuses.Query(pq.Array(statuses))
	if err != nil {
		return nil, fmt.Errorf(`pg: %w`, err)
	}
	defer rows.Close()

	orders := []model.Order{}
	for rows.Next() {
		currOrder := model.Order{}
		rows.Scan(&currOrder.UserID, &currOrder.Number, &currOrder.Status, &currOrder.Accrual, &currOrder.UploadedAt)
		orders = append(orders, currOrder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf(`pg: %w`, err)
	}

	return orders, nil
}

func (p *Pg) UpdateOrderStatuses(newOrderStatuses []model.Order) error {
	log.Debug().Msg("Pg.UpdateOrderStatuses START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.UpdateOrderStatuses END")
		} else {
			log.Debug().Msg("Pg.UpdateOrderStatuses END")
		}
	}()

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, order := range newOrderStatuses {
		if _, err = tx.Stmt(p.ordersStmts.stmtUpdateOrderStatus).Exec(order.Status, order.Accrual, order.Number); err != nil {
			return err
		}
		if _, err = tx.Stmt(p.balanceStmts.stmtIncreaseBalance).Exec(order.UserID, order.Accrual); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
