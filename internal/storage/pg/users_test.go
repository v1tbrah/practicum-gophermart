package pg

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"practicum-gophermart/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	dberr "practicum-gophermart/internal/storage/errors"
)

func TestPg_AddUser(t *testing.T) {
	testPg := Pg{}
	testPg.usersStmts = &usersStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testPg.db = db
	mock.ExpectPrepare(queryAddUser)
	if testPg.usersStmts.stmtAddUser, err = testPg.db.PrepareContext(context.Background(), queryAddUser); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}

	testPg.db = db

	tests := []struct {
		name         string
		mockBehavior func(*model.User)
		user         model.User
		expectedID   int64
		err          error
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(user *model.User) {
				mock.ExpectQuery(queryAddUser).
					WithArgs(user.Login, user.Password).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
			},
			user: model.User{
				Login:    "testLogin",
				Password: "testPassword",
			},
			expectedID: 1,
		},
		{
			name: "login already exists",
			mockBehavior: func(user *model.User) {
				mock.ExpectQuery(queryAddUser).
					WithArgs(user.Login, user.Password).
					WillReturnError(&pgconn.PgError{
						Code:           pgerrcode.UniqueViolation,
						ConstraintName: "users_login_key"})
			},
			user: model.User{
				Login:    "testLogin",
				Password: "testPassword",
			},
			expectedID: 0,
			err:        dberr.ErrLoginAlreadyExists,
			wantErr:    true,
		},
		{
			name: "unexpected err on query",
			mockBehavior: func(user *model.User) {
				mock.ExpectQuery(queryAddUser).
					WithArgs(user.Login, user.Password).
					WillReturnError(errors.New("unexpected error"))
			},
			user: model.User{
				Login:    "testLogin",
				Password: "testPassword",
			},
			expectedID: 0,
			err:        errors.New("unexpected error"),
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(&tt.user)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			id, err := testPg.AddUser(gCtx, &tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPg_GetUser(t *testing.T) {
	testPg := Pg{}
	testPg.usersStmts = &usersStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	testPg.db = db

	mock.ExpectPrepare(queryGetUser)
	if testPg.usersStmts.stmtGetUser, err = testPg.db.PrepareContext(context.Background(), queryGetUser); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}

	tests := []struct {
		name         string
		mockBehavior func(login, pwd string)
		expected     *model.User
		login        string
		password     string
		err          error
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(login, pwd string) {
				mock.ExpectQuery(queryGetUser).
					WithArgs(login, pwd).
					WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password"}).
						AddRow(int64(1), login, pwd))
			},
			login:    "testLogin",
			password: "testPassword",
			expected: &model.User{ID: 1, Login: "testLogin", Password: "testPassword"},
		},
		{
			name: "user is not found",
			mockBehavior: func(login, pwd string) {
				mock.ExpectQuery(queryGetUser).
					WithArgs(login, pwd).
					WillReturnError(sql.ErrNoRows)
			},
			err:     dberr.ErrInvalidLoginOrPassword,
			wantErr: true,
		},
		{
			name: "unexpected err",
			mockBehavior: func(login, pwd string) {
				mock.ExpectQuery(queryGetUser).
					WithArgs(login, pwd).
					WillReturnError(errors.New("unexpected error"))
			},
			err:     errors.New("unexpected error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.login, tt.password)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			user, err := testPg.GetUser(gCtx, tt.login, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, user)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPg_GetUserPassword(t *testing.T) {
	testPg := Pg{}
	testPg.usersStmts = &usersStmts{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	testPg.db = db

	mock.ExpectPrepare(queryGetUserPassword)
	if testPg.usersStmts.stmtGetUserPassword, err = testPg.db.PrepareContext(context.Background(), queryGetUserPassword); err != nil {
		t.Fatalf("an error '%s' was not expected when preparing create user statement", err)
	}

	tests := []struct {
		name         string
		mockBehavior func(login string)
		expected     string
		login        string
		err          error
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(login string) {
				mock.ExpectQuery(queryGetUserPassword).
					WithArgs(login).
					WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow("testPassword"))
			},
			expected: "testPassword",
		},
		{
			name: "user is not found",
			mockBehavior: func(login string) {
				mock.ExpectQuery(queryGetUserPassword).
					WithArgs(login).
					WillReturnError(sql.ErrNoRows)
			},
			err:     dberr.ErrInvalidLoginOrPassword,
			wantErr: true,
		},
		{
			name: "unexpected err",
			mockBehavior: func(login string) {
				mock.ExpectQuery(queryGetUserPassword).
					WithArgs(login).
					WillReturnError(errors.New("unexpected error"))
			},
			err:     errors.New("unexpected error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.login)
			gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
			user, err := testPg.GetUserPassword(gCtx, tt.login)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, user)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
