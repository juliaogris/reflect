package main

import (
	"bytes"
	_ "embed"
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

func TestFileDescriptorCmd(t *testing.T) {
	b64FD := "Chlnb29nbGUvcHJvdG9idWYvYW55LnByb3RvEg9nb29nbGUucHJvdG9idWYiNgoDQW55EhkKCHR5cGVfdXJsGAEgASgJUgd0eXBlVXJsEhQKBXZhbHVlGAIgASgMUgV2YWx1ZUJ2ChNjb20uZ29vZ2xlLnByb3RvYnVmQghBbnlQcm90b1ABWixnb29nbGUuZ29sYW5nLm9yZy9wcm90b2J1Zi90eXBlcy9rbm93bi9hbnlwYqICA0dQQqoCHkdvb2dsZS5Qcm90b2J1Zi5XZWxsS25vd25UeXBlc2IGcHJvdG8z"
	cmd := fdCmd{FileDescriptor: b64FD}
	b := &bytes.Buffer{}
	g := globals{
		Format: "json",
		out:    b,
	}
	err := cmd.Run(g)
	require.NoError(t, err)
	want := `{
  "name": "google/protobuf/any.proto",
  "package": "google.protobuf",
  "messageType": [
    {
      "name": "Any",
      "field": [
        { "name": "type_url", "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "jsonName": "typeUrl" },
        { "name": "value",    "number": 2, "label": "LABEL_OPTIONAL", "type": "TYPE_BYTES",  "jsonName": "value" }
      ]
    }
  ],
  "options": {
    "javaPackage": "com.google.protobuf",
    "javaOuterClassname": "AnyProto",
    "javaMultipleFiles": true,
    "goPackage": "google.golang.org/protobuf/types/known/anypb",
    "objcClassPrefix": "GPB",
    "csharpNamespace": "Google.Protobuf.WellKnownTypes"
  },
  "syntax": "proto3"
}
`
	require.JSONEq(t, want, b.String())
}

func TestFileDescriptorSetCmd(t *testing.T) {
	b64FDS := "CvkBCg9lY2hvL2VjaG8ucHJvdG8SBGVjaG8iKAoMSGVsbG9SZXF1ZXN0EhgKB21lc3NhZ2UYASABKAlSB21lc3NhZ2UiKwoNSGVsbG9SZXNwb25zZRIaCghyZXNwb25zZRgBIAEoCVIIcmVzcG9uc2UyWQoERWNobxJRCgVIZWxsbxISLmVjaG8uSGVsbG9SZXF1ZXN0GhMuZWNoby5IZWxsb1Jlc3BvbnNlIh+C0+STAhk6ASoiFC9hcGkvZWNoby5FY2hvL0hlbGxvQiZaJGdpdGh1Yi5jb20vanVsaWFvZ3Jpcy9ndXBweS9wa2cvZWNob2IGcHJvdG8zCrcOCjNyZWZsZWN0aW9uL2dycGNfcmVmbGVjdGlvbl92MWFscGhhL3JlZmxlY3Rpb24ucHJvdG8SF2dycGMucmVmbGVjdGlvbi52MWFscGhhIvgCChdTZXJ2ZXJSZWZsZWN0aW9uUmVxdWVzdBISCgRob3N0GAEgASgJUgRob3N0EioKEGZpbGVfYnlfZmlsZW5hbWUYAyABKAlIAFIOZmlsZUJ5RmlsZW5hbWUSNgoWZmlsZV9jb250YWluaW5nX3N5bWJvbBgEIAEoCUgAUhRmaWxlQ29udGFpbmluZ1N5bWJvbBJnChlmaWxlX2NvbnRhaW5pbmdfZXh0ZW5zaW9uGAUgASgLMikuZ3JwYy5yZWZsZWN0aW9uLnYxYWxwaGEuRXh0ZW5zaW9uUmVxdWVzdEgAUhdmaWxlQ29udGFpbmluZ0V4dGVuc2lvbhJCCh1hbGxfZXh0ZW5zaW9uX251bWJlcnNfb2ZfdHlwZRgGIAEoCUgAUhlhbGxFeHRlbnNpb25OdW1iZXJzT2ZUeXBlEiUKDWxpc3Rfc2VydmljZXMYByABKAlIAFIMbGlzdFNlcnZpY2VzQhEKD21lc3NhZ2VfcmVxdWVzdCJmChBFeHRlbnNpb25SZXF1ZXN0EicKD2NvbnRhaW5pbmdfdHlwZRgBIAEoCVIOY29udGFpbmluZ1R5cGUSKQoQZXh0ZW5zaW9uX251bWJlchgCIAEoBVIPZXh0ZW5zaW9uTnVtYmVyIscEChhTZXJ2ZXJSZWZsZWN0aW9uUmVzcG9uc2USHQoKdmFsaWRfaG9zdBgBIAEoCVIJdmFsaWRIb3N0ElsKEG9yaWdpbmFsX3JlcXVlc3QYAiABKAsyMC5ncnBjLnJlZmxlY3Rpb24udjFhbHBoYS5TZXJ2ZXJSZWZsZWN0aW9uUmVxdWVzdFIPb3JpZ2luYWxSZXF1ZXN0EmsKGGZpbGVfZGVzY3JpcHRvcl9yZXNwb25zZRgEIAEoCzIvLmdycGMucmVmbGVjdGlvbi52MWFscGhhLkZpbGVEZXNjcmlwdG9yUmVzcG9uc2VIAFIWZmlsZURlc2NyaXB0b3JSZXNwb25zZRJ3Ch5hbGxfZXh0ZW5zaW9uX251bWJlcnNfcmVzcG9uc2UYBSABKAsyMC5ncnBjLnJlZmxlY3Rpb24udjFhbHBoYS5FeHRlbnNpb25OdW1iZXJSZXNwb25zZUgAUhthbGxFeHRlbnNpb25OdW1iZXJzUmVzcG9uc2USZAoWbGlzdF9zZXJ2aWNlc19yZXNwb25zZRgGIAEoCzIsLmdycGMucmVmbGVjdGlvbi52MWFscGhhLkxpc3RTZXJ2aWNlUmVzcG9uc2VIAFIUbGlzdFNlcnZpY2VzUmVzcG9uc2USTwoOZXJyb3JfcmVzcG9uc2UYByABKAsyJi5ncnBjLnJlZmxlY3Rpb24udjFhbHBoYS5FcnJvclJlc3BvbnNlSABSDWVycm9yUmVzcG9uc2VCEgoQbWVzc2FnZV9yZXNwb25zZSJMChZGaWxlRGVzY3JpcHRvclJlc3BvbnNlEjIKFWZpbGVfZGVzY3JpcHRvcl9wcm90bxgBIAMoDFITZmlsZURlc2NyaXB0b3JQcm90byJqChdFeHRlbnNpb25OdW1iZXJSZXNwb25zZRIkCg5iYXNlX3R5cGVfbmFtZRgBIAEoCVIMYmFzZVR5cGVOYW1lEikKEGV4dGVuc2lvbl9udW1iZXIYAiADKAVSD2V4dGVuc2lvbk51bWJlciJZChNMaXN0U2VydmljZVJlc3BvbnNlEkIKB3NlcnZpY2UYASADKAsyKC5ncnBjLnJlZmxlY3Rpb24udjFhbHBoYS5TZXJ2aWNlUmVzcG9uc2VSB3NlcnZpY2UiJQoPU2VydmljZVJlc3BvbnNlEhIKBG5hbWUYASABKAlSBG5hbWUiUwoNRXJyb3JSZXNwb25zZRIdCgplcnJvcl9jb2RlGAEgASgFUgllcnJvckNvZGUSIwoNZXJyb3JfbWVzc2FnZRgCIAEoCVIMZXJyb3JNZXNzYWdlMuMBChBTZXJ2ZXJSZWZsZWN0aW9uEs4BChRTZXJ2ZXJSZWZsZWN0aW9uSW5mbxIwLmdycGMucmVmbGVjdGlvbi52MWFscGhhLlNlcnZlclJlZmxlY3Rpb25SZXF1ZXN0GjEuZ3JwYy5yZWZsZWN0aW9uLnYxYWxwaGEuU2VydmVyUmVmbGVjdGlvblJlc3BvbnNlIk2C0+STAkc6ASoiQi9hcGkvZ3JwYy5yZWZsZWN0aW9uLnYxYWxwaGEuU2VydmVyUmVmbGVjdGlvbi9TZXJ2ZXJSZWZsZWN0aW9uSW5mbygBMAFCO1o5Z29vZ2xlLmdvbGFuZy5vcmcvZ3JwYy9yZWZsZWN0aW9uL2dycGNfcmVmbGVjdGlvbl92MWFscGhhYgZwcm90bzM="
	cmd := fdsCmd{FileDescriptorSet: b64FDS}
	b := &bytes.Buffer{}
	g := globals{
		Format: "json",
		out:    b,
	}
	err := cmd.Run(g)
	require.NoError(t, err)
	want := `{
  "file": [
    {
      "name": "echo/echo.proto",
      "package": "echo",
      "messageType": [
        {
          "name": "HelloRequest",
          "field": [ { "name": "message", "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "jsonName": "message" } ]
        },
        {
          "name": "HelloResponse",
          "field": [ { "name": "response", "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "jsonName": "response" } ]
        }
      ],
      "service": [
        {
          "name": "Echo",
          "method": [
            {
              "name": "Hello",
              "inputType": ".echo.HelloRequest",
              "outputType": ".echo.HelloResponse",
              "options": {
                "[google.api.http]": {
                  "post": "/api/echo.Echo/Hello",
                  "body": "*"
                }
              }
            }
          ]
        }
      ],
      "options": {
        "goPackage": "github.com/juliaogris/guppy/pkg/echo"
      },
      "syntax": "proto3"
    },
    {
      "name": "reflection/grpc_reflection_v1alpha/reflection.proto",
      "package": "grpc.reflection.v1alpha",
      "messageType": [
        {
          "name": "ServerReflectionRequest",
          "field": [ { "name": "host",                          "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "jsonName": "host" },
                     { "name": "file_by_filename",              "number": 3, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "oneofIndex": 0, "jsonName": "fileByFilename" },
                     { "name": "file_containing_symbol",        "number": 4, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "oneofIndex": 0, "jsonName": "fileContainingSymbol" },
                     { "name": "file_containing_extension",     "number": 5, "label": "LABEL_OPTIONAL", "type": "TYPE_MESSAGE", "typeName": ".grpc.reflection.v1alpha.ExtensionRequest", "oneofIndex": 0, "jsonName": "fileContainingExtension" },
                     { "name": "all_extension_numbers_of_type", "number": 6, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "oneofIndex": 0, "jsonName": "allExtensionNumbersOfType"},
                     { "name": "list_services",                 "number": 7, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "oneofIndex": 0, "jsonName": "listServices"} ],
          "oneofDecl": [ { "name": "message_request" } ]
        },
        {
          "name": "ExtensionRequest",
          "field": [ { "name": "containing_type",  "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "jsonName": "containingType" },
                     { "name": "extension_number", "number": 2, "label": "LABEL_OPTIONAL", "type": "TYPE_INT32", "jsonName": "extensionNumber" } ]
        },
        {
          "name": "ServerReflectionResponse",
          "field": [ { "name": "valid_host",                     "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING",  "jsonName": "validHost" },
                     { "name": "original_request",               "number": 2, "label": "LABEL_OPTIONAL", "type": "TYPE_MESSAGE", "typeName": ".grpc.reflection.v1alpha.ServerReflectionRequest", "jsonName": "originalRequest" },
                     { "name": "file_descriptor_response",       "number": 4, "label": "LABEL_OPTIONAL", "type": "TYPE_MESSAGE", "typeName": ".grpc.reflection.v1alpha.FileDescriptorResponse", "oneofIndex": 0, "jsonName": "fileDescriptorResponse" },
                     { "name": "all_extension_numbers_response", "number": 5, "label": "LABEL_OPTIONAL", "type": "TYPE_MESSAGE", "typeName": ".grpc.reflection.v1alpha.ExtensionNumberResponse", "oneofIndex": 0, "jsonName": "allExtensionNumbersResponse" },
                     { "name": "list_services_response",         "number": 6, "label": "LABEL_OPTIONAL", "type": "TYPE_MESSAGE", "typeName": ".grpc.reflection.v1alpha.ListServiceResponse", "oneofIndex": 0, "jsonName": "listServicesResponse" },
                     { "name": "error_response",                 "number": 7, "label": "LABEL_OPTIONAL", "type": "TYPE_MESSAGE", "typeName": ".grpc.reflection.v1alpha.ErrorResponse", "oneofIndex": 0, "jsonName": "errorResponse" } ],
          "oneofDecl": [ { "name": "message_response" } ]
        },
        {
          "name": "FileDescriptorResponse",
          "field": [ { "name": "file_descriptor_proto", "number": 1, "label": "LABEL_REPEATED", "type": "TYPE_BYTES", "jsonName": "fileDescriptorProto" } ]
        },
        {
          "name": "ExtensionNumberResponse",
          "field": [ { "name": "base_type_name",   "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "jsonName": "baseTypeName" },
                     { "name": "extension_number", "number": 2, "label": "LABEL_REPEATED", "type": "TYPE_INT32", "jsonName": "extensionNumber" } ]
        },
        {
          "name": "ListServiceResponse",
          "field": [ { "name": "service", "number": 1, "label": "LABEL_REPEATED", "type": "TYPE_MESSAGE", "typeName": ".grpc.reflection.v1alpha.ServiceResponse", "jsonName": "service" } ]
        },
        {
          "name": "ServiceResponse",
          "field": [ { "name": "name", "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "jsonName": "name" } ]
        },
        {
          "name": "ErrorResponse",
          "field": [ { "name": "error_code",    "number": 1, "label": "LABEL_OPTIONAL", "type": "TYPE_INT32", "jsonName": "errorCode" },
                     { "name": "error_message", "number": 2, "label": "LABEL_OPTIONAL", "type": "TYPE_STRING", "jsonName": "errorMessage" } ]
        }
      ],
      "service": [
        {
          "name": "ServerReflection",
          "method": [
            {
              "name": "ServerReflectionInfo",
              "inputType": ".grpc.reflection.v1alpha.ServerReflectionRequest",
              "outputType": ".grpc.reflection.v1alpha.ServerReflectionResponse",
              "options": {
                "[google.api.http]": {
                  "post": "/api/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
                  "body": "*"
                }
              },
              "clientStreaming": true,
              "serverStreaming": true
            }
          ]
        }
      ],
      "options": {
        "goPackage": "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
      },
      "syntax": "proto3"
    }
  ]
}`
	require.JSONEq(t, want, b.String())
}

func TestReflectSuite(t *testing.T) {
	suite.Run(t, &ReflectSuite{format: "json", pbVersion: 3})
	suite.Run(t, &ReflectSuite{format: "base64", pbVersion: 3})
	suite.Run(t, &ReflectSuite{format: "bin", pbVersion: 3})
	suite.Run(t, &ReflectSuite{format: "json", pbVersion: 2})
	suite.Run(t, &ReflectSuite{format: "base64", pbVersion: 2})
	suite.Run(t, &ReflectSuite{format: "bin", pbVersion: 2})
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
	requireResponseNoErr(t, f.got, s.format)
}

func (s *ReflectSuite) TestSymbolCmd() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	symbol := fmt.Sprintf("echo%d.Echo", s.pbVersion)
	cmd := symbolCmd{Symbol: symbol}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
	requireResponseNoErr(t, f.got, s.format)
}

