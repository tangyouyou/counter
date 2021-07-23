package counter

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

var redisDb *redis.Client

var keyInt8 string
var keyInt16 string
var keyInt32 string

var keyUInt8 string
var keyUInt16 string
var keyUInt32 string

func init() {
	host := "192.168.244.78"
	port := 6379
	addr := fmt.Sprintf("%s:%d", host, port)
	pass := ""
	db := 0
	redisDb = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DB:           db,
		DialTimeout:  time.Duration(5) * time.Second, // 连接超时
		ReadTimeout:  time.Duration(3) * time.Second, // 读取超时
		WriteTimeout: time.Duration(3) * time.Second, // 写入超时
	})

	keyInt8 = "user_int8_190950"
	keyInt16 = "user_int16_190950"
	keyInt32 = "user_int32_190950"

	keyUInt8 = "user_uint8_190950"
	keyUInt16 = "user_uint16_190950"
	keyUInt32 = "user_uint32_190950"
}

func TestCounterCluster_SetInt8Value(t *testing.T) {
	var err error
	counter, _ := NewCounter(redisDb, keyInt8, int8Bits)

	err = counter.SetValue(0, 29)
	if err != nil {
		t.Fatal(err)
	}
	err = counter.SetValue(1, math.MaxInt8)
	assert.Nil(t, err)

	err = counter.SetValue(2, math.MaxInt8 + 1)
	assert.Error(t, err)

	err = counter.SetValue(3, math.MinInt8)
	assert.Nil(t, err)

	err = counter.SetValue(4, math.MinInt8-1)
	assert.Error(t, err)
}

func TestCounterCluster_SetUInt8Value(t *testing.T) {
	var err error
	counter, _ := NewCounter(redisDb, keyUInt8, uint8Bits)

	err = counter.SetValue(0, 29)
	assert.Nil(t, err)

	err = counter.SetValue(1, math.MaxInt8)
	assert.Nil(t, err)

	err = counter.SetValue(2, math.MaxInt8 + 1)
	if err != nil {
		t.Fatal(err)
	}

	err = counter.SetValue(3, -1)
	assert.Error(t, err)
}

func TestCounterCluster_SetInt16Value(t *testing.T) {
	var err error
	counter, _ := NewCounter(redisDb, keyInt16, int16Bits)

	err = counter.SetValue(0, 32767)
	if err != nil {
		t.Fatal(err)
	}
	err = counter.SetValue(1, math.MaxInt16)
	assert.Nil(t, err)

	err = counter.SetValue(2, math.MaxInt16 + 133)
	assert.Error(t, err)

	err = counter.SetValue(3, math.MinInt16)
	assert.Nil(t, err)

	err = counter.SetValue(4, math.MinInt16-1)
	assert.Error(t, err)
}

func TestCounterCluster_Incr(t *testing.T) {
	counter,_ := NewCounter(redisDb, keyInt32, int32Bits)
	counter.SetValue(0, 100)
	num0, err := counter.GetValue(0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, num0, 100)

	for i := 0; i < 100; i++ {
		num0, _ = counter.Incr(0)
	}
	assert.Equal(t, num0, 200)
}

func TestCounterCluster_IncrCount(t *testing.T) {
	counter,_ := NewCounter(redisDb, keyInt32, int32Bits)
	err := counter.SetValue(0, 100)
	assert.Nil(t, err)

	num0, err := counter.GetValue(0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, num0, 100)

	num0, _ = counter.IncrCount(0, 10000000)
	assert.Equal(t, num0, 10000000 + 100)
}

func TestCounterCluster_Decr(t *testing.T) {
	counter,_ := NewCounter(redisDb, keyInt32, int32Bits)
	err := counter.SetValue(0, 100)
	if err != nil {
		t.Fatal(err)
	}
	num0, _ := counter.Decr(0)
	assert.Equal(t, num0, 99)

	for i := 0; i < 90; i++ {
		num0, _ = counter.Decr(0)
	}
	assert.Equal(t, num0, 9)
}

func TestCounterCluster_DecrCount(t *testing.T) {
	counter,_ := NewCounter(redisDb, keyInt32, int32Bits)
	err := counter.SetValue(0, 100000000)
	if err != nil {
		t.Fatal(err)
	}

	num0, _ := counter.DecrCount(0, 18888888888)
	assert.Equal(t, num0, 100000000 - 18888888888)
}





