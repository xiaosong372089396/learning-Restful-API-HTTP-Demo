package host

import "context"

type Service interface {
	// 录入主机信息
	CreateHost(context.Context, *Host) (*Host, error)
	// 查询主机列表信息
	QueryHost(context.Context, *QueryHostRequest) (*Set, error)
	// 主机详情查询
	DesribeHost(context.Context, *DesribeHostRequest) (*Host, error)
	// 主机信息修改
	UpdateHost(context.Context, *UpdateHostRequest) (*Host, error)
	// 删除主机 GRPC, delete event system
	DeleteHost(context.Context, *DeleteHostRequest) (*Host, error)
}

// 查询数据
type QueryHostRequest struct {
	PageSize   int
	PageNumber int
	Keywords   string
}

func (req *QueryHostRequest) Offset() int {
	return (req.PageNumber - 1) * req.PageSize
}

func NewDescribeHostRequestWithID(id string) *DesribeHostRequest {
	return &DesribeHostRequest{
		Id: id,
	}
}

type DesribeHostRequest struct {
	Id string
}

const (
	PUT   UpdateMode = 0
	PATCH UpdateMode = 1
)

type UpdateMode int

type UpdateHostRequest struct {
	UpdateMode
	*Resource
	*Describe
}

func NewPatchUpdateHostRequest() *UpdateHostRequest {
	return &UpdateHostRequest{
		UpdateMode: PATCH,
		Resource:   &Resource{},
		Describe:   &Describe{},
	}
}

func NewPutUpdateHostRequest() *UpdateHostRequest {
	return &UpdateHostRequest{
		UpdateMode: PUT,
		Resource:   &Resource{},
		Describe:   &Describe{},
	}
}

type DeleteHostRequest struct {
	Id string
}
