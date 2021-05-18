package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/juliaogris/reflect/pkg/echo2"
	"github.com/juliaogris/reflect/pkg/echo3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestReflectSuite(t *testing.T) {
	suite.Run(t, &ReflectSuite{format: "json", pbVersion: 3})
	suite.Run(t, &ReflectSuite{format: "base64", pbVersion: 3})
	suite.Run(t, &ReflectSuite{format: "bin", pbVersion: 3})
	suite.Run(t, &ReflectSuite{format: "json", pbVersion: 2})
	suite.Run(t, &ReflectSuite{format: "base64", pbVersion: 2})
	suite.Run(t, &ReflectSuite{format: "bin", pbVersion: 3})
}

type ReflectSuite struct {
	suite.Suite
	format    string
	pbVersion int

	server  *grpc.Server
	globals globals
	subDir  string
}

func (s *ReflectSuite) SetupSuite() {
	t := s.T()
	s.server = grpc.NewServer()
	switch s.pbVersion {
	case 2:
		echo2Server := &echo2.Server{}
		echo2.RegisterEchoServer(s.server, echo2Server)
	case 3:
		echo3Server := &echo3.Server{}
		echo3.RegisterEchoServer(s.server, echo3Server)
	default:
		require.Fail(t, "unknown proto version")
	}
	reflection.Register(s.server)
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := lis.Addr().(*net.TCPAddr).Port
	serve := func() {
		if err := s.server.Serve(lis); err != nil {
			panic(err)
		}
	}
	s.globals = globals{
		Address:     fmt.Sprintf("localhost:%d", port),
		Plaintext:   true,
		Format:      s.format,
		hostAddress: "localhost:0",
	}
	s.subDir = fmt.Sprintf("proto%d-%s", s.pbVersion, s.format)
	go serve()
}

func (s *ReflectSuite) TearDownSuite() {
	s.server.Stop()
}

func (s *ReflectSuite) TestServicesCmd() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := servicesCmd{}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
}

func (s *ReflectSuite) TestSymbolCmd() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := symbolCmd{Symbol: "echo3.Echo"}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
}

func (s *ReflectSuite) TestSymbolCmdErr() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := symbolCmd{Symbol: "MISSING"}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
}

func (s *ReflectSuite) TestFilenameCmd() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := filenameCmd{Filename: "echo/echo.proto"}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
}

func (s *ReflectSuite) TestFilenameCmdErr() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := filenameCmd{Filename: "MISSING"}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
}

func (s *ReflectSuite) TestExtensionCmd() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := extensionCmd{
		Type:   "google.protobuf.MethodOptions",
		Number: 72295728,
	}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
}

func (s *ReflectSuite) TestExtensionCmdErr() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := extensionCmd{
		Type:   "MISSING",
		Number: 72295728,
	}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
}

func (s *ReflectSuite) TestExtensionsCmd() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := extensionsCmd{Type: "google.protobuf.MethodOptions"}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	// undetermined order of extension numbers in response call don't allow for golden comparison
	// requireContentEq(t, f.want, f.got, s.format)
	wantResp := reflectionResponse(t, f.want, s.format)
	gotResp := reflectionResponse(t, f.got, s.format)

	want := wantResp.GetAllExtensionNumbersResponse().GetExtensionNumber()
	got := gotResp.GetAllExtensionNumbersResponse().GetExtensionNumber()

	require.Greater(t, len(got), 0)
	require.ElementsMatch(t, want, got)
}

func (s *ReflectSuite) TestExtensionsCmdErr() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := extensionsCmd{Type: "MISSING"}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
}

func requireContentEq(t *testing.T, fname1, fname2, format string) {
	t.Helper()
	f1, err := os.ReadFile(fname1)
	require.NoError(t, err)
	f2, err := os.ReadFile(fname2)
	require.NoError(t, err)
	if format == "json" {
		// protojson adds random whitespace to avoid byte-by-byte comparison
		require.JSONEq(t, string(f1), string(f2))
	} else {
		require.Equal(t, f1, f2)
	}
}

func reflectionResponse(t *testing.T, fname string, format string) *rpb.ServerReflectionResponse {
	t.Helper()
	b, err := os.ReadFile(fname)
	require.NoError(t, err)
	r := rpb.ServerReflectionResponse{}
	switch format {
	case "json":
		err = protojson.Unmarshal(b, &r)
	case "base64":
		b, err = base64.StdEncoding.DecodeString(string(b))
		require.NoError(t, err)
		err = proto.Unmarshal(b, &r)
	case "bin":
		err = proto.Unmarshal(b, &r)
	default:
		err = fmt.Errorf("unknown format type")
	}
	require.NoError(t, err)
	return &r
}

type outputFiles struct {
	got  string
	want string

	out io.Writer
}

func files(t *testing.T, format, subDir string) outputFiles {
	t.Helper()
	name := strings.Split(path.Base(t.Name()), "#")[0] + "." + format

	f := outputFiles{
		// To rewrite or compare golden files use:
		// got: path.Join("testdata", subDir, strings.Split(path.Base(t.Name()), "#")[0]+"."+format),
		got:  path.Join(t.TempDir(), name),
		want: path.Join("testdata", subDir, name),
	}
	var err error
	f.out, err = os.Create(f.got)
	require.NoError(t, err)
	return f
}
