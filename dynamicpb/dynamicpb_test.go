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
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/golang/protobuf/descriptor"
	golang_proto "github.com/golang/protobuf/proto"
	"github.com/stripe/skycfg"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	_ "google.golang.org/protobuf/types/dynamicpb"
)

const (
	descriptorFile = "testdata/descriptor_set.bin"
	configFile     = "testdata/config.cfg"
)

// Copied from internal/go/skycfg/proto_util.go
func mustParseFileDescriptor(gzBytes []byte) *descriptorpb.FileDescriptorProto {
	gz, err := gzip.NewReader(bytes.NewReader(gzBytes))
	if err != nil {
		panic(fmt.Sprintf("EnumDescriptor: %v", err))
	}
	defer gz.Close()

	fileDescBytes, err := ioutil.ReadAll(gz)
	if err != nil {
		panic(fmt.Sprintf("EnumDescriptor: %v", err))
	}

	fileDesc := &descriptorpb.FileDescriptorProto{}
	if err := proto.Unmarshal(fileDescBytes, fileDesc); err != nil {
		panic(fmt.Sprintf("EnumDescriptor: %v", err))
	}
	return fileDesc
}

// Copied from internal/go/skycfg/proto_util.go
func messageTypeName(msg golang_proto.Message) string {
	if hasName, ok := msg.(interface {
		XXX_MessageName() string
	}); ok {
		return hasName.XXX_MessageName()
	}

	hasDesc, ok := msg.(descriptor.Message)
	if !ok {
		return golang_proto.MessageName(msg)
	}

	gzBytes, path := hasDesc.Descriptor()
	fileDesc := mustParseFileDescriptor(gzBytes)
	var chunks []string
	if pkg := fileDesc.GetPackage(); pkg != "" {
		chunks = append(chunks, pkg)
	}

	msgDesc := fileDesc.MessageType[path[0]]
	for ii := 1; ii < len(path); ii++ {
		chunks = append(chunks, msgDesc.GetName())
		msgDesc = msgDesc.NestedType[path[ii]]
	}
	chunks = append(chunks, msgDesc.GetName())
	return strings.Join(chunks, ".")
}

func registry(t *testing.T) unstableProtoRegistry {
	var fileDescriptorSet descriptorpb.FileDescriptorSet
	file, err := os.Open(descriptorFile)
	if err != nil {
		t.Fatalf("could not open %q: %+v", descriptorFile, err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("could not read file %q: %+v", descriptorFile, err)
	}

	if err := proto.Unmarshal(data, &fileDescriptorSet); err != nil {
		t.Fatalf("could not unmarshal descriptor: %+v", err)
	}

	registry, err := ProtoRegistry(&fileDescriptorSet)
	if err != nil {
		t.Fatalf("could not create registry: %+v", err)
	}

	return registry
}

func TestRegistry(t *testing.T) {
	registry := registry(t)

	mt, err := registry.UnstableProtoMessageType("testdata.AddressBook")
	if err != nil {
		t.Fatalf("could not find proto: %+v", err)
	}
	msg := mt.Empty()
	name := messageTypeName(msg)
	if name != "testdata.AddressBook" {
		t.Fatalf("unexpected name: %+v", name)
	}
}

func TestIntegration(t *testing.T) {
	registry := registry(t)

	config, err := skycfg.Load(context.Background(), configFile, skycfg.WithProtoRegistry(registry))
	if err != nil {
		t.Fatalf("could not load config: %+v", err)
	}
	protos, err := config.Main(context.Background())
	if err != nil {
		t.Fatalf("error evaluating %q: %v\n", config.Filename(), err)
		os.Exit(1)
	}
	log.Printf("%+v", protos)
}
