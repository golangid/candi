package validator

import (
	"errors"
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

// StructValidator abstraction
type StructValidator interface {
	ValidateStruct(data interface{}) error
}

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
type StructValidatorOptionFunc func(*structValidator)

// SetTranslatorStructValidatorOption option func
func SetTranslatorStructValidatorOption(translator ut.Translator) StructValidatorOptionFunc {
	return func(v *structValidator) {
		v.translator = translator
	}
}

// SetCoreStructValidatorOption option func
func SetCoreStructValidatorOption(additionalConfigFunc ...func(*validatorEngine.Validate)) StructValidatorOptionFunc {

	return func(v *structValidator) {
		ve := validatorEngine.New()
		for _, additionalFunc := range additionalConfigFunc {
			additionalFunc(ve)
		}
		v.validator = ve
	}
}

// getDefaultStructTranslatorConfig config func
func getDefaultStructTranslatorConfig(vv *validatorEngine.Validate) ut.Translator {

	// set default option
	// set lang id locales
	id := idLocales.New()
	// set universal translator
	uni := ut.New(id, id)
	// set translator
	translator, _ := uni.GetTranslator("id")

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

	idTranslations.RegisterDefaultTranslations(vv, translator)
	return translator
}

// checkRegex function
func checkRegex(fl validatorEngine.FieldLevel) bool {
	return !regexp.MustCompile(fmt.Sprintf(`^[%s]+$+`, fl.Param())).MatchString(fl.Field().String())
}

// structValidator struct
type structValidator struct {
	translator ut.Translator
	validator  *validatorEngine.Validate
}

// NewStructValidator using go library
// https://github.com/go-playground/validator (all struct tags will be here)
// https://godoc.org/github.com/go-playground/validator (documentation using it)
// NewStructValidator function
func NewStructValidator(opts ...StructValidatorOptionFunc) StructValidator {

	// set struct validator
	sv := &structValidator{}
	for _, opt := range opts {
		opt(sv)
	}

	if sv.validator == nil {
		sv.validator = validatorEngine.New()
	}
	if sv.translator == nil {
		sv.translator = getDefaultStructTranslatorConfig(sv.validator)
	}

	return sv
}

// regexError function
func (v *structValidator) regexError(errString string) string {
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

// ValidateStruct function
func (v *structValidator) ValidateStruct(data interface{}) error {

	err := v.validator.Struct(data)
	if err != nil {
		switch errs := err.(type) {
		case validatorEngine.ValidationErrors:
			multiError := candihelper.NewMultiError()
			for _, e := range errs {
				message := err.Error()
				if v.translator != nil {
					message = e.Translate(v.translator)
				}
				if e.Tag() == customTagRegexp {
					message = message + v.regexError(e.Param())
				}
				multiError.Append(strings.ToLower(e.Field()), errors.New(message))
			}
			if multiError.HasError() {
				return multiError
			}
		default:
			return err
		}
	}

	return nil
}
