package main

const (
	deliveryRestTemplate = `// {{.Header}}

package resthandler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
	"{{.LibraryName}}/wrapper"
)

// RestHandler handler
type RestHandler struct {
	mw        interfaces.Middleware
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewRestHandler create new rest handler
func NewRestHandler(uc usecase.Usecase, deps dependency.Dependency) *RestHandler {
	return &RestHandler{
		uc: uc, mw: deps.GetMiddleware(), validator: deps.GetValidator(),
	}
}

// Mount handler with root "/"
// handling version in here
func (h *RestHandler) Mount(root *echo.Group) {
	v1Root := root.Group(candihelper.V1)

	{{camel .ModuleName}} := v1Root.Group("/{{kebab .ModuleName}}", echo.WrapMiddleware(h.mw.HTTPBearerAuth))
	{{camel .ModuleName}}.GET("", h.getAll{{upper (camel .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("getAll{{upper (camel .ModuleName)}}")))
	{{camel .ModuleName}}.GET("/:id", h.getDetail{{upper (camel .ModuleName)}}ByID, echo.WrapMiddleware(h.mw.HTTPPermissionACL("getDetail{{upper (camel .ModuleName)}}")))
	{{camel .ModuleName}}.POST("", h.create{{upper (camel .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("create{{upper (camel .ModuleName)}}")))
	{{camel .ModuleName}}.PUT("/:id", h.update{{upper (camel .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("update{{upper (camel .ModuleName)}}")))
	{{camel .ModuleName}}.DELETE("/:id", h.delete{{upper (camel .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("delete{{upper (camel .ModuleName)}}")))
}

func (h *RestHandler) getAll{{upper (camel .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{upper (camel .ModuleName)}}DeliveryREST:GetAll{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using HTTPBearerAuth in middleware for this handler

	var filter domain.Filter{{upper (camel .ModuleName)}}
	if err := candihelper.ParseFromQueryParam(c.Request().URL.Query(), &filter); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed parse filter", err).JSON(c.Response())
	}

	if err := h.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/get_all", filter); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed validate filter", err).JSON(c.Response())
	}

	data, meta, err := h.uc.{{upper (camel .ModuleName)}}().GetAll{{upper (camel .ModuleName)}}(ctx, &filter)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	message := "Success, with your user id (" + tokenClaim.Subject + ") and role (" + tokenClaim.Role + ")"
	return wrapper.NewHTTPResponse(http.StatusOK, message, meta, data).JSON(c.Response())
}

func (h *RestHandler) getDetail{{upper (camel .ModuleName)}}ByID(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{upper (camel .ModuleName)}}DeliveryREST:GetDetail{{upper (camel .ModuleName)}}ByID")
	defer trace.Finish()

	data, err := h.uc.{{upper (camel .ModuleName)}}().GetDetail{{upper (camel .ModuleName)}}(ctx, c.Param("id"))
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success", data).JSON(c.Response())
}

func (h *RestHandler) create{{upper (camel .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{upper (camel .ModuleName)}}DeliveryREST:Create{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	body, _ := io.ReadAll(c.Request().Body)
	if err := h.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/save", body); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed validate payload", err).JSON(c.Response())
	}

	var payload domain.Request{{upper (camel .ModuleName)}}
	if err := json.Unmarshal(body, &payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	err := h.uc.{{upper (camel .ModuleName)}}().Create{{upper (camel .ModuleName)}}(ctx, &payload)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}

func (h *RestHandler) update{{upper (camel .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{upper (camel .ModuleName)}}DeliveryREST:Update{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	body, _ := io.ReadAll(c.Request().Body)
	if err := h.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/save", body); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed validate payload", err).JSON(c.Response())
	}

	var payload domain.Request{{upper (camel .ModuleName)}}
	if err := json.Unmarshal(body, &payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	payload.ID = c.Param("id")
	err := h.uc.{{upper (camel .ModuleName)}}().Update{{upper (camel .ModuleName)}}(ctx, &payload)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}

func (h *RestHandler) delete{{upper (camel .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{upper (camel .ModuleName)}}DeliveryREST:Delete{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	if err := h.uc.{{upper (camel .ModuleName)}}().Delete{{upper (camel .ModuleName)}}(ctx, c.Param("id")); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}
`

	deliveryRestTestTemplate = `// {{.Header}}

package resthandler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	mockusecase "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/usecase"
	mocksharedusecase "{{$.PackagePrefix}}/pkg/mocks/shared/usecase"

	"{{.LibraryName}}/candishared"
	mockdeps "{{.LibraryName}}/mocks/codebase/factory/dependency"
	mockinterfaces "{{.LibraryName}}/mocks/codebase/interfaces"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testCase struct {
	name, reqBody                       string
	wantValidateError, wantUsecaseError error
	wantRespCode                        int
}

var (
	errFoo = errors.New("Something error")
)

func TestNewRestHandler(t *testing.T) {
	mockMiddleware := &mockinterfaces.Middleware{}
	mockMiddleware.On("HTTPPermissionACL", mock.Anything).Return(func(http.Handler) http.Handler { return nil })
	mockValidator := &mockinterfaces.Validator{}

	mockDeps := &mockdeps.Dependency{}
	mockDeps.On("GetMiddleware").Return(mockMiddleware)
	mockDeps.On("GetValidator").Return(mockValidator)

	handler := NewRestHandler(nil, mockDeps)
	assert.NotNil(t, handler)

	e := echo.New()
	handler.Mount(e.Group("/"))
}

func TestRestHandler_getAll{{upper (camel .ModuleName)}}(t *testing.T) {
	tests := []testCase{
		{
			name: "Testcase #1: Positive", wantUsecaseError: nil, wantRespCode: 200,
		},
		{
			name: "Testcase #2: Negative", reqBody: "?page=str", wantUsecaseError: errFoo, wantRespCode: 400,
		},
		{
			name: "Testcase #3: Negative", wantUsecaseError: errFoo, wantRespCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			{{camel .ModuleName}}Usecase := &mockusecase.{{upper (camel .ModuleName)}}Usecase{}
			{{camel .ModuleName}}Usecase.On("GetAll{{upper (camel .ModuleName)}}", mock.Anything, mock.Anything).Return(
				[]domain.Response{{upper (camel .ModuleName)}}{}, candishared.Meta{}, tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{upper (camel .ModuleName)}}").Return({{camel .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodGet, "/"+tt.reqBody, strings.NewReader(tt.reqBody))
			req = req.WithContext(candishared.SetToContext(req.Context(), candishared.ContextKeyTokenClaim, &candishared.TokenClaim{}))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			err := handler.getAll{{upper (camel .ModuleName)}}(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}

func TestRestHandler_getDetail{{upper (camel .ModuleName)}}ByID(t *testing.T) {
	tests := []testCase{
		{
			name: "Testcase #1: Positive", wantUsecaseError: nil, wantRespCode: 200,
		},
		{
			name: "Testcase #2: Negative", wantUsecaseError: errFoo, wantRespCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			{{camel .ModuleName}}Usecase := &mockusecase.{{upper (camel .ModuleName)}}Usecase{}
			{{camel .ModuleName}}Usecase.On("GetDetail{{upper (camel .ModuleName)}}", mock.Anything, mock.Anything).Return(domain.Response{{upper (camel .ModuleName)}}{}, tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{upper (camel .ModuleName)}}").Return({{camel .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.reqBody))
			req = req.WithContext(candishared.SetToContext(req.Context(), candishared.ContextKeyTokenClaim, &candishared.TokenClaim{}))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			err := handler.getDetail{{upper (camel .ModuleName)}}ByID(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}

func TestRestHandler_create{{upper (camel .ModuleName)}}(t *testing.T) {
	tests := []testCase{
		{
			name: "Testcase #1: Positive", reqBody: ` + "`" + `{"email": "test@test.com"}` + "`" + `, wantUsecaseError: nil, wantRespCode: 200,
		},
		{
			name: "Testcase #2: Negative", reqBody: ` + "`" + `{"email": test@test.com}` + "`" + `, wantUsecaseError: nil, wantRespCode: 400,
		},
		{
			name: "Testcase #3: Negative", reqBody: ` + "`" + `{"email": "test@test.com"}` + "`" + `, wantUsecaseError: errFoo, wantRespCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			{{camel .ModuleName}}Usecase := &mockusecase.{{upper (camel .ModuleName)}}Usecase{}
			{{camel .ModuleName}}Usecase.On("Create{{upper (camel .ModuleName)}}", mock.Anything, mock.Anything).Return(tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{upper (camel .ModuleName)}}").Return({{camel .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.reqBody))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			echoContext.SetParamNames("id")
			echoContext.SetParamValues("001")
			err := handler.create{{upper (camel .ModuleName)}}(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}

func TestRestHandler_update{{upper (camel .ModuleName)}}(t *testing.T) {
	tests := []testCase{
		{
			name: "Testcase #1: Positive", reqBody: ` + "`" + `{"email": "test@test.com"}` + "`" + `, wantUsecaseError: nil, wantRespCode: 200,
		},
		{
			name: "Testcase #2: Negative", reqBody: ` + "`" + `{"email": test@test.com}` + "`" + `, wantValidateError: errFoo, wantRespCode: 400,
		},
		{
			name: "Testcase #3: Negative", reqBody: ` + "`" + `{"email": test@test.com}` + "`" + `, wantUsecaseError: nil, wantRespCode: 400,
		},
		{
			name: "Testcase #4: Negative", reqBody: ` + "`" + `{"email": "test@test.com"}` + "`" + `, wantUsecaseError: errFoo, wantRespCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			{{camel .ModuleName}}Usecase := &mockusecase.{{upper (camel .ModuleName)}}Usecase{}
			{{camel .ModuleName}}Usecase.On("Update{{upper (camel .ModuleName)}}", mock.Anything, mock.Anything, mock.Anything).Return(tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{upper (camel .ModuleName)}}").Return({{camel .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.reqBody))
			req = req.WithContext(candishared.SetToContext(req.Context(), candishared.ContextKeyTokenClaim, &candishared.TokenClaim{}))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			err := handler.update{{upper (camel .ModuleName)}}(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}

func TestRestHandler_delete{{upper (camel .ModuleName)}}(t *testing.T) {
	tests := []testCase{
		{
			name: "Testcase #1: Positive", wantUsecaseError: nil, wantRespCode: 200,
		},
		{
			name: "Testcase #2: Negative", wantUsecaseError: errFoo, wantRespCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			{{camel .ModuleName}}Usecase := &mockusecase.{{upper (camel .ModuleName)}}Usecase{}
			{{camel .ModuleName}}Usecase.On("Delete{{upper (camel .ModuleName)}}", mock.Anything, mock.Anything).Return(tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{upper (camel .ModuleName)}}").Return({{camel .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.reqBody))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			err := handler.delete{{upper (camel .ModuleName)}}(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}
`
)
