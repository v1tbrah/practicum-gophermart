package pg

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

func TestPg_UpdateRefreshSession(t *testing.T) {
	testPg := Pg{}
	testPg.refreshSessionStmts = &refreshSessionStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryDeleteRefreshSessions)
	if testPg.refreshSessionStmts.stmtDeleteRefreshSession, err = testPg.db.PrepareContext(context.Background(), queryDeleteRefreshSessions); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}
	mock.ExpectPrepare(queryAddRefreshSession)
	if testPg.refreshSessionStmts.stmtAddRefreshSession, err = testPg.db.PrepareContext(context.Background(), queryAddRefreshSession); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}

	testPg.db = db

	tests := []struct {
		name           string
		mockBehavior   func(*model.RefreshSession)
		refreshSession *model.RefreshSession
		err            error
		wantErr        bool
	}{
		{
			name: "OK",
			mockBehavior: func(refreshSession *model.RefreshSession) {
				mock.ExpectBegin()
				mock.ExpectExec(queryDeleteRefreshSessions).
					WithArgs(&refreshSession.UserID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(queryAddRefreshSession).
					WithArgs(&refreshSession.UserID, &refreshSession.Token, &refreshSession.ExpiresIn).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			refreshSession: &model.RefreshSession{
				UserID:    1,
				Token:     "2",
				ExpiresIn: 3,
			},
		},
		{
			name: "err after begin",
			mockBehavior: func(refreshSession *model.RefreshSession) {
				mock.ExpectBegin().WillReturnError(errors.New("unexpected err"))

			},
			refreshSession: &model.RefreshSession{
				UserID:    1,
				Token:     "2",
				ExpiresIn: 3,
			},
			err:     errors.New("unexpected err"),
			wantErr: true,
		},
		{
			name: "err after delete sessions",
			mockBehavior: func(refreshSession *model.RefreshSession) {
				mock.ExpectBegin()
				mock.ExpectExec(queryDeleteRefreshSessions).
					WithArgs(&refreshSession.UserID).
					WillReturnError(errors.New("unexpected err"))
				mock.ExpectRollback()
			},
			refreshSession: &model.RefreshSession{
				UserID:    1,
				Token:     "2",
				ExpiresIn: 3,
			},
			err:     errors.New("unexpected err"),
			wantErr: true,
		},
		{
			name: "err after add session",
			mockBehavior: func(refreshSession *model.RefreshSession) {
				mock.ExpectBegin()
				mock.ExpectExec(queryDeleteRefreshSessions).
					WithArgs(&refreshSession.UserID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(queryAddRefreshSession).
					WithArgs(&refreshSession.UserID, &refreshSession.Token, &refreshSession.ExpiresIn).
					WillReturnError(errors.New("unexpected err"))
				mock.ExpectRollback()
			},
			refreshSession: &model.RefreshSession{
				UserID:    1,
				Token:     "2",
				ExpiresIn: 3,
			},
			err:     errors.New("unexpected err"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.refreshSession)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			err := testPg.UpdateRefreshSession(gCtx, tt.refreshSession)
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

func TestPg_GetRefreshSessionByToken(t *testing.T) {
	testPg := Pg{}
	testPg.refreshSessionStmts = &refreshSessionStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryGetRefreshSessionByToken)
	if testPg.refreshSessionStmts.stmtGetRefreshSessionByToken, err = testPg.db.PrepareContext(context.Background(), queryGetRefreshSessionByToken); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}

	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func(string)
		token        string
		expected     *model.RefreshSession
		err          error
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(token string) {
				mock.ExpectQuery(queryGetRefreshSessionByToken).
					WithArgs(token).
					WillReturnRows(sqlmock.NewRows([]string{"user_id", "expiresIn"}).AddRow(1, 2))
			},
			expected: &model.RefreshSession{
				UserID:    1,
				ExpiresIn: 2,
			},
		},
		{
			name: "err sql no rows",
			mockBehavior: func(token string) {
				mock.ExpectQuery(queryGetRefreshSessionByToken).
					WithArgs(token).
					WillReturnError(sql.ErrNoRows)
			},
			expected: &model.RefreshSession{
				UserID:    1,
				ExpiresIn: 2,
			},
			err:     dberr.ErrRefreshSessionIsNotExists,
			wantErr: true,
		},
		{
			name: "unexpected err",
			mockBehavior: func(token string) {
				mock.ExpectQuery(queryGetRefreshSessionByToken).
					WithArgs(token).
					WillReturnError(errors.New("unexpected err"))
			},
			expected: &model.RefreshSession{
				UserID:    1,
				ExpiresIn: 2,
			},
			err:     errors.New("unexpected err"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.token)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			refreshSession, err := testPg.GetRefreshSessionByToken(gCtx, tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, refreshSession)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
