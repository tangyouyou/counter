package counter

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var redisDb *redis.Client

var keyInt8 string
var keyInt16 string
var keyInt32 string
var keyInt64 string

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
	keyInt64 = "user_int64_190950"
}

func TestCounterCluster_SetInt8Value(t *testing.T) {
	var err error
	counter := NewCounter(redisDb, keyInt8)

	err = counter.SetInt8Value(0, 29)
	if err != nil {
		t.Fatal(err)
	}
	err = counter.SetInt8Value(1, 125)
	assert.Nil(t, err)
	err = counter.SetInt8Value(2, -5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCounterCluster_SetInt16Value(t *testing.T) {
	var err error
	counter := NewCounter(redisDb, keyInt16)

	err = counter.SetInt16Value(0, 29)
	if err != nil {
		t.Fatal(err)
	}
	err = counter.SetInt16Value(1, 125)
	assert.Nil(t, err)
	err = counter.SetInt16Value(2, 32767)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCounterCluster_SetInt32Value(t *testing.T) {
	var err error
	counter := NewCounter(redisDb, keyInt32)

	err = counter.SetInt32Value(0, 29)
	if err != nil {
		t.Fatal(err)
	}
	err = counter.SetInt32Value(1, 123456789)
	if err != nil {
		t.Fatal(err)
	}
	err = counter.SetInt32Value(2, -5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCounterCluster_SetInt64Value(t *testing.T) {
	var err error
	counter := NewCounter(redisDb, keyInt64)

	err = counter.SetInt64Value(0, 29)
	if err != nil {
		t.Fatal(err)
	}
	err = counter.SetInt64Value(1, 123456789)
	if err != nil {
		t.Fatal(err)
	}
	err = counter.SetInt64Value(2, 3335555666677)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCounterCluster_GetInt32Value(t *testing.T) {
	counter := NewCounter(redisDb, keyInt32)
	num1, _ := counter.GetInt32Value(0)
	assert.Equal(t, num1, 29)
	num2, _ := counter.GetInt32Value(1)
	assert.Equal(t, num2, 123456789)
	num3, _ := counter.GetInt32Value(2)
	assert.Equal(t, num3, -5)
}
