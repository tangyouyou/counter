package counter

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"strconv"
	"strings"
)

const (
	int8Bits  = uint(8)
	int16Bits = uint(16)
	int32Bits = uint(32)
	int64Bits = uint(64)
	signPositive = 1
	signNegative = -1
	setScript = `
local offsetIndex = 2
for _, value in ipairs(ARGV) do
	redis.call("setbit", KEYS[1], KEYS[offsetIndex], value)
    offsetIndex = offsetIndex + 1
end
`
	getScript = `
local bitArr = {}
local bitIndex = 1
for _, offset in ipairs(ARGV) do
	bitArr[bitIndex] = tostring(redis.call("getbit", KEYS[1], offset))
	bitIndex = bitIndex + 1
end
return bitArr
`
)

type (
	Counter interface {
		SetInt8Value(offset uint32, value int8) error
		SetInt16Value(offset uint32, value int16) error
		SetInt32Value(offset uint32, value int32) error
		SetInt64Value(offset uint32, value int64) error
		SetUInt8Value(offset uint32, value uint8) error
		SetUInt16Value(offset uint32, value uint16) error
		SetUInt32Value(offset uint32, value uint32) error
		SetUInt64Value(offset uint32, value uint64) error
		GetInt8Value(offset uint32) (int, error)
		GetInt16Value(offset uint32) (int, error)
		GetInt32Value(offset uint32) (int, error)
		GetInt64Value(offset uint32) (int, error)
		GetUInt8Value(offset uint32) (int, error)
		GetUInt16Value(offset uint32) (int, error)
		GetUInt32Value(offset uint32) (int, error)
		GetUInt64Value(offset uint32) (int, error)
	}

	counterCluster struct {
		r   *redis.Client
		key string
		bits uint
	}
)

func NewCounter(redisDb *redis.Client, key string) Counter {
	counter := counterCluster{
		r:   redisDb,
		key: key,
	}

	return counter
}

// Binary coded decimal is converted to Binary
func convertBcdToBin(n int, bin int) (string, error) {
	var b string
	switch {
	case n == 0:
		for i := 0; i < bin; i++ {
			b += "0"
		}
	case n > 0:
		for ; n > 0; n /= 2 {
			b = strconv.Itoa(n%2) + b
		}
		//加0
		j := bin - len(b)
		for i := 0; i < j; i++ {
			b = "0" + b
		}
	case n < 0:
		n = n * -1
		s, _ := convertBcdToBin(n, bin)
		for i := 0; i < len(s); i++ {
			if s[i:i+1] == "1" {
				b += "0"
			} else {
				b += "1"
			}
		}
		n, err := strconv.ParseInt(b, 2, 64)
		if err != nil {
			return "", err
		}
		b, _ = convertBcdToBin(int(n+1), bin)
	}
	return b, nil
}

// Binary is converted to Binary coded decimal
func covertBinToBcd(s []string) (num int) {
	l := len(s)
	for i := l - 1; i >= 0; i-- {
		str := s[l-i-1]
		tmpInt, _ := strconv.Atoi(str)
		num += (tmpInt & 0xf) << uint8(i)
	}
	return
}

// Build value to offset
func (c counterCluster) buildOffsetArgs(value int, bits uint) ([]string, error) {
	var args []string

	binStr, err := convertBcdToBin(value, int(bits))

	args = strings.Split(binStr, "")

	return args, err
}

func (c counterCluster) setValue(offset uint32, value int, bits uint) error {
	args, _ := c.buildOffsetArgs(value, bits)
	keys := make([]string, 0)
	keys = append(keys, c.key)
	startOffset := int(offset * uint32(bits))
	for i := 0; i < int(bits); i++ {
		currentOffset := startOffset + i
		keys = append(keys, strconv.Itoa(currentOffset))
	}
	cmd := c.r.Eval(setScript, keys, args)

	_, err := cmd.Result()

	if err == redis.Nil {
		return nil
	}

	return err
}

