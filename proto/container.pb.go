// Code generated by protoc-gen-go. DO NOT EDIT.
// source: container.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type NetworkSpec struct {
	Type    string            `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	Options map[string]string `protobuf:"bytes,2,rep,name=options" json:"options,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Subnet  string            `protobuf:"bytes,3,opt,name=subnet" json:"subnet,omitempty"`
	Addr    string            `protobuf:"bytes,4,opt,name=addr" json:"addr,omitempty"`
}

func (m *NetworkSpec) Reset()                    { *m = NetworkSpec{} }
func (m *NetworkSpec) String() string            { return proto.CompactTextString(m) }
func (*NetworkSpec) ProtoMessage()               {}
func (*NetworkSpec) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

func (m *NetworkSpec) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *NetworkSpec) GetOptions() map[string]string {
	if m != nil {
		return m.Options
	}
	return nil
}

func (m *NetworkSpec) GetSubnet() string {
	if m != nil {
		return m.Subnet
	}
	return ""
}

func (m *NetworkSpec) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

type Container struct {
	// Image describes a Docker image name. Required.
	Image string `protobuf:"bytes,1,opt,name=image" json:"image,omitempty"`
	// Registry describes Docker registry.
	Registry string `protobuf:"bytes,2,opt,name=registry" json:"registry,omitempty"`
	// Auth describes authentication info used for registry.
	Auth string `protobuf:"bytes,3,opt,name=auth" json:"auth,omitempty"`
	// SSH public key used to attach to the container.
	PublicKeyData string `protobuf:"bytes,4,opt,name=publicKeyData" json:"publicKeyData,omitempty"`
	// CommitOnStop points whether a container should commit when stopped.
	// Committed containers can be fetched later while there is an active
	// deal.
	CommitOnStop bool `protobuf:"varint,5,opt,name=commitOnStop" json:"commitOnStop,omitempty"`
	// Env describes environment variables forwarded into the container.
	Env map[string]string `protobuf:"bytes,7,rep,name=env" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// Volumes describes network volumes that are used to be mounted inside
	// the container.
	// Mapping from the volume type (cifs, nfs, etc.) to its settings.
	Volumes map[string]*Volume `protobuf:"bytes,8,rep,name=volumes" json:"volumes,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// Mounts describes mount points from the volume name to the container.
	// TODO: Dragons nearby - beware of injection attacks.
	Mounts   []string       `protobuf:"bytes,9,rep,name=mounts" json:"mounts,omitempty"`
	Networks []*NetworkSpec `protobuf:"bytes,10,rep,name=networks" json:"networks,omitempty"`
}

func (m *Container) Reset()                    { *m = Container{} }
func (m *Container) String() string            { return proto.CompactTextString(m) }
func (*Container) ProtoMessage()               {}
func (*Container) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{1} }

func (m *Container) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *Container) GetRegistry() string {
	if m != nil {
		return m.Registry
	}
	return ""
}

func (m *Container) GetAuth() string {
	if m != nil {
		return m.Auth
	}
	return ""
}

func (m *Container) GetPublicKeyData() string {
	if m != nil {
		return m.PublicKeyData
	}
	return ""
}

func (m *Container) GetCommitOnStop() bool {
	if m != nil {
		return m.CommitOnStop
	}
	return false
}

func (m *Container) GetEnv() map[string]string {
	if m != nil {
		return m.Env
	}
	return nil
}

func (m *Container) GetVolumes() map[string]*Volume {
	if m != nil {
		return m.Volumes
	}
	return nil
}

func (m *Container) GetMounts() []string {
	if m != nil {
		return m.Mounts
	}
	return nil
}

func (m *Container) GetNetworks() []*NetworkSpec {
	if m != nil {
		return m.Networks
	}
	return nil
}

func init() {
	proto.RegisterType((*NetworkSpec)(nil), "sonm.NetworkSpec")
	proto.RegisterType((*Container)(nil), "sonm.Container")
}

func init() { proto.RegisterFile("container.proto", fileDescriptor4) }

