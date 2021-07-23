package counter

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"math"
	"strconv"
	"strings"
)

const (
	int8Bits     = 8
	int16Bits    = 16
	int32Bits    = 32
	uint8Bits    = -8
	uint16Bits   = -16
	uint32Bits   = -32
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
		SetValue(offset uint32, value int) error
		GetValue(offset uint32) (int, error)
		Incr(offset uint32) (int, error)
		Decr(offset uint32) (int, error)
		IncrCount(offset uint32, count int) (int, error)
		DecrCount(offset uint32, count int) (int, error)
	}

	counterCluster struct {
		r    *redis.Client
		key  string
		bits int
	}
)

func NewCounter(redisDb *redis.Client, key string, bits int) (Counter, error) {
	if bits != int8Bits && bits != int16Bits && bits != int32Bits &&
		bits != uint8Bits && bits != uint16Bits && bits != uint32Bits  {
		return nil, fmt.Errorf("the bits is invalid")
	}
	counter := counterCluster{
		r:    redisDb,
		key:  key,
		bits: bits,
	}

	return counter, nil
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

func (c counterCluster) validValue(value int) error {
	switch c.bits {
	case int8Bits:
		if value > math.MaxInt8 || value < math.MinInt8 {
			return fmt.Errorf("the value overflows int8")
		}
	case uint8Bits:
		if value > math.MaxUint8 || value < 0 {
			return fmt.Errorf("the value overflows uint8")
		}
	case int16Bits:
		if value > math.MaxInt16 || value < math.MinInt16 {
			return fmt.Errorf("the value overflows int16")
		}
	case uint16Bits:
		if value > math.MaxUint16 || value < 0 {
			return fmt.Errorf("the value overflows uint16")
		}
	case int32Bits:
		if value > math.MaxInt32 || value < math.MinInt32 {
			return fmt.Errorf("the value overflows int32")
		}
	case uint32Bits:
		if value > math.MaxUint32 || value < 0 {
			return fmt.Errorf("the value overflows uint32")
		}
	}
	return nil
}

// Build value to offset
func (c counterCluster) buildOffsetArgs(value int, bits int) ([]string, error) {
	var args []string

	binStr, err := convertBcdToBin(value, int(bits))

	args = strings.Split(binStr, "")

	return args, err
}

func (c counterCluster) setValue(offset uint32, value int, bits int) error {
	if err := c.validValue(value); err != nil {
		return err
	}
	if bits < 0 {
		bits = -bits
	}
	args, _ := c.buildOffsetArgs(value, bits)
	keys := make([]string, 0)
	keys = append(keys, c.key)
	startOffset := int(offset * uint32(bits))
	for i := 0; i < bits; i++ {
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

func (c counterCluster) getValue(offset uint32, bits int, sign int) (int, error) {
	args := make([]string, 0)
	startOffset := int(offset * uint32(bits))
	for i := 0; i < bits; i++ {
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

func (c counterCluster) SetValue(offset uint32, value int) error {
	return c.setValue(offset, value, c.bits)
}

func (c counterCluster) GetValue(offset uint32) (int, error) {
	sign := signPositive
	if c.bits < 0 {
		sign = signNegative
	}
	return c.getValue(offset, c.bits, sign)
}

func (c counterCluster) Incr(offset uint32) (int, error) {
	// 并发一致性保证 & 数据位数调整？ todo lua 脚本
	value, err := c.GetValue(offset)
	if err != nil {
		return value, err
	}
	value = value + 1
	if err := c.validValue(value); err != nil {
		return value, err
	}
	err = c.SetValue(offset, value)
	return value, err
}

func (c counterCluster) Decr(offset uint32) (int, error) {
	// 并发一致性保证 & 数据位数调整？ todo lua 脚本
	value, err := c.GetValue(offset)
	if err != nil {
		return value, err
	}
	value = value - 1
	if err := c.validValue(value); err != nil {
		return value, err
	}
	err = c.SetValue(offset, value)
	return value, err
}

func (c counterCluster) IncrCount(offset uint32, count int) (int, error) {
	// 并发一致性保证 & 数据位数调整？ todo lua 脚本
	value, err := c.GetValue(offset)
	if err != nil {
		return value, err
	}
	value = value + count
	if err := c.validValue(value); err != nil {
		return value, err
	}
	err = c.SetValue(offset, value)
	return value, err
}

func (c counterCluster) DecrCount(offset uint32, count int) (int, error) {
	// 并发一致性保证 & 数据位数调整？ todo lua 脚本
	value, err := c.GetValue(offset)
	if err != nil {
		return value, err
	}
	value = value - count
	if err := c.validValue(value); err != nil {
		return value, err
	}
	err = c.SetValue(offset, value)
	return value, err
}
