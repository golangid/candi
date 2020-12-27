package candihelper

import (
	"time"
)

const (
	// Version of this library
	Version = "v1.2.0"
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

	// WORKDIR const for workdir environment
	WORKDIR = "WORKDIR"
	// RepositorySQL unit of work for sql repository
	RepositorySQL = "repositorySQL"
	// RepositoryMongo unit of work for mongodb repository
	RepositoryMongo = "repositoryMongo"
	// UsecaseUOW unit of work for usecase
	UsecaseUOW = "usecaseUOW"
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
