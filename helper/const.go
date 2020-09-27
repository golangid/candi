package helper

import (
	"errors"
	"time"
)

const (
	// TimeZoneAsia constanta
	TimeZoneAsia = "Asia/Jakarta"
	// TokenClaimKey const
	TokenClaimKey = "tokenClaim"
	// TokenKey const
	TokenKey = "18608c7d-b319-0xc000165c80-0xc0000da000-11478e4e2650"

	// TimeFormatLogger const
	TimeFormatLogger = "2006/01/02 15:04:05"

	// V1 const
	V1 = "/v1"
	// V2 const
	V2 = "/v2"

	// GRPCBanner const
	GRPCBanner = `    __________  ____  ______
   / ____/ __ \/ __ \/ ____/
  / / __/ /_/ / /_/ / /     
 / /_/ / _, _/ ____/ /___   
 \____/_/ |_/_/    \____/   							
`

	// RedisBanner const
	RedisBanner = `    ____  __________  _________
   / __ \/ ____/ __ \/  _/ ___/
  / /_/ / __/ / / / // / \__ \ 
 / _, _/ /___/ /_/ // / ___/ / 
/_/ |_/_____/_____/___//____/                                
`

	// Byte ...
	Byte uint64 = 1
	// KByte ...
	KByte = Byte * 1024
	// MByte ...
	MByte = KByte * 1024
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

	// ErrTokenFormat var
	ErrTokenFormat = errors.New("Invalid token format")
	// ErrTokenExpired var
	ErrTokenExpired = errors.New("Token is expired")

	// AsiaJakartaLocalTime location
	AsiaJakartaLocalTime *time.Location
)

func init() {
	AsiaJakartaLocalTime, _ = time.LoadLocation(TimeZoneAsia)
}
