package internal

import (
	"encoding/binary"
	"encoding/json"
	"github.com/fengxsong/pubmgmt/api"
)

var (
	Marshal   = json.Marshal
	Unmarshal = json.Unmarshal
)

func MarshalUser(user *pub.User) ([]byte, error) { return Marshal(user) }

func UnmarshalUser(data []byte, user *pub.User) error { return Unmarshal(data, user) }

func Itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
