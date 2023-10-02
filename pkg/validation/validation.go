package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	maxURLRuneCount        = 2083
	minURLRuneCount        = 3
	IP              string = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	URLUsername     string = `(\S+(:\S*)?@)`
	URLPath         string = `((\/|\?|#)[^\s]*)`
	URLSchema       string = `((ftp|tcp|udp|wss?|https?):\/\/)`
	URLPort         string = `(:(\d{1,5}))`
	URLIP           string = `([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))`
	URLSubdomain    string = `((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))`
	URL                    = `^` + URLSchema + `?` + URLUsername + `?` + `((` + URLIP + `|(\[` + IP + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + URLSubdomain + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + URLPort + `?` + URLPath + `?$`
)

var (
	ErrValidationIncorrectCountryCode = errors.New("incorrect country code")
	ErrValidationEmptyString          = errors.New("empty string")
	ErrValidationNumberString         = errors.New("number string")
	ErrValidationShortString          = errors.New("short string")
	ErrValidationLongString           = errors.New("long string")
	ErrValidationGeneric              = errors.New("validation error")
	ErrValidationWeakPassword         = errors.New("weak password")
	ErrValidationInvalidUsername      = errors.New("invalid username")
	ErrValidationInvalidMobilePhone   = errors.New("invalid mobile phone number")
	ErrValidationInvalidEmail         = errors.New("invalid email")
	ErrValidationIncorrectMaxResult   = errors.New("incorrect max result")
	ErrValidationIncorrectMinResult   = errors.New("incorrect min result")
	ErrValidationIncorrectTime        = errors.New("incorrect time")
	ErrValidationInvalidDomain        = errors.New("invalid domain")
	ErrValidationInvalidSubDomain     = errors.New("invalid subdomain")
	ErrValidationInvalidURL           = errors.New("invalid url")

	usernameValidator    = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)
	mobilePhoneValidator = regexp.MustCompile(`^993[6-7][1-5]\d{6}$`)
	emailValidator       = regexp.MustCompile("(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$)")
	keyValidator         = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)
	// Domain regex source: https://stackoverflow.com/a/7933253
	domainValidator = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-zA-Z]{1,63}| xn--[a-z0-9]{1,59})$`)
	// Subdomain regex source: https://stackoverflow.com/a/7933253
	subdomainValidator = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?$`)
	URLValidator       = regexp.MustCompile(URL)

	maxResultsMax = 100
)

type PasswordPolicy struct {
	MinPasswordLength int
	MinNumericSymbols int
	MinUpperCaseChars int
	MinLowerCaseChars int
	MinSpecialChars   int
}

var DefaultPasswordPolicy = PasswordPolicy{
	MinPasswordLength: 8,
	MinNumericSymbols: 2,
	MinUpperCaseChars: 1,
	MinLowerCaseChars: 1,
	MinSpecialChars:   1,
}

type validateFunc func() error

type ValidationBox struct {
	Name string
	Func validateFunc
}

func (b *ValidationBox) Validate() error {
	return b.Func()
}

func IsValid(lazy bool, validations []ValidationBox) bool {
	var validation ValidationBox
	var err error
	var hasErrors = false
	for i := range validations {
		validation = validations[i]
		err = validation.Validate()
		if err != nil {
			hasErrors = true
			log.Error().Err(err).Str("param-name", validation.Name).Msg("validation failed")
			if lazy {
				return false
			}
		}
	}
	return !hasErrors
}

func StringMustBeKey(s string) error {
	if !keyValidator.MatchString(s) {
		return ErrValidationGeneric
	}
	return nil
}

func PasswordValidate(pp PasswordPolicy, password string) error {
	numCount := 0
	upperCount := 0
	lowerCount := 0
	specialCharCount := 0
	totalLength := len(password)
	for _, s := range password {
		if unicode.IsNumber(s) {
			numCount++
		} else if unicode.IsUpper(s) {
			upperCount++
		} else if unicode.IsLetter(s) || s == ' ' {
			lowerCount++
		} else if unicode.IsPunct(s) || unicode.IsSymbol(s) {
			specialCharCount++
		}
	}
	if totalLength >= pp.MinPasswordLength && numCount >= pp.MinNumericSymbols &&
		upperCount >= pp.MinUpperCaseChars && lowerCount >= pp.MinLowerCaseChars &&
		specialCharCount >= pp.MinSpecialChars {
		return nil
	} else {
		log.Error().
			Int("totalLength", totalLength-pp.MinPasswordLength).
			Int("numCount", numCount-pp.MinNumericSymbols).
			Int("upperCount", upperCount-pp.MinUpperCaseChars).
			Int("lowerCount", lowerCount-pp.MinLowerCaseChars).
			Int("specialCount", specialCharCount-pp.MinSpecialChars).
			Msg("password validation failed")
		return ErrValidationWeakPassword
	}
}