func (s *ReflectSuite) TestSymbolCmdErr() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := symbolCmd{Symbol: "MISSING"}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
	requireResponseErr(t, f.got, s.format)
}

func (s *ReflectSuite) TestFilenameCmd() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	filename := fmt.Sprintf("echo%d/echo%d.proto", s.pbVersion, s.pbVersion)
	cmd := filenameCmd{Filename: filename}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
	requireResponseNoErr(t, f.got, s.format)
}

func (s *ReflectSuite) TestFilenameCmdErr() {
	t := s.T()
	f := files(t, s.format, s.subDir)
	s.globals.out = f.out

	cmd := filenameCmd{Filename: "MISSING"}
	err := cmd.Run(s.globals)

	require.NoError(t, err)
	requireContentEq(t, f.want, f.got, s.format)
	requireResponseErr(t, f.got, s.format)
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
	requireResponseNoErr(t, f.got, s.format)
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
	requireResponseErr(t, f.got, s.format)
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
	requireResponseErr(t, f.got, s.format)
}

func requireResponseErr(t *testing.T, fname, format string) {
	t.Helper()
	resp := reflectionResponse(t, fname, format)
	require.NotNilf(t, resp.GetErrorResponse(), "expected error response for %q got nil", fname)
}

func requireResponseNoErr(t *testing.T, fname, format string) {
	t.Helper()
	resp := reflectionResponse(t, fname, format)
	require.Nilf(t, resp.GetErrorResponse(), "expected no error for %q got %q", fname, resp.GetErrorResponse())
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