func (c counterCluster) SetInt8Value(offset uint32, value int8) error {
	return c.setValue(offset, int(value), int8Bits)
}

func (c counterCluster) SetInt16Value(offset uint32, value int16) error {
	return c.setValue(offset, int(value), int16Bits)
}

func (c counterCluster) SetInt32Value(offset uint32, value int32) error {
	return c.setValue(offset, int(value), int32Bits)
}

func (c counterCluster) SetInt64Value(offset uint32, value int64) error {
	return c.setValue(offset, int(value), int64Bits)
}

func (c counterCluster) SetUInt8Value(offset uint32, value uint8) error {
	return c.setValue(offset, int(value), int8Bits)
}

func (c counterCluster) SetUInt16Value(offset uint32, value uint16) error {
	return c.setValue(offset, int(value), int16Bits)
}

func (c counterCluster) SetUInt32Value(offset uint32, value uint32) error {
	return c.setValue(offset, int(value), int32Bits)
}

func (c counterCluster) SetUInt64Value(offset uint32, value uint64) error {
	return c.setValue(offset, int(value), int64Bits)
}

func (c counterCluster) getValue(offset uint32, bits uint, sign int) (int, error) {
	args := make([]string, 0)
	startOffset := int(offset * uint32(bits))
	for i := 0; i < int(bits); i++ {
		currentOffset := startOffset + i
		args = append(args, strconv.Itoa(currentOffset))
	}
	cmd := c.r.Eval(getScript, []string{c.key}, args)

	var binArr []string
	res, err := cmd.Result()
	inputByte, _ := json.Marshal(res)
	_ = json.Unmarshal(inputByte, &binArr)
	var num int
	if sign == -1 && binArr[0] == "1" {
		// the sign bit is reversed
		for i := 0; i < len(binArr); i++ {
			if binArr[i] == "0" {
				binArr[i] = "1"
			} else {
				binArr[i] = "0"
			}
		}
		num = covertBinToBcd(binArr)
		num = 0 - (num + 1)
	} else {
		num = covertBinToBcd(binArr)
	}

	if err == redis.Nil {
		return num, nil
	}

	return num, err
}

func (c counterCluster) GetInt8Value(offset uint32) (int, error) {
	return c.getValue(offset, int8Bits, signNegative)
}

func (c counterCluster) GetInt16Value(offset uint32) (int, error) {
	return c.getValue(offset, int16Bits, signNegative)
}

func (c counterCluster) GetInt32Value(offset uint32) (int, error) {
	return c.getValue(offset, int32Bits, signNegative)
}

func (c counterCluster) GetInt64Value(offset uint32) (int, error) {
	return c.getValue(offset, int64Bits, signNegative)
}

func (c counterCluster) GetUInt8Value(offset uint32) (int, error) {
	return c.getValue(offset, int8Bits, signPositive)
}

func (c counterCluster) GetUInt16Value(offset uint32) (int, error) {
	return c.getValue(offset, int16Bits, signPositive)
}

func (c counterCluster) GetUInt32Value(offset uint32) (int, error) {
	return c.getValue(offset, int32Bits, signPositive)
}

func (c counterCluster) GetUInt64Value(offset uint32) (int, error) {
	return c.getValue(offset, int64Bits, signPositive)
}

func (c counterCluster) Incr(offset uint32) (int, error) {
	// 并发一致性保证 & 数据位数调整？ todo lua 脚本
	curNum, err := c.getValue(offset, int32Bits, signPositive)
	curNum = curNum + 1
	err = c.setValue(offset, curNum, int32Bits)
	return curNum, err
}

func (c counterCluster) Decr(offset uint32) (int, error) {
	// 并发一致性保证 & 数据位数调整？ todo lua 脚本
	curNum, err := c.getValue(offset, int32Bits, signPositive)
	curNum = curNum - 1
	err = c.setValue(offset, curNum, int32Bits)
	return curNum, err
}