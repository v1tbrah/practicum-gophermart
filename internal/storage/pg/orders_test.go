package pg

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"practicum-gophermart/internal/model"
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
			name: "OK",
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
