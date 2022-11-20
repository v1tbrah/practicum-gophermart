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

func TestAPI_balanceHandler(t *testing.T) {
	tests := []struct {
		name         string
		mockApp      *mocks.Application
		authorized   bool
		expectedCode int
	}{
		{
			name: "OK",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetBalance", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64")).
					Return(1.11, 11.33, nil).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "unauthorized",
			authorized:   false,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "unexpected error",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetBalance", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64")).
					Return(-1.0, -1.0, errors.New("unexpected error")).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAPI := API{}
			testAPI.app = tt.mockApp
			testAPI.authMngr = newAuthMngr()

			rec := httptest.NewRecorder()

			testCtx, _ := gin.CreateTestContext(rec)
			if tt.authorized {
				testCtx.Set("id", int64(1))
			}

			testAPI.balanceHandler(testCtx)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}

func TestAPI_withdrawPointsHandler(t *testing.T) {
	tests := []struct {
		name         string
		payload      string
		mockApp      *mocks.Application
		authorized   bool
		expectedCode int
	}{
		{
			name:    "OK",
			payload: "{\"order\": \"12345678903\", \"sum\": 751}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("WithdrawFromBalance", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64"), mock.AnythingOfType("model.Withdraw")).
					Return(nil).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "unauthorized",
			authorized:   false,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:    "bad payload",
			payload: "{\"order\": \"12345678903\", \"sum\": \"\"\"\"\"\"\"\"\"}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("WithdrawFromBalance", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64"), mock.AnythingOfType("model.Withdraw")).
					Return(nil).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid order number",
			payload:      "{\"order\": \"1\", \"sum\": 755}",
			authorized:   true,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:    "err insufficient funds",
			payload: "{\"order\": \"12345678903\", \"sum\": 751}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("WithdrawFromBalance", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64"), mock.AnythingOfType("model.Withdraw")).
					Return(app.ErrInsufficientFunds).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusPaymentRequired,
		},
		{
			name:    "unexpected err on adding withdraw",
			payload: "{\"order\": \"12345678903\", \"sum\": 751}",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("WithdrawFromBalance", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64"), mock.AnythingOfType("model.Withdraw")).
					Return(errors.New("unexpected error")).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAPI := API{}
			testAPI.app = tt.mockApp
			testAPI.authMngr = newAuthMngr()

			rec := httptest.NewRecorder()

			testCtx, _ := gin.CreateTestContext(rec)
			if tt.authorized {
				testCtx.Set("id", int64(1))
			}

			b := &bytes.Buffer{}
			b.WriteString(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/withdrawPointsMockEndpoint", b)

			testCtx.Request = req

			testAPI.withdrawPointsHandler(testCtx)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}

func TestAPI_withdrawnPointsHandler(t *testing.T) {
	tests := []struct {
		name         string
		mockApp      *mocks.Application
		authorized   bool
		expectedCode int
	}{
		{
			name: "OK",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetWithdrawals", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64")).
					Return([]model.Withdraw{{Order: "123"}, {Order: "321"}}, nil).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "unauthorized",
			authorized:   false,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "unexpected error",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetWithdrawals", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64")).
					Return(nil, errors.New("unexpected error")).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "haven't withdrawals",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetWithdrawals", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64")).
					Return([]model.Withdraw{}, nil).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusNoContent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAPI := API{}
			testAPI.app = tt.mockApp
			testAPI.authMngr = newAuthMngr()

			rec := httptest.NewRecorder()

			testCtx, _ := gin.CreateTestContext(rec)
			if tt.authorized {
				testCtx.Set("id", int64(1))
			}

			testAPI.withdrawnPointsHandler(testCtx)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