func UsernameValidate(s string) error {
	if err := StringMustBeNotEmptyWithMaxLength(s, 32); err != nil {
		return ErrValidationInvalidUsername
	}
	if !usernameValidator.MatchString(s) {
		return ErrValidationInvalidUsername
	}
	return nil
}

func MobilePhoneValidate(s string) error {
	if mobilePhoneValidator.MatchString(s) {
		return nil
	}
	return ErrValidationInvalidMobilePhone
}

func MobilePhoneValidateUser(s string) error {
	if mobilePhoneValidator.MatchString(s) {
		return nil
	}
	return ErrValidationInvalidMobilePhone
}

func EmailValidate(s string) error {
	if err := StringMustBeNotEmptyWithMaxLength(s, 64); err != nil {
		return ErrValidationInvalidEmail
	}
	if !emailValidator.MatchString(s) {
		return ErrValidationInvalidEmail
	}
	return nil
}

func DomainValidate(s string) error {
	// Slightly modified: Removed 255 max length validation since Go regex does not
	// support lookarounds. More info: https://stackoverflow.com/a/38935027
	if err := StringMustBeNotEmptyWithMaxLength(s, 255); err != nil {
		return ErrValidationInvalidDomain
	}
	if !domainValidator.MatchString(s) {
		return ErrValidationInvalidDomain
	}
	return nil
}

func SubDomainValidate(s string) error {
	if !subdomainValidator.MatchString(s) {
		return ErrValidationInvalidSubDomain
	}
	return nil
}

// urlValidate checks if the string is an url.
func UrlValidate(str string) error {
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return ErrValidationInvalidURL
	}
	strTemp := str
	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		strTemp = fmt.Sprintf("http://%s", str)
	}
	u, err := url.Parse(strTemp)
	if err != nil {
		return ErrValidationInvalidURL
	}
	if strings.HasPrefix(u.Host, ".") {
		return ErrValidationInvalidURL
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return ErrValidationInvalidURL
	}
	if !URLValidator.MatchString(str) {
		return ErrValidationInvalidURL
	}
	return nil
}

func StringMustBeNotEmptyWithMinAndMaxLengths(s string, minLength, maxLength int) error {
	r := []rune(s)
	if s == "" {
		return ErrValidationEmptyString
	} else if len(r) < minLength {
		return ErrValidationShortString
	} else if len(r) > maxLength {
		return ErrValidationLongString
	}
	return nil
}

func StringMustBeNotEmptyWithMaxLength(s string, maxLength int) error {
	return StringMustBeNotEmptyWithMinAndMaxLengths(s, 1, maxLength)
}

func StringCanBeEmptyButNotExceedMaxLength(s string, maxLength int) error {
	r := []rune(s)
	if s == "" {
		return nil
	} else if len(r) > maxLength {
		return ErrValidationLongString
	}
	return nil
}

func StringMustBeNotEmptyWithMinLength(s string, minLength int) error {
	r := []rune(s)
	if s == "" {
		return ErrValidationEmptyString
	} else if len(r) < minLength {
		return ErrValidationShortString
	}
	return nil
}
func IsNumber(s string) error {
	for len(s) > 0 {
		rn, sz := utf8.DecodeRuneInString(s)
		if !unicode.IsDigit(rn) {
			return ErrValidationNumberString
		}
		s = s[sz:]
	}
	return nil
}

func MaxResultsStringWithMaxCheck(s string, maxRes int) (maxResult int, err error) {
	maxResult, err = strconv.Atoi(s)
	if err != nil {
		return
	}
	if maxResult < 1 || maxRes < maxResult {
		err = ErrValidationIncorrectMaxResult
		return
	}
	return
}

func IntWithMinCheck(result, minResult int) error {
	if result < minResult {
		return ErrValidationIncorrectMinResult
	}
	return nil
}

func After(start, end time.Time) error {
	if start.After(end) {
		return ErrValidationIncorrectTime
	}
	return nil
}

func TimestampValidate(s string) (t time.Time, err error) {
	var ts int64
	ts, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return
	}
	t = time.Unix(ts, 0).UTC()
	return
}
