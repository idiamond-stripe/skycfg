module github.com/stripe/skycfg

go 1.15

require (
	github.com/emicklei/proto v1.9.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.1
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	go.starlark.net v0.0.0-20190604130855-6ddc71c0ba77
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.2.1
)

replace github.com/kylelemons/godebug => github.com/jmillikin-stripe/godebug v0.0.0-20180620173319-8279e1966bc1
