package ssm

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/require"
)

func TestAWSError(t *testing.T) {
	client := newFakeSSM(t, "/borked/", []getParamsRequest{{
		resultErr: fmt.Errorf("aws down"),
	}})
	b := NewBackend(client, "/borked/")
	_, err := b.Get(context.Background(), "some_key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "aws down")
}

func TestNilNameAndValue(t *testing.T) {
	// nil names and values in the response are ignored.
	client := newFakeSSM(t, "/borked/", []getParamsRequest{{
		result: &ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{{
				Name:  nil,
				Value: newString("ignorevalue"),
			}, {
				Name:  newString("ignorename"),
				Value: nil,
			}, {
				Name:  newString("ignoreboth"),
				Value: newString("ignoreboth"),
			}, {
				Name:  newString("/sup/key"),
				Value: newString("hello"),
			}},
		},
	}})
	b := NewBackend(client, "/borked/")
	val, err := b.Get(context.Background(), "key")
	require.NoError(t, err)
	require.Equal(t, "hello", string(val))
}

func TestEmptyKey(t *testing.T) {
	client := newFakeSSM(t, "/borked/", []getParamsRequest{{
		result: &ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{{
				Name:  newString("/sup/key"),
				Value: newString("hello"),
			}},
		},
	}})
	b := NewBackend(client, "/borked/")
	val, err := b.Get(context.Background(), "")
	require.Equal(t, backend.ErrNotFound, err)
	require.Equal(t, "", string(val))
}

func TestKeyNotFound(t *testing.T) {
	client := newFakeSSM(t, "/whatevs/", []getParamsRequest{{
		result: &ssm.GetParametersByPathOutput{},
	}})
	b := NewBackend(client, "/whatevs/")
	val, err := b.Get(context.Background(), "some_key")
	require.Equal(t, backend.ErrNotFound, err)
	require.Equal(t, "", string(val))
}

func TestKeysFound(t *testing.T) {
	client := newFakeSSM(t, "/yo/whatup/", []getParamsRequest{{
		result: &ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{Name: newString("/yo/whatup/a_key"), Value: newString("wow")},
				{Name: newString("/yo/whatup/some_key"), Value: newString("wondrous")},
			},
		},
	}})
	b := NewBackend(client, "/yo/whatup/")
	val, err := b.Get(context.Background(), "a_key")
	require.Nil(t, err)
	require.Equal(t, "wow", string(val))
	val, err = b.Get(context.Background(), "some_key")
	require.Nil(t, err)
	require.Equal(t, "wondrous", string(val))
}

func TestSSMPagedCall(t *testing.T) {
	client := newFakeSSM(t, "/a/path/", []getParamsRequest{{
		result: &ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{Name: newString("/yo/whatup/a_key"), Value: newString("wow")},
				{Name: newString("/yo/whatup/some_key"), Value: newString("wondrous")},
			},
			NextToken: newString("/a/path/your_key"),
		},
	}, {
		expectToken: newString("/a/path/your_key"),
		result: &ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{Name: newString("/a/path/your_key"), Value: newString("shazam")},
				{Name: newString("/a/path/another_key"), Value: newString("kazam")},
			},
			NextToken: nil,
		},
	}})

	b := NewBackend(client, "/a/path/")
	actual, err := b.Get(context.Background(), "your_key")
	require.Nil(t, err)
	require.Equal(t, "shazam", string(actual))
	actual, err = b.Get(context.Background(), "another_key")
	require.Nil(t, err)
	require.Equal(t, "kazam", string(actual))
}

type getParamsRequest struct {
	expectToken *string
	result      *ssm.GetParametersByPathOutput
	resultErr   error
}

type fakeSSM struct {
	ssmiface.SSMAPI
	t *testing.T
	// We always expect the backend to use the same path.
	expectPath string
	// The sequence of calls we're expecting.
	calls []getParamsRequest
	// The index of the next expected call.
	call int
}

func newFakeSSM(t *testing.T, path string, calls []getParamsRequest) *fakeSSM {
	return &fakeSSM{
		t:          t,
		expectPath: path,
		calls:      calls,
	}
}

func (f *fakeSSM) GetParametersByPathWithContext(ctx aws.Context, arg *ssm.GetParametersByPathInput, opts ...request.Option) (*ssm.GetParametersByPathOutput, error) {
	if f.call >= len(f.calls) {
		f.t.Errorf("too many calls to SSM (expected max of %d)", len(f.calls))
	}
	call := f.calls[f.call]
	f.call++
	require.Equal(f.t, newString(f.expectPath), arg.Path)
	require.Equal(f.t, call.expectToken, arg.NextToken)
	return call.result, call.resultErr
}

func newString(str string) *string {
	return &str
}
