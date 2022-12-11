package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"practicum-gophermart/internal/api/mocks"
	"practicum-gophermart/internal/app"
	"practicum-gophermart/internal/model"
)

func TestAPI_signUpHandler(t *testing.T) {

	tests := []struct {
		mockApp      *mocks.Application
		name         string
		payload      string
		expectedCode int
	}{
		{
			name:    "valid",
			payload: "{\"login\": \"validLogin\", \"password\": \"validPassword\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("CreateUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.User")).
					Return(int64(1), nil).
					Once()
				testApp.On("NewRefreshSession", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.RefreshSession")).
					Return(nil).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusOK,
		},
		{
			name:    "login already exists",
			payload: "{\"login\": \"usedLogin\", \"password\": \"validPassword\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("CreateUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.User")).
					Return(int64(0), app.ErrUserAlreadyExists).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusConflict,
		},
		{
			name:    "unexpected err on creating user",
			payload: "{\"login\": \"validLogin\", \"password\": \"validPassword\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("CreateUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.User")).
					Return(int64(0), errors.New("unexpected error")).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid body",
			payload:      "{\"login\": \"validLogin\", \"password",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:    "err on new refresh session",
			payload: "{\"login\": \"validLogin\", \"password\": \"validPassword\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("CreateUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.User")).
					Return(int64(1), nil).
					Once()
				testApp.On("NewRefreshSession", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.RefreshSession")).
					Return(errors.New("unexpected error")).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAPI := API{}
			testAPI.app = tt.mockApp
			testAPI.authMngr = newAuthMngr()

			rec := httptest.NewRecorder()

			router := gin.New()
			router.POST("/signUpMockEndpoint", testAPI.signUpHandler)

			b := &bytes.Buffer{}
			b.WriteString(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/signUpMockEndpoint", b)

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}

func TestAPI_signInHandler(t *testing.T) {

	tests := []struct {
		mockApp      *mocks.Application
		name         string
		payload      string
		expectedCode int
	}{
		{
			name:    "valid",
			payload: "{\"login\": \"validLogin\", \"password\": \"validPassword\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(&model.User{Login: "validLogin"}, nil).
					Once()
				testApp.On("NewRefreshSession", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.RefreshSession")).
					Return(nil).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusOK,
		},
		{
			name:    "invalid login or password",
			payload: "{\"login\": \"invalidLogin\", \"password\": \"invalidPassword\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(nil, app.ErrInvalidLoginOrPassword).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:    "unexpected err on getting user",
			payload: "{\"login\": \"validLogin\", \"password\": \"validPassword\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(nil, errors.New("unexpected error")).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:    "invalid body",
			payload: "{\"login\": \"validLogin\", \"password",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(&model.User{Login: "validLogin"}, nil).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:    "\"err on new refresh session\"",
			payload: "{\"login\": \"validLogin\", \"password\": \"validPassword\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(&model.User{Login: "validLogin"}, nil).
					Once()
				testApp.On("NewRefreshSession", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.RefreshSession")).
					Return(errors.New("unexpected error")).
					Once()
				return &testApp
			}(),
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAPI := API{}
			testAPI.app = tt.mockApp
			testAPI.authMngr = newAuthMngr()

			rec := httptest.NewRecorder()

			router := gin.New()
			router.POST("/signInMockEndpoint", testAPI.signInHandler)

			b := &bytes.Buffer{}
			b.WriteString(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/signInMockEndpoint", b)

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
