package candihelper

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommon(t *testing.T) {
	time := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())

	StringGreen("green")
	StringYellow("yellow")
	assert.NotNil(t, ToBoolPtr(true))
	assert.NotNil(t, ToStringPtr("str"))
	assert.NotNil(t, ToIntPtr(1))
	assert.NotNil(t, ToFloatPtr(1.3))
	assert.NotNil(t, ToFloat32Ptr(1.3))
	assert.NotNil(t, ToTimePtr(time))
	assert.Equal(t, "str", PtrToString(ToStringPtr("str")))
	assert.Equal(t, true, PtrToBool(ToBoolPtr(true)))
	assert.Equal(t, 1, PtrToInt(ToIntPtr(1)))
	assert.Equal(t, 1.3, PtrToFloat(ToFloatPtr(1.3)))
	assert.Equal(t, float32(1.2), PtrToFloat32(ToFloat32Ptr(1.2)))
	assert.Equal(t, time, PtrToTime(ToTimePtr(time)))
	assert.Equal(t, true, StringInSlice("a", []string{"a", "b", "c"}))
	assert.Equal(t, false, StringInSlice("z", []string{"a", "b", "c"}))
	assert.Equal(t, []byte("a"), ToBytes([]byte("a")))
	assert.Equal(t, []byte("a"), ToBytes("a"))
	assert.Equal(t, []byte(`{"a":"a"}`), ToBytes(map[string]string{"a": "a"}))
}

func TestMustParseEnv(t *testing.T) {
	type Embed struct {
		UseHTTP  bool          `env:"USE_HTTP"`
		Float    float64       `env:"FLOAT"`
		Now      time.Time     `env:"NOW"`
		Duration time.Duration `env:"DURATION"`
		Multi    []string      `env:"MULTI"`
	}
	type SubField struct {
		SubString string `env:"SUBSTRING"`
	}

	t.Run("Testcase #1: Positive", func(t *testing.T) {
		now := time.Now().Format(time.RFC3339)
		os.Setenv("HOST", "localhost")
		os.Setenv("PORT", "8000")
		os.Setenv("USE_HTTP", "true")
		os.Setenv("FLOAT", "1.3")
		os.Setenv("NOW", now)
		os.Setenv("DURATION", "10m")
		os.Setenv("UNEXPORTED", "none")
		os.Setenv("SUBSTRING", "substring")
		os.Setenv("MULTI", "a,b,c,d")
		var env struct {
			Host       string `env:"HOST"`
			Port       int    `env:"PORT"`
			unexported string `env:"UNEXPORTED"`
			SubField   SubField
			Embed
		}
		MustParseEnv(&env)
		assert.Equal(t, "localhost", env.Host)
		assert.Equal(t, 8000, env.Port)
		assert.Equal(t, true, env.UseHTTP)
		assert.Equal(t, 1.3, env.Float)
		assert.Equal(t, now, env.Now.Format(time.RFC3339))
		assert.Equal(t, time.Duration(10)*time.Minute, env.Duration)
		assert.Equal(t, "substring", env.SubField.SubString)
		assert.Equal(t, "", env.unexported)
		assert.Equal(t, []string{"a", "b", "c", "d"}, env.Multi)
		os.Clearenv()
	})
	t.Run("Testcase #2: Negative", func(t *testing.T) {
		assert.Panics(t, func() {
			os.Setenv("HOST", "localhost")
			os.Setenv("PORT", "localhost")
			os.Setenv("USE_HTTP", "ok")
			os.Setenv("FLOAT", "ok")
			os.Setenv("NOW", "99:99")
			os.Setenv("DURATION", "a")
			var env struct {
				Ignore   string        `env:"-"`
				Missing  string        `env:"MISSING"`
				Host     string        `env:"HOST"`
				Port     int           `env:"PORT"`
				UseHTTP  bool          `env:"USE_HTTP"`
				Float    float64       `env:"FLOAT"`
				Now      time.Time     `env:"NOW"`
				Duration time.Duration `env:"DURATION"`
			}
			MustParseEnv(&env)
			os.Clearenv()
		})
	})
}

func TestMaskingPasswordURL(t *testing.T) {
	tests := []struct {
		name      string
		stringURL string
		want      string
	}{
		{
			name:      "Testcase #1: Positive",
			stringURL: "mongodb://pass:pass@localhost:27017",
			want:      "mongodb://pass:xxxxx@localhost:27017",
		},
		{
			name:      "Testcase #2: Positive",
			stringURL: "mongodb://pass:@localhost:27017",
			want:      "xxxxx",
		},
		{
			name:      "Testcase #3: Negative",
			stringURL: "()$%!#!#@!",
			want:      "xxxxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskingPasswordURL(tt.stringURL))
		})
	}
}

func TestGetFuncName(t *testing.T) {
	assert.Equal(t, "MustParseEnv", GetFuncName(MustParseEnv))
	assert.Equal(t, "LoadAllFile", GetFuncName(LoadAllFile))
}
