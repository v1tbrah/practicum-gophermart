package pg

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPg_GetBalance(t *testing.T) {
	testPg := Pg{}
	testPg.balanceStmts = &balanceStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryGetBalance)
	if testPg.balanceStmts.stmtGetBalance, err = testPg.db.PrepareContext(context.Background(), queryGetBalance); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	testPg.db = db

	tests := []struct {
		name              string
		mockBehavior      func(userID int64)
		userID            int64
		expectedBalance   float64
		expectedWithdrawn float64
		err               string
		wantErr           bool
	}{
		{
			name: "OK",
			mockBehavior: func(userID int64) {
				mock.ExpectQuery(queryGetBalance).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"current", "withdrawn"}).AddRow(1.33, 11.44))
			},
			userID:            1,
			expectedBalance:   1.33,
			expectedWithdrawn: 11.44,
		},
		{
			name: "unexpected error",
			mockBehavior: func(userID int64) {
				mock.ExpectQuery(queryGetBalance).
					WithArgs(userID).
					WillReturnError(errors.New("unexpected error"))
			},
			userID:  1,
			err:     "unexpected error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.userID)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			balance, withdrawn, err := testPg.GetBalance(gCtx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
				assert.Equal(t, tt.expectedWithdrawn, withdrawn)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
