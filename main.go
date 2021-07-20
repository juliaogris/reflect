package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	dpb "google.golang.org/protobuf/types/descriptorpb"
)

var version = "v0.0.0"

type globals struct {
	Address   string `short:"a" env:"GRPC_ADDRESS" help:"gRPC server address"`
	Plaintext bool   `short:"p" help:"Use plain-text; no TLS" env:"GRPC_PLAINTEXT"`
	Format    string `short:"f" help:"output protoset as one of json, base64, bin, text" enum:"json,base64,bin,text" default:"json"`
	Out       string `short:"o" help:"output file, default: stdout" default:"-"`

	out         io.Writer
	hostAddress string // used in tests to work with localhost:0
}

type config struct {
	globals
	Version    kong.VersionFlag `short:"V" help:"Print version information" group:"Other:"`
	Services   servicesCmd      `cmd:"" help:"Call list_services"`
	Symbol     symbolCmd        `cmd:"" help:"Call file_containing_symbol"`
	Filename   filenameCmd      `cmd:"" help:"Call file_by_filename"`
	Extension  extensionCmd     `cmd:"" help:"Call file_containing_extension"`
	Extensions extensionsCmd    `cmd:"" help:"Call all_extension_numbers_of_type"`
	FDS        fdsCmd           `cmd:"" help:"Decode base64 encoded FileDescriptorSet"`
	FD         fdCmd            `cmd:"" help:"Decode base64 encoded FileDescriptor"`
	FDSF       fdsfCmd          `cmd:"" help:"Decode proto encoded FileDescriptorSet.pb file"`
}

type servicesCmd struct{}

type symbolCmd struct {
	Symbol string `arg:""`
}

type filenameCmd struct {
	Filename string `arg:""`
}

type extensionCmd struct {
	Type   string `arg:""`
	Number int32  `arg:""`
}

type extensionsCmd struct {
	Type string `arg:""`
}

type fdCmd struct {
	FileDescriptor string `arg:""`
}

type fdsCmd struct {
	FileDescriptorSet string `arg:""`
}

type fdsfCmd struct {
	FileDescriptorSetFile kong.FileContentFlag `arg:""`
}

func main() {
	cfg := &config{}
	kctx := kong.Parse(cfg,
		kong.Vars{"version": version},
		kong.Description("gRPC reflection API toolkit"),
	)
	err := kctx.Run(cfg.globals)
	kctx.FatalIfErrorf(err)
}

func (cfg *config) AfterApply() error {
	cfg.out = os.Stdout
	if cfg.Out != "-" {
		var err error
		if cfg.out, err = os.Create(cfg.Out); err != nil {
			return errors.Wrap(err, "cannot create output file")
		}
	}
	cfg.hostAddress = cfg.Address
	return nil
}

func (f *fdCmd) Run(g globals) error {
	m := &dpb.FileDescriptorProto{}
	return decode(f.FileDescriptor, m, g)
}

func (f *fdsCmd) Run(g globals) error {
	m := &dpb.FileDescriptorSet{}
	return decode(f.FileDescriptorSet, m, g)
}

func (f *fdsfCmd) Run(g globals) error {
	m := &dpb.FileDescriptorSet{}
	err := proto.Unmarshal(f.FileDescriptorSetFile, m)
	if err != nil {
		return errors.Wrap(err, "cannot decode proto message")
	}
	return printProto(g.out, m, g.Format)
}

func decode(b64 string, m protoreflect.ProtoMessage, g globals) error {
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return errors.Wrap(err, "cannot decode b64 string")
	}
	err = proto.Unmarshal(b, m)
	if err != nil {
		return errors.Wrap(err, "cannot decode proto message")
	}
	return printProto(g.out, m, g.Format)
}

func (s *servicesCmd) Run(g globals) error {
	req := &rpb.ServerReflectionRequest{
		Host:           g.hostAddress,
		MessageRequest: &rpb.ServerReflectionRequest_ListServices{},
	}
	return run(req, g)
}

