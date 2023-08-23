package compass

import (
	"encoding/json"
	"fmt"
	"reflect"

	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type AnyMapMessage map[string]interface{}

// provides a wrapper around types.InterfaceRegistry which allows (dese/se)rialization
// of arbitrary cosmos transactions that may use messages not currently known
type AnyInterfaceRegistry struct {
	types.InterfaceRegistry
}

func (am AnyMapMessage) Reset() {
	am = make(map[string]interface{})
}

func (am AnyMapMessage) String() string {
	data, err := json.Marshal(am)
	if err != nil {
		return ""
	}
	return string(data)
}

func (am AnyMapMessage) ProtoMessage() {

}

// do we need this??
func (am AnyMapMessage) Descriptor() ([]byte, []int) {
	return nil, nil
}
func NewAnyInterfaceRegistry(registry types.InterfaceRegistry) types.InterfaceRegistry {
	return &AnyInterfaceRegistry{registry}
}

func (a AnyInterfaceRegistry) Resolve(typeURL string) (proto.Message, error) {
	// if this fails, attempt to default to map[string]interface{}
	msg, err := a.InterfaceRegistry.Resolve(typeURL)
	if err == nil {
		return msg, nil
	}
	msg, ok := reflect.New(reflect.TypeOf(AnyMapMessage{})).Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("fallback resolve failed")
	}
	return msg, nil
}

func (a *AnyInterfaceRegistry) RegisterInterface(protoName string, iface interface{}, impls ...proto.Message) {
	a.InterfaceRegistry.RegisterInterface(protoName, iface, impls...)
}
func (a *AnyInterfaceRegistry) RegisterImplementations(iface interface{}, impls ...proto.Message) {
	a.InterfaceRegistry.RegisterImplementations(iface, impls...)
}
func (a *AnyInterfaceRegistry) ListAllInterfaces() []string {
	return a.InterfaceRegistry.ListAllInterfaces()
}
func (a *AnyInterfaceRegistry) ListImplementations(ifaceTypeURL string) []string {
	return a.InterfaceRegistry.ListImplementations(ifaceTypeURL)
}
func (a *AnyInterfaceRegistry) RangeFiles(f func(protoreflect.FileDescriptor) bool) {
	a.InterfaceRegistry.RangeFiles(f)
}
func (a *AnyInterfaceRegistry) SigningContext() *signing.Context {
	return a.InterfaceRegistry.SigningContext()
}
func (a *AnyInterfaceRegistry) EnsureRegistered(iface interface{}) error {
	return a.InterfaceRegistry.EnsureRegistered(iface)
}
func (a *AnyInterfaceRegistry) FindFileByPath(msg string) (protoreflect.FileDescriptor, error) {
	return a.InterfaceRegistry.FindFileByPath(msg)
}
func (a *AnyInterfaceRegistry) FindDescriptorByName(msg protoreflect.FullName) (protoreflect.Descriptor, error) {
	return a.InterfaceRegistry.FindDescriptorByName(msg)
}
func (a *AnyInterfaceRegistry) UnpackAny(any *types.Any, iface interface{}) error {
	return a.InterfaceRegistry.UnpackAny(any, iface)
}
func (a *AnyInterfaceRegistry) mustEmbedInterfaceRegistry() {
}
