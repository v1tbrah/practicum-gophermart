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

func TestAPI_setOrderHandler(t *testing.T) {
	tests := []struct {
		name         string
		payload      string
		mockApp      *mocks.Application
		authorized   bool
		expectedCode int
	}{
		{
			name:    "OK",
			payload: "12345678903",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("AddOrder", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.Order")).
					Return(nil).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusAccepted,
		},
		{
			name:         "unauthorized",
			authorized:   false,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid number",
			payload:      "invalid_number",
			authorized:   true,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "invalid number 2",
			payload:      "1",
			authorized:   true,
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:    "order was uploaded by another user",
			payload: "12345678903",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("AddOrder", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.Order")).
					Return(app.ErrOrderWasUploadedByAnotherUser).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusConflict,
		},
		{
			name:    "order was uploaded by current user",
			payload: "12345678903",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("AddOrder", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.Order")).
					Return(app.ErrOrderWasUploadedByCurrentUser).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusOK,
		},
		{
			name:    "unexpected err on adding order",
			payload: "12345678903",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("AddOrder", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*model.Order")).
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
			req := httptest.NewRequest(http.MethodPost, "/setOrderMockEndpoint", b)

			testCtx.Request = req

			testAPI.setOrderHandler(testCtx)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}

func TestAPI_ordersHandler(t *testing.T) {
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
				testApp.On("GetOrdersByUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64")).
					Return([]model.Order{{UserID: 1}}, nil).
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
			name: "unexpected err",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetOrdersByUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64")).
					Return(nil, errors.New("unexpected error")).
					Once()
				return &testApp
			}(),
			authorized:   true,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "no orders",
			mockApp: func() *mocks.Application {
				testApp := mocks.Application{}
				testApp.On("GetOrdersByUser", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("int64")).
					Return(nil, nil).
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

			testAPI.ordersHandler(testCtx)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
