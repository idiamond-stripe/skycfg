// Copyright 2019 The Skycfg Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package dynamicpb is a compatibility shim that allows reading
// dynamicpb messages.
package dynamicpb

import (
	"fmt"
	"log"

	golang_proto "github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	impl "github.com/stripe/skycfg/internal/go/skycfg"
)

type dynamicProtoMessageType struct {
	fileDesc *descriptorpb.FileDescriptorProto
	msgDesc  *descriptorpb.DescriptorProto
	emptyMsg protoiface.MessageV1
}

func (d *dynamicProtoMessageType) Descriptors() (*descriptorpb.FileDescriptorProto, *descriptorpb.DescriptorProto) {
	return d.fileDesc, d.msgDesc
}

func (d *dynamicProtoMessageType) Empty() protoiface.MessageV1 {
	return d.emptyMsg
}

type unstableProtoRegistry interface {
	impl.ProtoRegistry
}

type protoRegistry struct {
	files *protoregistry.Files
}

func (r *protoRegistry) UnstableProtoMessageType(name string) (impl.ProtoMessageType, error) {
	log.Printf("looking up msg: %+v", name)
	descriptor, err := r.files.FindDescriptorByName(protoreflect.FullName(name))
	d, ok := descriptor.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("not a message: %#v", descriptor)
	}
	if err != nil {
		return nil, err
	}
	t := dynamicpb.NewMessageType(d)
	return &dynamicProtoMessageType{
		fileDesc: protodesc.ToFileDescriptorProto(t.Descriptor().ParentFile()),
		msgDesc:  protodesc.ToDescriptorProto(t.Descriptor()),
		emptyMsg: golang_proto.MessageV1(proto.Message(t.New().Interface())),
	}, nil
}

func (*protoRegistry) UnstableEnumValueMap(name string) map[string]int32 {
	log.Printf("looking up enum: %+v", name)
	return nil
}

// ProtoRegistry returns a Protobuf message registry that falls back to GoGo.
//
// To support types that might differ between Protobuf and GoGo registrations,
// the special prefix "gogo:" can be used to skip looking up messages in the
// standard Protobuf registry.
//
//  pb = proto.package("google.protobuf")
//  gogo_pb = proto.package("gogo:google.protobuf")
//  # pb.Timestamp and gogo_pb.Timestamp are distinct types.
//
// The exact type of the return value is not yet stabilized, but the result
// is guaranteed to be accepted by the skycfg.WithProtoRegistry() load option.
func ProtoRegistry(fd *descriptorpb.FileDescriptorSet) (unstableProtoRegistry, error) {
	files, err := protodesc.NewFiles(fd)
	if err != nil {
		return nil, err
	}

	registry := &protoRegistry{
		files: files,
	}

	return registry, nil
}