var fileDescriptor4 = []byte{
	// 374 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x4d, 0x6b, 0xe3, 0x30,
	0x10, 0xc5, 0x71, 0x3e, 0x9c, 0x89, 0x97, 0xdd, 0x15, 0xcb, 0x22, 0x4c, 0x29, 0xc6, 0xf4, 0x10,
	0x0a, 0xf5, 0x21, 0x85, 0x10, 0x72, 0x6d, 0x03, 0x85, 0x42, 0x03, 0x0e, 0xf4, 0xee, 0x38, 0x22,
	0x35, 0x89, 0x25, 0x23, 0xc9, 0x2e, 0xfe, 0x7d, 0xbd, 0xf4, 0x67, 0x15, 0x7d, 0x38, 0x38, 0x6d,
	0x2f, 0xbd, 0xcd, 0x1b, 0xbf, 0x79, 0x7a, 0xf3, 0xc6, 0xf0, 0x3b, 0x63, 0x54, 0xa6, 0x39, 0x25,
	0x3c, 0x2e, 0x39, 0x93, 0x0c, 0xf5, 0x05, 0xa3, 0x45, 0xe0, 0xd7, 0xec, 0x58, 0x15, 0xc4, 0xf4,
	0xa2, 0x37, 0x07, 0x26, 0x4f, 0x44, 0xbe, 0x32, 0x7e, 0xd8, 0x94, 0x24, 0x43, 0x08, 0xfa, 0xb2,
	0x29, 0x09, 0x76, 0x42, 0x67, 0x3a, 0x4e, 0x74, 0x8d, 0x16, 0x30, 0x62, 0xa5, 0xcc, 0x19, 0x15,
	0xb8, 0x17, 0xba, 0xd3, 0xc9, 0xec, 0x32, 0x56, 0x4a, 0x71, 0x67, 0x2e, 0x5e, 0x1b, 0xc2, 0x8a,
	0x4a, 0xde, 0x24, 0x2d, 0x1d, 0xfd, 0x87, 0xa1, 0xa8, 0xb6, 0x94, 0x48, 0xec, 0x6a, 0x3d, 0x8b,
	0xd4, 0x2b, 0xe9, 0x6e, 0xc7, 0x71, 0xdf, 0xbc, 0xa2, 0xea, 0x60, 0x09, 0x7e, 0x57, 0x04, 0xfd,
	0x01, 0xf7, 0x40, 0x1a, 0x6b, 0x44, 0x95, 0xe8, 0x1f, 0x0c, 0xea, 0xf4, 0x58, 0x11, 0xdc, 0xd3,
	0x3d, 0x03, 0x96, 0xbd, 0x85, 0x13, 0xbd, 0xbb, 0x30, 0xbe, 0x6b, 0xb7, 0x55, 0xbc, 0xbc, 0x48,
	0xf7, 0xed, 0x12, 0x06, 0xa0, 0x00, 0x3c, 0x4e, 0xf6, 0xb9, 0x90, 0xbc, 0xb1, 0x02, 0x27, 0xac,
	0xfd, 0x54, 0xf2, 0xc5, 0xba, 0xd4, 0x35, 0xba, 0x82, 0x5f, 0x65, 0xb5, 0x3d, 0xe6, 0xd9, 0x23,
	0x69, 0xee, 0x53, 0x99, 0x5a, 0xb3, 0xe7, 0x4d, 0x14, 0x81, 0x9f, 0xb1, 0xa2, 0xc8, 0xe5, 0x9a,
	0x6e, 0x24, 0x2b, 0xf1, 0x20, 0x74, 0xa6, 0x5e, 0x72, 0xd6, 0x43, 0xd7, 0xe0, 0x12, 0x5a, 0xe3,
	0x91, 0xce, 0x0e, 0x9b, 0xec, 0x4e, 0x6e, 0xe3, 0x15, 0xad, 0x4d, 0x6a, 0x8a, 0x84, 0xe6, 0x30,
	0x32, 0xf7, 0x11, 0xd8, 0xd3, 0xfc, 0x8b, 0xcf, 0xfc, 0x67, 0xf3, 0xd9, 0x26, 0x6d, 0xc9, 0x2a,
	0xe9, 0x82, 0x55, 0x54, 0x0a, 0x3c, 0x0e, 0x5d, 0x95, 0xb4, 0x41, 0xe8, 0x06, 0x3c, 0x6a, 0xce,
	0x24, 0x30, 0x68, 0xc1, 0xbf, 0x5f, 0x8e, 0x97, 0x9c, 0x28, 0xc1, 0x1c, 0xbc, 0xd6, 0xcf, 0x4f,
	0x0e, 0x10, 0x3c, 0x80, 0xdf, 0xf5, 0xf5, 0xcd, 0x6c, 0xd4, 0x9d, 0x9d, 0xcc, 0x7c, 0xe3, 0xc2,
	0x0c, 0x75, 0x94, 0xb6, 0x43, 0xfd, 0x5f, 0xde, 0x7e, 0x04, 0x00, 0x00, 0xff, 0xff, 0xcf, 0x3f,
	0xbf, 0x2d, 0xbe, 0x02, 0x00, 0x00,
}
