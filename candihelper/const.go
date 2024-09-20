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
	// HeaderCacheControl header const
	HeaderCacheControl = "Cache-Control"
	// HeaderExpires header const
	HeaderExpires = "Expires"
	// HeaderLastModified header const
	HeaderLastModified = "Last-Modified"
	// HeaderIfModifiedSince header const
	HeaderIfModifiedSince = "If-Modified-Since"
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
	// AsiaJakartaLocalTime location
	AsiaJakartaLocalTime *time.Location
)

func init() {
	AsiaJakartaLocalTime, _ = time.LoadLocation(TimeZoneAsia)
}
