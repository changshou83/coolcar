package locdesc

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"hash/fnv"

	"google.golang.org/protobuf/proto"
)

var locDesc = []string{
	"中关村",
	"天安门",
	"陆家嘴",
	"迪士尼",
	"天河体育中心",
	"广州塔",
}

type Manager struct{}

func (*Manager) Resolve(c context.Context, loc *rentalpb.Location) (string, error) {
	bytes, err := proto.Marshal(loc)
	if err != nil {
		return "", err
	}

	h := fnv.New32()
	h.Write(bytes)
	return locDesc[int(h.Sum32())%len(locDesc)], nil
}
