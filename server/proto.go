package server

import (
	chaprv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	"github.com/srerickson/ocfl-go"
)

// User is used to convert to/from a protobuf User
type User ocfl.User

func (user User) AsProto() *chaprv1.User {
	return &chaprv1.User{
		Name:    user.Name,
		Address: user.Address,
	}
}

func (user *User) FromProto(proto *chaprv1.User) {
	user.Name = proto.Name
	user.Address = proto.Address
}

func UserFromProto(proto *chaprv1.User) *ocfl.User {
	if proto == nil {
		return nil
	}
	var user User
	user.FromProto(proto)
	return (*ocfl.User)(&user)
}
