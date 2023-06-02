package candihelper

import (
	"time"
)

const (
	// TimeZoneAsia constanta
	TimeZoneAsia = "Asia/Jakarta"
	// TokenClaimKey const
	TokenClaimKey = "tokenClaim"

	// TimeFormatLogger const
	TimeFormatLogger = "2006/01/02 15:04:05"

	// V1 const
	V1 = "/v1"
	// V2 const
	V2 = "/v2"

	// Byte ...
	Byte uint64 = 1
	// KByte ...
	KByte = Byte * 1024
	// MByte ...
	MByte = KByte * 1024
	// GByte ...
	GByte = MByte * 1024
	// TByte ...
	TByte = GByte * 1024

	// WORKDIR const for workdir environment
	WORKDIR = "WORKDIR"
	// RepositorySQL unit of work for sql repository
	RepositorySQL = "repositorySQL"
	// RepositoryMongo unit of work for mongodb repository
	RepositoryMongo = "repositoryMongo"

	// HeaderDisableTrace const
	HeaderDisableTrace = "X-Disable-Trace"
	// HeaderXForwardedFor const
	HeaderXForwardedFor = "X-Forwarded-For"
	// HeaderXRealIP const
	HeaderXRealIP = "X-Real-IP"
	// HeaderContentType const
	HeaderContentType = "Content-Type"
	// HeaderAuthorization const
	HeaderAuthorization = "Authorization"
	// HeaderMIMEApplicationJSON const
	HeaderMIMEApplicationJSON = "application/json"
	// HeaderMIMEApplicationXML const
	HeaderMIMEApplicationXML = "application/xml"
	// HeaderMIMEApplicationForm const
	HeaderMIMEApplicationForm = "application/x-www-form-urlencoded"
	// HeaderMIMEMultipartForm const
	HeaderMIMEMultipartForm = "multipart/form-data"
	// HeaderMIMEOctetStream const
	HeaderMIMEOctetStream = "application/octet-stream"

	// DateFormatMonday date format
	DateFormatMonday = "Monday"
	// DateFormatYYYYMM date format
	DateFormatYYYYMM = "2006-01"
	// DateFormatYYYYMMDD date format
	DateFormatYYYYMMDD = "2006-01-02"
	// DateFormatYYYYMMDDHHmmss date format
	DateFormatYYYYMMDDHHmmss = "2006-01-02 15:04:05"
	// DateFormatYYYYMMDDClean date format
	DateFormatYYYYMMDDClean = "20060102"
	// DateFormatHHmmss date format
	DateFormatHHmm = "15:04"
	// DateFormatDDMMYYYY date format
	DateFormatDDMMYYYY = "02-01-2006"
)

var (
	// Green color
	Green = []byte{27, 91, 57, 55, 59, 52, 50, 109}
	// White color
	White = []byte{27, 91, 57, 48, 59, 52, 55, 109}
	// Yellow color
	Yellow = []byte{27, 91, 57, 48, 59, 52, 51, 109}
	// Red color
	Red = []byte{27, 91, 57, 55, 59, 52, 49, 109}
	// Blue color
	Blue = []byte{27, 91, 57, 55, 59, 52, 52, 109}
	// Magenta color
	Magenta = []byte{27, 91, 57, 55, 59, 52, 53, 109}
	// Cyan color
	Cyan = []byte{27, 91, 57, 55, 59, 52, 54, 109}
	// Reset color
	Reset = []byte{27, 91, 48, 109}

	// AsiaJakartaLocalTime location
	AsiaJakartaLocalTime *time.Location
)

func init() {
	AsiaJakartaLocalTime, _ = time.LoadLocation(TimeZoneAsia)
}
