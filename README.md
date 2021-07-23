# counter
计数服务，采用位图存储，大幅度节省内存，一种高性能、低内存消耗的解决方案

# 背景 #

不少业务场景中，需要存储业务的数字信息，例如：“视频的点赞数量”、“视频的评论数量”、“视频的分享数量”，随着业务的量级提升，业务的数字需要大量的存储空间。业界称之为“计数服务”，关于“计数服务”常见的解决方案

**方案1：MySQL 存储解决方案**

使用  MySQL 存储业务的计数，该方案在低并发场景下可以比较好的支撑业务，但是在高并发的场景，即计数高频率增加时（每秒大于2000次），MySQL 的并发写性能无法支撑计数服务的 TPS，将直接导致系统异常

**方案2：Redis 存储解决方案**

由于 MySQL 无法支撑计数的并发性能，使用 Redis 作为计数服务，常见的使用 Redis 的 string 结构或者 hash 结构存储计数；该方案使用了 Redis，每秒可以支持 3W+的并发写性能，但是 string 结构需要消耗的内存，在数据量持续上升的背景下，需要占用大量的内存空间，而在阿里云上一台8核64G的服务器，一年的使用价格为 15648 元

基于方案1、方案2，那么有没有一种即能支持高并发的写入Tps、又能大幅度节省内存的空间的方案呢？这就是本方案的目的，同样使用 Redis 作为底层的存储，但是不再直接使用 string 的方式存储数值，而是基于 Bitmap（位图）的数据结构存储。Redis 默认的键值最少需要占用“56”字节的存储空间，关键就在于节省的 “56”个字节的反复创建消耗，键值数量越多、节省内存效果越好，是一种可持续发展的“计数服务”解决方案。

## 测试数据的搭建 ##
**附录资料**

Redis 默认的内存占用为 874K，为了方便统计，默认 Redis 启动后需要占用 1M 的内存。

![avatar](https://github.com/tangyouyou/counter/blob/main/redis-1.png)

**图1:Redis 默认内存占用**

**方案1：存储为 int 类型的数字**

从 1, 500000，分别设置 video_$i_ding、video_$i_comment、video_$i_share 等3个计数器，代表的是 “视频点赞数量”、“视频评论数量”、“视频分享数量”
计数器的数量统一采用 2147483647，代表32位有符号整数的最大值
导入数据，查看 Redis 目前的内存使用量

![avatar](https://github.com/tangyouyou/counter/blob/main/redis-2.png)

**图2：int 类型存储内存消耗**

结论：
从内存消耗可以看到，150万个计数器的占用内存为 108 M - 1M = 107M

**方案2：存储为 Bitmap 类型的字符串**

从 1, 500000，分别设置 video_$i 计数器，其中 0 - 31 位代表 “视频点赞数量”、32-63 位代表“视频评论数量”、64-95位“视频分享数量”
计数器的数量统一采用 2147483647，代表32位有符号整数的最大值
导入数据，查看 Redis 目前的内存使用量

![avatar](https://github.com/tangyouyou/counter/blob/main/redis-4.png)

**图3：Bitmap 类型存储内存消耗**

结论：
从内存消耗可以看到，150万个计数器的占用内存为 46.8M - 1M= 45.8M，比直接采用 int 类型存储节约了 61.M 的内存空间；而且随着计数器数量的增加，节省的内存数量会持续增加。

# 计数器实现原理 #
## 内存消耗原理 ##
Redis 存储 int 类型、Bitmap 类型时，对于计数器数量的消耗都是固定的。例如 32位有符号整数，一共需要 4byte 的存储空间；主要节省的空间是 Redis 的键值占用的空间，int 类型的键值与业务数量为 O(N) 的关系，而 Bitmap 类型的键值与业务数量为 O(1) 关系，总节约内存空间 =  (N-1) * 每个键值占用内存 
Redis 主要通过 setbit、getbit 的命令设置 Bitmap 的数据结构，在 Redis 底层中最终存储为字符串的方式

## 数据存储原理 ##
基于内存消耗的分析，现在可以将计数器的数量转换为二进制的方式，并通过 lua 脚本将二进制存储到 Redis 中
每一个业务需要单独设置偏移量，例如 “视频点赞数量”占用 0-31 一共 32位，“视频评论数量”就不能修改0-31位上的内容，否则会引发数据异常，需要占用 32-63位的存储空间，以此类推，每个计数器需要进行偏移量的递增

## 数据读取原理 ##
首先拿到业务的偏移量，将业务对应的二进制读取处理，使用的命令为 getbit video_1 32；注意项：需要将 getbit 命令封装到 lua 脚本中处理
现在有了二进制的内容，通过程序将二进制转换为十进制，就可以还原业务的计数器数量

细心的同学会发现，计数器默认使用 32位进行存储，但是对于一部分业务场景来说，只需要使用 8位、16位的数字；对于这种场景，实现原理：数据存储、数据读取都是按照 8-16位的方式进行存储，这样可以进一步的节约空间，当前仅当 int 类存储不下时，再进行 int 型的升位处理，该过程是不可逆的。

# Go语言实现计数器 #
对于计数器的实现原理，主要需要考虑以下方面的内容，“Redis setbit、gitbit命令操作”、Lua脚本保证数据一致性”、“二进制转十进制”、"十进制转二进制"、支持不同 int 类型的操作方法”

**二进制转十进制代码**
<pre>
  func covertBinToBcd(s []string) (num int) {
	l := len(s)
	for i := l - 1; i >= 0; i-- {
		str := s[l-i-1]
		tmpInt, _ := strconv.Atoi(str)
		num += (tmpInt & 0xf) << uint8(i)
	}
	return
}
</pre>

**十进制转二进制代码**
<pre>
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
</pre>


**Lua脚本保证数据一致性代码**
<pre>
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
</pre>

**支持不同 int 类型的操作方法**
<pre>
type (
	Counter interface {
		SetValue(offset uint8, value int) error
		GetValue(offset uint8) (int, error)
		Incr(offset uint8) (int, error)
		Decr(offset uint8) (int, error)
		IncrCount(offset uint8, count int) (int, error)
		DecrCount(offset uint8, count int) (int, error)
	}

	counterCluster struct {
		r    *redis.Client
		key  string
		bits int
	}
)
</pre>
