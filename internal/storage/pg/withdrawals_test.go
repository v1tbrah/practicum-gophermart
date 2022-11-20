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
	dberr "practicum-gophermart/internal/storage/errors"
)

func TestPg_AddWithdrawal(t *testing.T) {
	testPg := Pg{}
	testPg.withdrawalsStmts = &withdrawalsStmts{}
	testPg.balanceStmts = &balanceStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryAddWithdrawal)
	if testPg.withdrawalsStmts.stmtAddWithdrawal, err = testPg.db.PrepareContext(context.Background(), queryAddWithdrawal); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	mock.ExpectPrepare(queryReduceBalance)
	if testPg.balanceStmts.stmtReduceBalance, err = testPg.db.PrepareContext(context.Background(), queryReduceBalance); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func(userID int64, withdraw model.Withdraw)
		userID       int64
		withdraw     model.Withdraw
		err          string
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(userID int64, withdraw model.Withdraw) {
				mock.ExpectBegin()
				mock.ExpectExec(queryAddWithdrawal).
					WithArgs(userID, withdraw.Order, withdraw.Sum, withdraw.ProcessedAt).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery(queryReduceBalance).
					WithArgs(userID, withdraw.Sum).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow("1"))
				mock.ExpectCommit()
			},
			userID: 1,
			withdraw: model.Withdraw{
				Order:       "123",
				Sum:         1.1,
				ProcessedAt: time.Now(),
			},
		},
		{
			name: "err on begin tx",
			mockBehavior: func(userID int64, withdraw model.Withdraw) {
				mock.ExpectBegin().WillReturnError(errors.New("unexpected err"))
			},
			userID: 1,
			withdraw: model.Withdraw{
				Order:       "123",
				Sum:         1.1,
				ProcessedAt: time.Now(),
			},
			err:     "unexpected err",
			wantErr: true,
		},
		{
			name: "unexpected err on adding withdraw",
			mockBehavior: func(userID int64, withdraw model.Withdraw) {
				mock.ExpectBegin()
				mock.ExpectExec(queryAddWithdrawal).
					WithArgs(userID, withdraw.Order, withdraw.Sum, withdraw.ProcessedAt).
					WillReturnError(errors.New("unexpected error"))
				mock.ExpectRollback()
			},
			userID: 1,
			withdraw: model.Withdraw{
				Order:       "123",
				Sum:         1.1,
				ProcessedAt: time.Now(),
			},
			err:     "unexpected error",
			wantErr: true,
		},
		{
			name: "unexpected err on reducing balance",
			mockBehavior: func(userID int64, withdraw model.Withdraw) {
				mock.ExpectBegin()
				mock.ExpectExec(queryAddWithdrawal).
					WithArgs(userID, withdraw.Order, withdraw.Sum, withdraw.ProcessedAt).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery(queryReduceBalance).
					WithArgs(userID, withdraw.Sum).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow("-1"))
				mock.ExpectRollback()
			},
			userID: 1,
			withdraw: model.Withdraw{
				Order:       "123",
				Sum:         1.1,
				ProcessedAt: time.Now(),
			},
			err:     dberr.ErrNegativeBalance.Error(),
			wantErr: true,
		},
		{
			name: "err negative balance",
			mockBehavior: func(userID int64, withdraw model.Withdraw) {
				mock.ExpectBegin()
				mock.ExpectExec(queryAddWithdrawal).
					WithArgs(userID, withdraw.Order, withdraw.Sum, withdraw.ProcessedAt).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery(queryReduceBalance).
					WithArgs(userID, withdraw.Sum).
					WillReturnError(errors.New("unexpected error"))
				mock.ExpectRollback()
			},
			userID: 1,
			withdraw: model.Withdraw{
				Order:       "123",
				Sum:         1.1,
				ProcessedAt: time.Now(),
			},
			err:     "unexpected error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.userID, tt.withdraw)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			err := testPg.AddWithdrawal(gCtx, tt.userID, tt.withdraw)
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

func TestPg_GetWithdrawals(t *testing.T) {
	testPg := Pg{}
	testPg.withdrawalsStmts = &withdrawalsStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryGetWithdrawals)
	if testPg.withdrawalsStmts.stmtGetWithdrawals, err = testPg.db.PrepareContext(context.Background(), queryGetWithdrawals); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func(userID int64)
		userID       int64
		expected     []model.Withdraw
		err          error
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(userID int64) {
				mock.ExpectQuery(queryGetWithdrawals).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
						AddRow("123", 123.0, time.Unix(1, 1)).
						AddRow("321", 321.0, time.Unix(2, 3)))

			},
			userID: 1,
			expected: []model.Withdraw{
				{
					Order:       "123",
					Sum:         123.0,
					ProcessedAt: time.Unix(1, 1),
				},
				{
					Order:       "321",
					Sum:         321.0,
					ProcessedAt: time.Unix(2, 3),
				},
			},
		},
		{
			name: "unexpected error",
			mockBehavior: func(userID int64) {
				mock.ExpectQuery(queryGetWithdrawals).
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
			withdraws, err := testPg.GetWithdrawals(gCtx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, withdraws)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
