package pg

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func Test_initTables(t *testing.T) {
	testPg := Pg{}
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func()
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func() {
				mock.ExpectBegin()
				mock.ExpectExec(queryCreateTypeOrderStatus).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(queryCreateTableUsers).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(queryCreateTableRefreshSessions).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(queryCreateTableOrders).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(queryCreateTableBalance).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(queryCreateTableWithdrawals).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()
			err = initTables(context.Background(), &testPg)
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
