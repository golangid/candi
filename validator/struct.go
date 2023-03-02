package validator

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	idLocales "github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	validatorEngine "github.com/go-playground/validator/v10"
	idTranslations "github.com/go-playground/validator/v10/translations/id"
	"github.com/golangid/candi/candihelper"
)

const (
	validateValue = "{0}"

	customTagRegexp = "regexp"

	// RegexAlphabetLower const
	RegexAlphabetLower = "a-z"
	// RegexAlphabetUpper const
	RegexAlphabetUpper = "A-Z"
	// RegexNumeric const
	RegexNumeric = "0-9"
	// RegexDash const
	RegexDash = "-"

	// AlphabetLower const
	AlphabetLower = "alfabet kecil"
	// AlphabetUpper const
	AlphabetUpper = "alfabet besar"
	// Numeric const
	Numeric = "numerik"
	// Dash const
	Dash = "strip"
)

// regex replace
var regexList = map[string]string{
	RegexAlphabetLower: fmt.Sprintf("%s[%s]", AlphabetLower, RegexAlphabetLower),
	RegexAlphabetUpper: fmt.Sprintf("%s[%s]", AlphabetUpper, RegexAlphabetUpper),
	RegexNumeric:       fmt.Sprintf("%s[%s]", Numeric, RegexNumeric),
	RegexDash:          fmt.Sprintf("%s[%s]", Dash, RegexDash),
}

// StructValidatorOptionFunc type
type StructValidatorOptionFunc func(*StructValidator)

// SetTranslatorStructValidatorOption option func
func SetTranslatorStructValidatorOption(translator ut.Translator) StructValidatorOptionFunc {
	return func(v *StructValidator) {
		v.translator = translator
	}
}

// SetCoreStructValidatorOption option func
func SetCoreStructValidatorOption(additionalConfigFunc ...func(*validatorEngine.Validate)) StructValidatorOptionFunc {

	ve := validatorEngine.New()
	for _, additionalFunc := range additionalConfigFunc {
		additionalFunc(ve)
	}
	return func(v *StructValidator) {
		v.validator = ve
	}
}

// StructValidator struct
type StructValidator struct {
	translator ut.Translator
	validator  *validatorEngine.Validate
}

// NewStructValidator using go library
// https://github.com/go-playground/validator (all struct tags will be here)
// https://godoc.org/github.com/go-playground/validator (documentation using it)
// NewStructValidator function
func NewStructValidator(opts ...StructValidatorOptionFunc) *StructValidator {

	// set default option
	// set lang id locales
	id := idLocales.New()
	// set universal translator
	uni := ut.New(id, id)
	// set translator
	translator, _ := uni.GetTranslator("id")

	// set validator
	vv := validatorEngine.New()
	vv.RegisterValidation(customTagRegexp, checkRegex, false)
	if err := vv.RegisterTranslation(customTagRegexp, translator,
		func(ut ut.Translator) error {
			return ut.Add(customTagRegexp, fmt.Sprintf("Parameter %s harus berupa =", validateValue), true) // see universal-translator for details
		},
		func(ut ut.Translator, fe validatorEngine.FieldError) string {
			t, _ := ut.T(customTagRegexp, fe.Field())
			return t
		},
	); err != nil {
		log.Println(candihelper.StringYellow(fmt.Sprintf("Struct Validator: warning, failed set translation validator on tag [%s]", customTagRegexp)))
	}

	// register id translations
	idTranslations.RegisterDefaultTranslations(vv, translator)

	// set struct validator
	sv := &StructValidator{
		translator: translator,
		validator:  vv,
	}

	// overide with custom option
	for _, opt := range opts {
		opt(sv)
	}

	return sv
}

// regexError function
func (v *StructValidator) regexError(errString string) string {
	var (
		result   string
		replacer []string
	)

	for old, new := range regexList {
		replacer = append(replacer, old, fmt.Sprintf(" %s", new))
	}

	r := strings.NewReplacer(replacer...)
	result = r.Replace(errString)

	return result
}

// CUSTOM FUNCTION SECTION
// checkRegex function
func checkRegex(fl validatorEngine.FieldLevel) bool {
	var (
		param  = fl.Param()
		value  = fl.Field().String()
		result = true
	)

	// regexp
	regex := regexp.MustCompile(fmt.Sprintf(`^[%s]+$+`, param))

	if !regex.MatchString(value) {
		result = false
	}

	return result
}

// ValidateStruct function
func (v *StructValidator) ValidateStruct(data interface{}) error {
	multiError := candihelper.NewMultiError()

	err := v.validator.Struct(data)
	if err != nil {
		switch errs := err.(type) {
		case validatorEngine.ValidationErrors:
			for _, e := range errs {
				message := e.Translate(v.translator)

				if e.Tag() == customTagRegexp {
					message = fmt.Sprintf("%s%s", message, v.regexError(e.Param()))
				}

				// can translate each error one at a time.
				multiError.Append(strings.ToLower(e.Field()), fmt.Errorf(message))

				if multiError.HasError() {
					return multiError
				}
			}
		default:
			return err
		}
	}

	return nil
}
