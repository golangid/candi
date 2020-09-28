package validator

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	idLocales "github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	idTranslations "github.com/go-playground/validator/v10/translations/id"
	"pkg.agungdwiprasetyo.com/candi/helper"
)

const (
	validateValue = "{0}"

	customTagRegexp = "regexp"

	RegexAlphabetLower = "a-z"
	RegexAlphabetUpper = "A-Z"
	RegexNumeric       = "0-9"
	RegexDash          = "-"

	AlphabetLower = "alfabet kecil"
	AlphabetUpper = "alfabet besar"
	Numeric       = "numerik"
	Dash          = "strip"
)

// custom error list
var errorList = map[string]string{
	customTagRegexp: fmt.Sprintf("Parameter %s harus berupa =", validateValue),
}

// regex replace
var regexList = map[string]string{
	RegexAlphabetLower: fmt.Sprintf("%s[%s]", AlphabetLower, RegexAlphabetLower),
	RegexAlphabetUpper: fmt.Sprintf("%s[%s]", AlphabetUpper, RegexAlphabetUpper),
	RegexNumeric:       fmt.Sprintf("%s[%s]", Numeric, RegexNumeric),
	RegexDash:          fmt.Sprintf("%s[%s]", Dash, RegexDash),
}

// custom function
var funcList = map[string]validator.Func{
	customTagRegexp: checkRegex,
}

// StructValidator struct
type StructValidator struct {
	translator ut.Translator
	validator  *validator.Validate
}

// using go library
// https://github.com/go-playground/validator (all struct tags will be here)
// https://godoc.org/github.com/go-playground/validator (documentation using it)
// NewStructValidator function
func NewStructValidator() *StructValidator {
	// set lang id locales
	id := idLocales.New()

	// set universal translator
	uni := ut.New(id, id)

	// set translator
	translator, _ := uni.GetTranslator("id")

	// set validator
	validator := validator.New()

	// register id translations
	idTranslations.RegisterDefaultTranslations(validator, translator)

	// set struct validator
	structValidator := &StructValidator{
		translator: translator,
		validator:  validator,
	}

	// add custom function
	if len(funcList) > 0 {
		for tag, function := range funcList {
			structValidator.customFunc(tag, function)
		}
	}

	// override translation
	if len(errorList) > 0 {
		for tag, message := range errorList {
			structValidator.setTranslationOverride(tag, message)
		}
	}

	return &StructValidator{
		translator: translator,
		validator:  validator,
	}
}

// setTranslationOverride function
func (v *StructValidator) setTranslationOverride(tag string, message string) {
	// override error
	err := v.validator.RegisterTranslation(tag, v.translator, v.registerFunc(tag, message), v.translationFunc(tag))
	if err != nil {
		log.Println(helper.StringYellow(fmt.Sprintf("Struct Validator: warning, failed set translation validator on tag [%s]", tag)))
	}
}

// registerFunc function
func (v *StructValidator) registerFunc(tag string, message string) validator.RegisterTranslationsFunc {
	register := func(ut ut.Translator) error {
		return ut.Add(tag, message, true) // see universal-translator for details
	}
	return register
}

// translationFunc function
func (v *StructValidator) translationFunc(tag string) validator.TranslationFunc {
	trans := func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field())

		return t
	}
	return trans
}

// customFunc function
func (v *StructValidator) customFunc(tag string, function validator.Func) {
	v.validator.RegisterValidation(tag, function, false)
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
func checkRegex(fl validator.FieldLevel) bool {
	var (
		param       = fl.Param()
		value       = fl.Field().String()
		result bool = true
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
	multiError := helper.NewMultiError()

	err := v.validator.Struct(data)
	if err != nil {
		errs := err.(validator.ValidationErrors)

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
	}

	return nil
}
