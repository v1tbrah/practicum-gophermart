package pg

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

func TestPg_AddOrder(t *testing.T) {
	testPg := Pg{}
	testPg.ordersStmts = &ordersStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryAddOrder)
	if testPg.ordersStmts.stmtAddOrder, err = testPg.db.PrepareContext(context.Background(), queryAddOrder); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func(order *model.Order)
		order        model.Order
		err          error
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(order *model.Order) {
				mock.ExpectExec(queryAddOrder).
					WithArgs(order.UserID, order.Number, order.Status, order.Accrual, order.UploadedAt).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			order: model.Order{
				UserID:     1,
				Number:     "123456789",
				Status:     model.OrderStatusNew.String(),
				Accrual:    0,
				UploadedAt: time.Now(),
			},
		},
		{
			name: "unexpected error",
			mockBehavior: func(order *model.Order) {
				mock.ExpectExec(queryAddOrder).
					WithArgs(order.UserID, order.Number, order.Status, order.Accrual, order.UploadedAt).
					WillReturnError(errors.New("unexpected error"))
			},
			order: model.Order{
				UserID:     1,
				Number:     "123456789",
				Status:     model.OrderStatusNew.String(),
				Accrual:    0,
				UploadedAt: time.Now(),
			},
			err:     errors.New("unexpected error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(&tt.order)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			err := testPg.AddOrder(gCtx, &tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPg_GetOrdersByUser(t *testing.T) {
	testPg := Pg{}
	testPg.ordersStmts = &ordersStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryGetOrdersByUser)
	if testPg.ordersStmts.stmtGetUserOrders, err = testPg.db.PrepareContext(context.Background(), queryGetOrdersByUser); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func(userID int64)
		userID       int64
		expected     []model.Order
		err          error
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(userID int64) {
				mock.ExpectQuery(queryGetOrdersByUser).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"user_id", "number", "status", "accrual", "uploaded_at"}).
						AddRow(1, "123", "NEW", 0, time.Unix(1, 1)))
			},
			userID: 1,
			expected: []model.Order{
				{
					UserID:     1,
					Number:     "123",
					Status:     "NEW",
					Accrual:    0,
					UploadedAt: time.Unix(1, 1),
				},
			},
		},
		{
			name: "unexpected error",
			mockBehavior: func(userID int64) {
				mock.ExpectQuery(queryGetOrdersByUser).
					WithArgs(userID).
					WillReturnError(errors.New("unexpected error"))
			},
			userID:  1,
			wantErr: true,
			err:     errors.New("unexpected error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.userID)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			orders, err := testPg.GetOrdersByUser(gCtx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, orders)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPg_GetOrder(t *testing.T) {
	testPg := Pg{}
	testPg.ordersStmts = &ordersStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryGetOrder)
	if testPg.ordersStmts.stmtGetOrder, err = testPg.db.PrepareContext(context.Background(), queryGetOrder); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func(number string)
		number       string
		expected     *model.Order
		err          string
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(number string) {
				mock.ExpectQuery(queryGetOrder).
					WithArgs(number).
					WillReturnRows(sqlmock.NewRows([]string{"user_id", "number", "status", "accrual", "uploaded_at"}).
						AddRow(1, "123", "NEW", 0, time.Unix(1, 1)))
			},
			number: "123",
			expected: &model.Order{
				UserID:     1,
				Number:     "123",
				Status:     "NEW",
				Accrual:    0,
				UploadedAt: time.Unix(1, 1),
			},
		},
		{
			name: "order is not exists",
			mockBehavior: func(number string) {
				mock.ExpectQuery(queryGetOrder).
					WithArgs(number).
					WillReturnError(sql.ErrNoRows)
			},
			number:  "123",
			wantErr: true,
			err:     dberr.ErrOrderIsNotExists.Error(),
		},
		{
			name: "unexpected error",
			mockBehavior: func(number string) {
				mock.ExpectQuery(queryGetOrder).
					WithArgs(number).
					WillReturnError(errors.New("unexpected error"))
			},
			number:  "123",
			wantErr: true,
			err:     "unexpected error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.number)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			orders, err := testPg.GetOrder(gCtx, tt.number)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, orders)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPg_GetOrderNumbersByStatuses(t *testing.T) {
	testPg := Pg{}
	testPg.ordersStmts = &ordersStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryGetOrderNumbersByStatuses)
	if testPg.ordersStmts.stmtGetOrderNumbersByStatuses, err = testPg.db.PrepareContext(context.Background(), queryGetOrderNumbersByStatuses); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func()
		expected     []string
		err          string
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func() {
				mock.ExpectQuery(queryGetOrderNumbersByStatuses).
					WithArgs(pq.Array([]string{"NEW", "PROCESSING"})).
					WillReturnRows(sqlmock.NewRows([]string{"number"}).AddRow("123").AddRow("321"))
			},
			expected: []string{"123", "321"},
		},
		{
			name: "unexpected error",
			mockBehavior: func() {
				mock.ExpectQuery(queryGetOrderNumbersByStatuses).
					WithArgs(pq.Array([]string{"NEW", "PROCESSING"})).
					WillReturnError(errors.New("unexpected error"))
			},
			err:     "unexpected error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()
			orders, err := testPg.GetOrderNumbersByStatuses([]string{"NEW", "PROCESSING"})
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, orders)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPg_UpdateOrderStatuses(t *testing.T) {
	testPg := Pg{}
	testPg.ordersStmts = &ordersStmts{}
	testPg.balanceStmts = &balanceStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryUpdateOrderStatus)
	if testPg.ordersStmts.stmtUpdateOrderStatus, err = testPg.db.PrepareContext(context.Background(), queryUpdateOrderStatus); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	mock.ExpectPrepare(queryIncreaseBalance)
	if testPg.balanceStmts.stmtIncreaseBalance, err = testPg.db.PrepareContext(context.Background(), queryIncreaseBalance); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	testPg.db = db

	tests := []struct {
		name                            string
		mockBehavior                    func(newOrderStatuses []model.Order, wantErrOnUpdatingThisOrderIndex int)
		newOrderStatuses                []model.Order
		wantErrOnUpdatingThisOrderIndex int
		err                             string
		wantErr                         bool
	}{
		{
			name: "OK",
			mockBehavior: func(newOrderStatuses []model.Order, wantErrOnUpdatingThisOrderIndex int) {
				mock.ExpectBegin()
				for _, order := range newOrderStatuses {
					mock.ExpectExec(queryUpdateOrderStatus).
						WithArgs(order.Status, order.Accrual, order.Number).
						WillReturnResult(sqlmock.NewResult(0, 0))
					mock.ExpectExec(queryIncreaseBalance).
						WithArgs(order.UserID, order.Accrual).
						WillReturnResult(sqlmock.NewResult(0, 0))
				}
				mock.ExpectCommit()
			},
			newOrderStatuses: []model.Order{
				{Status: "PROCESSED", Accrual: 22.33, Number: "123"},
				{Status: "PROCESSING", Accrual: 33.22, Number: "321"},
			},
		},
		{
			name: "err on begin tx",
			mockBehavior: func(newOrderStatuses []model.Order, wantErrOnUpdatingThisOrderIndex int) {
				mock.ExpectBegin().WillReturnError(errors.New("unexpected err"))
			},
			newOrderStatuses: []model.Order{
				{Status: "PROCESSED", Accrual: 22.33, Number: "123"},
				{Status: "PROCESSING", Accrual: 33.22, Number: "321"},
			},
			err:     "unexpected err",
			wantErr: true,
		},
		{
			name: "err on updating order on index [N]",
			mockBehavior: func(newOrderStatuses []model.Order, wantErrOnUpdatingThisOrderIndex int) {
				mock.ExpectBegin()
				for i, order := range newOrderStatuses {
					if i == wantErrOnUpdatingThisOrderIndex {
						mock.ExpectExec(queryUpdateOrderStatus).
							WithArgs(order.Status, order.Accrual, order.Number).
							WillReturnError(errors.New("unexpected error"))
						mock.ExpectRollback()
						return
					}
					mock.ExpectExec(queryUpdateOrderStatus).
						WithArgs(order.Status, order.Accrual, order.Number).
						WillReturnResult(sqlmock.NewResult(0, 0))
					mock.ExpectExec(queryIncreaseBalance).
						WithArgs(order.UserID, order.Accrual).
						WillReturnResult(sqlmock.NewResult(0, 0))
				}
				mock.ExpectCommit()
			},
			wantErrOnUpdatingThisOrderIndex: 1,
			newOrderStatuses: []model.Order{
				{Status: "PROCESSED", Accrual: 22.33, Number: "123"},
				{Status: "PROCESSING", Accrual: 33.22, Number: "321"},
			},
			err:     "unexpected error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErrOnUpdatingThisOrderIndex == 0 {
				tt.mockBehavior(tt.newOrderStatuses, -1)
			} else {
				tt.mockBehavior(tt.newOrderStatuses, tt.wantErrOnUpdatingThisOrderIndex)
			}

			err := testPg.UpdateOrderStatuses(tt.newOrderStatuses)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
