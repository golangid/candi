package main

const (
	deliveryRestTemplate = `// {{.Header}}

package resthandler

import (
	"net/http"

	"github.com/labstack/echo"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/model"
	"{{$.PackagePrefix}}/pkg/shared/sharedmodel"
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

	{{clean .ModuleName}} := v1Root.Group("/{{clean .ModuleName}}", echo.WrapMiddleware(h.mw.HTTPBearerAuth))
	{{clean .ModuleName}}.GET("", h.getAll{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.GET("/:id", h.getDetail{{clean (upper .ModuleName)}}ByID, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.POST("", h.create{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.PUT("/:id", h.update{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.DELETE("/:id", h.delete{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
}

func (h *RestHandler) getAll{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using HTTPBearerAuth in middleware for this handler

	var filter model.Filter{{clean (upper .ModuleName)}}
	if err := candihelper.ParseFromQueryParam(c.Request().URL.Query(), &filter); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	data, meta, err := h.uc.{{clean (upper .ModuleName)}}().GetAll{{clean (upper .ModuleName)}}(ctx, &filter)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	message := "Success, with your user id (" + tokenClaim.Subject + ") and role (" + tokenClaim.Role + ")"
	return wrapper.NewHTTPResponse(http.StatusOK, message, meta, data).JSON(c.Response())
}

func (h *RestHandler) getDetail{{clean (upper .ModuleName)}}ByID(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:GetDetail{{clean (upper .ModuleName)}}ByID")
	defer trace.Finish()

	data, err := h.uc.{{clean (upper .ModuleName)}}().GetDetail{{clean (upper .ModuleName)}}(ctx, c.Param("id"))
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success", data).JSON(c.Response())
}

func (h *RestHandler) create{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Create{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	var payload sharedmodel.{{clean (upper .ModuleName)}}
	if err := c.Bind(&payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	err := h.uc.{{clean (upper .ModuleName)}}().Create{{clean (upper .ModuleName)}}(ctx, &payload)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}

func (h *RestHandler) update{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Update{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	var payload sharedmodel.{{clean (upper .ModuleName)}}
	if err := c.Bind(&payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	err := h.uc.{{clean (upper .ModuleName)}}().Update{{clean (upper .ModuleName)}}(ctx, c.Param("id"), &payload)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}

func (h *RestHandler) delete{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Delete{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	if err := h.uc.{{clean (upper .ModuleName)}}().Delete{{clean (upper .ModuleName)}}(ctx, c.Param("id")); err != nil {
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

	mockusecase "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/usecase"
	mocksharedusecase "{{$.PackagePrefix}}/pkg/mocks/shared/usecase"
	"{{$.PackagePrefix}}/pkg/shared/sharedmodel"

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

func TestRestHandler_getAll{{clean (upper .ModuleName)}}(t *testing.T) {
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

			{{cleanPathModule .ModuleName}}Usecase := &mockusecase.{{clean (upper .ModuleName)}}Usecase{}
			{{cleanPathModule .ModuleName}}Usecase.On("GetAll{{clean (upper .ModuleName)}}", mock.Anything, mock.Anything).Return(
				[]sharedmodel.{{clean (upper .ModuleName)}}{}, candishared.Meta{}, tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{clean (upper .ModuleName)}}").Return({{cleanPathModule .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodGet, "/"+tt.reqBody, strings.NewReader(tt.reqBody))
			req = req.WithContext(candishared.SetToContext(req.Context(), candishared.ContextKeyTokenClaim, &candishared.TokenClaim{}))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			err := handler.getAll{{clean (upper .ModuleName)}}(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}

func TestRestHandler_getDetail{{clean (upper .ModuleName)}}ByID(t *testing.T) {
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

			{{cleanPathModule .ModuleName}}Usecase := &mockusecase.{{clean (upper .ModuleName)}}Usecase{}
			{{cleanPathModule .ModuleName}}Usecase.On("GetDetail{{clean (upper .ModuleName)}}", mock.Anything, mock.Anything).Return(sharedmodel.{{clean (upper .ModuleName)}}{}, tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{clean (upper .ModuleName)}}").Return({{cleanPathModule .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.reqBody))
			req = req.WithContext(candishared.SetToContext(req.Context(), candishared.ContextKeyTokenClaim, &candishared.TokenClaim{}))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			err := handler.getDetail{{clean (upper .ModuleName)}}ByID(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}

func TestRestHandler_create{{clean (upper .ModuleName)}}(t *testing.T) {
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

			{{cleanPathModule .ModuleName}}Usecase := &mockusecase.{{clean (upper .ModuleName)}}Usecase{}
			{{cleanPathModule .ModuleName}}Usecase.On("Create{{clean (upper .ModuleName)}}", mock.Anything, mock.Anything).Return(tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{clean (upper .ModuleName)}}").Return({{cleanPathModule .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.reqBody))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			echoContext.SetParamNames("id")
			echoContext.SetParamValues("001")
			err := handler.create{{clean (upper .ModuleName)}}(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}

func TestRestHandler_update{{clean (upper .ModuleName)}}(t *testing.T) {
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

			{{cleanPathModule .ModuleName}}Usecase := &mockusecase.{{clean (upper .ModuleName)}}Usecase{}
			{{cleanPathModule .ModuleName}}Usecase.On("Update{{clean (upper .ModuleName)}}", mock.Anything, mock.Anything, mock.Anything).Return(tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{clean (upper .ModuleName)}}").Return({{cleanPathModule .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.reqBody))
			req = req.WithContext(candishared.SetToContext(req.Context(), candishared.ContextKeyTokenClaim, &candishared.TokenClaim{}))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			err := handler.update{{clean (upper .ModuleName)}}(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}

func TestRestHandler_delete{{clean (upper .ModuleName)}}(t *testing.T) {
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

			{{cleanPathModule .ModuleName}}Usecase := &mockusecase.{{clean (upper .ModuleName)}}Usecase{}
			{{cleanPathModule .ModuleName}}Usecase.On("Delete{{clean (upper .ModuleName)}}", mock.Anything, mock.Anything).Return(tt.wantUsecaseError)
			mockValidator := &mockinterfaces.Validator{}
			mockValidator.On("ValidateDocument", mock.Anything, mock.Anything).Return(tt.wantValidateError)

			uc := &mocksharedusecase.Usecase{}
			uc.On("{{clean (upper .ModuleName)}}").Return({{cleanPathModule .ModuleName}}Usecase)

			handler := RestHandler{uc: uc, validator: mockValidator}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.reqBody))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			echoContext := echo.New().NewContext(req, res)
			err := handler.delete{{clean (upper .ModuleName)}}(echoContext)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRespCode, res.Code)
		})
	}
}
`
)
