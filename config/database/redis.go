package database

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

// InitRedis connection
func InitRedis(isUse bool) (readPool, writePool *redis.Pool) {
	if !isUse {
		return
	}

	hostRead, portRead, passRead := os.Getenv("REDIS_READ_HOST"), os.Getenv("REDIS_READ_PORT"), os.Getenv("REDIS_READ_AUTH")
	tlsRead, _ := strconv.ParseBool(os.Getenv("REDIS_READ_TLS"))
	hostWrite, portWrite, passWrite := os.Getenv("REDIS_WRITE_HOST"), os.Getenv("REDIS_WRITE_PORT"), os.Getenv("REDIS_WRITE_AUTH")
	tlsWrite, _ := strconv.ParseBool(os.Getenv("REDIS_WRITE_TLS"))

	readPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%s", hostRead, portRead), redis.DialPassword(passRead), redis.DialUseTLS(tlsRead))
		},
	}

	pingRead := readPool.Get()
	defer pingRead.Close()
	_, err := pingRead.Do("PING")
	if err != nil {
		panic("redis read: " + err.Error())
	}

	writePool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%s", hostWrite, portWrite), redis.DialPassword(passWrite), redis.DialUseTLS(tlsWrite))
		},
	}

	pingWrite := writePool.Get()
	defer pingWrite.Close()
	_, err = pingWrite.Do("PING")
	if err != nil {
		panic("redis write: " + err.Error())
	}

	log.Println("Success load Redis connection")
	return
}
