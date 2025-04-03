package snowflake

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	// Epoch 是 Twitter 的 Snowflake 时间戳初始时间（毫秒），设为 2010 年 11 月 4 日 01:42:54 UTC
	// 根据需要自定义初始时间。
	Epoch int64 = 1288834974657

	// NodeBits 表示用于节点（Node）的位数
	// 总共 22 位可以分配给节点和步数（Step）
	NodeBits uint8 = 10

	// StepBits 表示用于步数（Step）的位数
	// 总共 22 位可以分配给节点和步数
	StepBits uint8 = 12

	mu        sync.Mutex
	nodeMax   int64 = -1 ^ (-1 << NodeBits)
	nodeMask        = nodeMax << StepBits
	stepMask  int64 = -1 ^ (-1 << StepBits)
	timeShift       = NodeBits + StepBits
	nodeShift       = StepBits
)

// JSONSyntaxError 是在解析 JSON 时遇到无效 ID 时返回的错误类型。
type JSONSyntaxError struct{ original []byte }

func (j JSONSyntaxError) Error() string {
	return fmt.Sprintf("无效的 Snowflake ID %q", string(j.original))
}

// Node 结构体包含生成 Snowflake ID 所需的基本信息
type Node struct {
	mu    sync.Mutex
	epoch time.Time
	time  int64
	node  int64
	step  int64

	nodeMax   int64
	nodeMask  int64
	stepMask  int64
	timeShift uint8
	nodeShift uint8
}

// ID 是用于 Snowflake ID 的自定义类型，用于附加方法
type ID int64

// NewNode 返回一个可以用来生成 Snowflake ID 的新节点
func NewNode(node int64) (*Node, error) {

	mu.Lock()
	nodeMax = -1 ^ (-1 << NodeBits)
	nodeMask = nodeMax << StepBits
	stepMask = -1 ^ (-1 << StepBits)
	timeShift = NodeBits + StepBits
	nodeShift = StepBits
	mu.Unlock()

	n := Node{}
	n.node = node
	n.nodeMax = -1 ^ (-1 << NodeBits)
	n.nodeMask = n.nodeMax << StepBits
	n.stepMask = -1 ^ (-1 << StepBits)
	n.timeShift = NodeBits + StepBits
	n.nodeShift = StepBits

	if n.node < 0 || n.node > n.nodeMax {
		return nil, errors.New("节点号必须在 0 到 " + strconv.FormatInt(n.nodeMax, 10) + " 之间")
	}

	var curTime = time.Now()
	// 加入 time.Duration，以确保使用单调时钟（若可用）
	n.epoch = curTime.Add(time.Unix(Epoch/1000, (Epoch%1000)*1000000).Sub(curTime))

	return &n, nil
}

// Generate 创建并返回一个唯一的 Snowflake ID
// 确保唯一性：
// - 系统时间准确
// - 不会有多个节点使用相同的节点 ID
func (n *Node) Generate() ID {

	n.mu.Lock()

	now := time.Since(n.epoch).Nanoseconds() / 1000000

	if now == n.time {
		n.step = (n.step + 1) & n.stepMask

		if n.step == 0 {
			for now <= n.time {
				now = time.Since(n.epoch).Nanoseconds() / 1000000
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	r := ID((now)<<n.timeShift |
		(n.node << n.nodeShift) |
		(n.step),
	)

	n.mu.Unlock()
	return r
}

// Int64 返回 Snowflake ID 的 int64 表示
func (f ID) Int64() int64 {
	return int64(f)
}

// Get12Id 生成 12 位整数形式的 ID
func GetInt12Id(node *Node) int64 {
	id := node.Generate().Int64()
	// 获取高位时间戳部分
	timestampPart := id >> 22           // 提取时间戳高位（42 位）
	timestampPart = timestampPart % 1e6 // 取时间戳后 6 位
	// 获取低位节点和步数部分
	lowPart := id & ((1 << 22) - 1) // 取低 22 位
	lowPart = lowPart % 1e6         // 取低位后 6 位
	// 合并高位和低位部分，组成 12 位整数
	result := timestampPart*1e6 + lowPart
	// 如果超出 12 位，则取模确保是 12 位整数
	if result >= 1e12 {
		result = result % 1e12
	}
	return result
}

// String 返回 Snowflake ID 的字符串表示
func (f ID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// GetString12Id 生成 12 位字符串形式的 ID
func GetString12Id(node *Node) string {
	id := node.Generate().Int64()
	timestampPart := id >> 22
	timestampPart = timestampPart % 1e6
	lowPart := id & ((1 << 22) - 1)
	lowPart = lowPart % 1e6
	// 合并高低部分，组成 12 位字符串
	result := fmt.Sprintf("%06d%06d", timestampPart, lowPart)
	// 如果超出 12 位，则截取前 12 位
	if len(result) > 12 {
		result = result[:12]
	}
	return result
}