func (s *symbolCmd) Run(g globals) error {
	req := &rpb.ServerReflectionRequest{
		Host: g.hostAddress,
		MessageRequest: &rpb.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: s.Symbol,
		},
	}
	return run(req, g)
}

func (f *filenameCmd) Run(g globals) error {
	req := &rpb.ServerReflectionRequest{
		Host: g.hostAddress,
		MessageRequest: &rpb.ServerReflectionRequest_FileByFilename{
			FileByFilename: f.Filename,
		},
	}
	return run(req, g)
}

func (e *extensionCmd) Run(g globals) error {
	req := &rpb.ServerReflectionRequest{
		Host: g.hostAddress,
		MessageRequest: &rpb.ServerReflectionRequest_FileContainingExtension{
			FileContainingExtension: &rpb.ExtensionRequest{
				ContainingType:  e.Type,
				ExtensionNumber: e.Number,
			},
		},
	}
	return run(req, g)
}

func (e *extensionsCmd) Run(g globals) error {
	req := &rpb.ServerReflectionRequest{
		Host: g.hostAddress,
		MessageRequest: &rpb.ServerReflectionRequest_AllExtensionNumbersOfType{
			AllExtensionNumbersOfType: e.Type,
		},
	}
	return run(req, g)
}

func run(req *rpb.ServerReflectionRequest, g globals) error {
	stream, err := newStream(context.Background(), g)
	if err != nil {
		return err
	}
	defer closeAndDrain(stream)
	resp, err := send(stream, req)
	if err != nil {
		return err
	}

	return printProto(g.out, resp, g.Format)
}

func closeAndDrain(stream rpb.ServerReflection_ServerReflectionInfoClient) {
	_ = stream.CloseSend()
	for {
		if _, err := stream.Recv(); err != nil {
			return
		}
	}
}

func newStream(ctx context.Context, g globals) (rpb.ServerReflection_ServerReflectionInfoClient, error) {
	var opts []grpc.DialOption
	if g.Plaintext {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(g.Address, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot grpc dial %s", g.Address)
	}
	client := rpb.NewServerReflectionClient(conn)
	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot setup reflection stream")
	}
	return stream, nil
}

func send(stream rpb.ServerReflection_ServerReflectionInfoClient, req *rpb.ServerReflectionRequest) (*rpb.ServerReflectionResponse, error) {
	if err := stream.Send(req); err != nil {
		return nil, errors.Wrap(err, "cannot send reflection request")
	}
	resp, err := stream.Recv()
	if err != nil {
		return nil, errors.Wrap(err, "cannot receive reflection response")
	}
	return resp, nil
}

func printProto(w io.Writer, m protoreflect.ProtoMessage, format string) error {
	var b []byte
	var err error
	switch format {
	case "json":
		b, err = jsonString(m)
	case "base64":
		b, err = base64String(m)
	case "text":
		b, err = textString(m)
	case "bin":
		b, err = binString(m)
	default:
		err = fmt.Errorf("unknown format %s", format)
	}
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return errors.Wrap(err, "cannot print proto")
}

func jsonString(m protoreflect.ProtoMessage) ([]byte, error) {
	marshaler := protojson.MarshalOptions{Multiline: true}
	out, err := marshaler.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "cannot jsonString")
	}
	return out, nil
}

func base64String(m protoreflect.ProtoMessage) ([]byte, error) {
	b, err := proto.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "cannot base64String")
	}
	return []byte(base64.StdEncoding.EncodeToString(b)), nil
}

func binString(m protoreflect.ProtoMessage) ([]byte, error) {
	out, err := proto.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "cannot binString")
	}
	return out, nil
}

func textString(m protoreflect.ProtoMessage) ([]byte, error) {
	out, err := prototext.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "cannot textString")
	}
	return out, nil
}
