package host

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/imdario/mergo"
)

var (
	validate = validator.New()
)

func NewDefaultHost() *Host {
	return &Host{
		Resource: &Resource{
			CreateAt: time.Now().UnixNano() / 1000000,
		},
		Describe: &Describe{},
	}
}

// 为了后期做资源解索， <ip> --> host, eip, slb, redis, mysql
type Host struct {
	ResourceHash string `json:"resource_hash"`
	DescribeHash string `json:"describe_hash"`
	*Resource
	*Describe
}

func (h *Host) Validate() error {
	return validate.Struct(h)
}

func (h *Host) Patch(res *Resource, desc *Describe) error {
	h.UpdateAt = time.Now().UnixNano() / 1000000

	//
	if res != nil {
		err := mergo.MergeWithOverwrite(h.Resource, res)
		if err != nil {
			return err
		}
	}

	if desc != nil {
		err := mergo.MergeWithOverwrite(h.Describe, desc)
		if err != nil {
			return err
		}
	}
	return nil
}

// go 1.17 允许获取毫秒了
func (h *Host) Update(res *Resource, desc *Describe) {
	h.UpdateAt = time.Now().UnixNano() / 1000000
	h.Resource = res
	h.Describe = desc
}

type Vendor int

const (
	ALI_CLOUD Vendor = iota
	TX_CLOUD
	HW_CLOUD
)

// 主机的元数据信息, Region 创建时间
type Resource struct {
	Id     string `json:"id"  validate:"required"`     // 全局唯一Id
	Vendor Vendor `json:"vendor"`                      // 厂商
	Region string `json:"region"  validate:"required"` // 地域
	Zone   string `json:"zone"`                        // 区域
	// 使用13位的时间戳
	// 为什么不只用Datetime，如果使用数据库的时间, 数据库会给你默认加上时区
	CreateAt    int64             `json:"create_at"`                 // 创建时间
	ExpireAt    int64             `json:"expire_at"`                 // 过期时间
	Category    string            `json:"category"`                  // 种类
	Type        string            `json:"type"  validate:"required"` // 规格
	InstanceId  string            `json:"instance_id"`               // 实例id
	Name        string            `json:"name"  validate:"required"` // 名称
	Description string            `json:"description"`               // 描述
	Status      string            `json:"status"`                    // 服务商中的状态
	Tags        map[string]string `json:"tags"`                      // 标签
	UpdateAt    int64             `json:"update_at"`                 // 更新时间
	SyncAt      int64             `json:"sync_at"`                   // 同步时间
	SyncAccount string            `json:"sync_account"`              // 同步的账号
	// Account     string            `json:"accout"`                      // 资源的所属账号
	PublicIP  string `json:"public_ip"`  // 公网IP
	PrivateIP string `json:"private_ip"` // 内网IP
	PayType   string `json:"pay_type"`   // 实例付费方式
}

// 主机的具体信息
type Describe struct {
	// ResourceId              string `json:"resource_id"`                // 关联Resource
	CPU                     int    `json:"cpu" validate:"required"`    // 核数
	Memory                  int    `json:"memory" validate:"required"` // 内存
	GPUAmount               int    `json:"gpu_amount"`                 // GPU数量
	GPUSpec                 string `json:"gpu_spec"`                   // GPU类型
	OSType                  string `json:"os_type"`                    // 操作系统类型，分为Windows和Linux
	OSName                  string `json:"os_name"`                    // 操作系统名称
	SerialNumber            string `json:"serial_number"`              // 序列号
	ImageID                 string `json:"image_id"`                   // 镜像ID
	InternetMaxBandwidthOut int    `json:"internet_max_bandwidth_out"` // 公网出带宽最大值, 单位Mbps
	InternetMaxBandwidthIn  int    `json:"internet_max_bandwidth_in"`  // 公网入带宽最大值, 单位Mbps
	KeyPairName             string `json:"key_pair_name"`              // 密钥对名称
	SecurityGroups          string `json:"security_groups"`            // 安全组, 采用逗号分隔
}

// 分页查询响应数据
type Set struct {
	Total int64   `json:"total"`
	Items []*Host `json:"items"`
}

func (s *Set) Add(item *Host) {
	s.Items = append(s.Items, item)
}

func NewSet() *Set {
	return &Set{
		Items: []*Host{},
	}
}
