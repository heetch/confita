package ssm

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/heetch/confita/backend"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSSM struct {
	mock.Mock
	ssmiface.SSMAPI
}

func (_m *mockSSM) GetParametersByPathWithContext(_a0 context.Context, _a1 *ssm.GetParametersByPathInput, _a2 ...request.Option) (*ssm.GetParametersByPathOutput, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *ssm.GetParametersByPathOutput
	if rf, ok := ret.Get(0).(func(context.Context, *ssm.GetParametersByPathInput, ...request.Option) *ssm.GetParametersByPathOutput); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ssm.GetParametersByPathOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ssm.GetParametersByPathInput, ...request.Option) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func TestAWSError(t *testing.T) {
	client := new(mockSSM)
	ssmOpts := getSSMOpts("/borked/")
	ctx := context.Background()
	expected := fmt.Errorf("aws down")
	client.On("GetParametersByPathWithContext", ctx, ssmOpts).Return(
		nil, expected)

	b := NewBackend(client, "/borked/")
	_, actual := b.Get(context.Background(), "some_key")
	require.Equal(t, expected, actual)
}

func TestNilNameAndValue(t *testing.T) {
	client := new(mockSSM)
	ssmOpts := getSSMOpts("/sup/")
	ctx := context.Background()

	client.On("GetParametersByPathWithContext", ctx, ssmOpts).Return(&ssm.GetParametersByPathOutput{
		Parameters: []*ssm.Parameter{
			{
				Name:  nil,
				Value: nil,
			},
			{
				Name:  ptrString("/sup/key"),
				Value: nil,
			},
		},
	}, nil)

	b := NewBackend(client, "/sup/")

	_, actual := b.Get(context.Background(), "key")
	require.Equal(t, backend.ErrNotFound, actual)
}

func TestKeyNotFound(t *testing.T) {
	client := new(mockSSM)
	ssmOpts := getSSMOpts("/whatevs/")
	ctx := context.Background()
	client.On("GetParametersByPathWithContext", ctx, ssmOpts).Return(
		&ssm.GetParametersByPathOutput{}, nil)

	b := NewBackend(client, "/whatevs/")
	_, actual := b.Get(context.Background(), "some_key")
	require.Equal(t, backend.ErrNotFound, actual)
}

func ptrString(str string) *string {
	return &str
}

func TestKeysFound(t *testing.T) {
	client := new(mockSSM)
	ctx := context.Background()
	ssmOpts := getSSMOpts("/yo/whatup/")
	client.On("GetParametersByPathWithContext", ctx, ssmOpts).Return(
		&ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{Name: ptrString("/yo/whatup/a_key"), Value: ptrString("wow")},
				{Name: ptrString("/yo/whatup/some_key"), Value: ptrString("wondrous")},
			},
		}, nil)

	b := NewBackend(client, "/yo/whatup/")
	actual, err := b.Get(context.Background(), "a_key")
	require.Nil(t, err)
	require.Equal(t, "wow", string(actual))
	actual, err = b.Get(context.Background(), "some_key")
	require.Nil(t, err)
	require.Equal(t, "wondrous", string(actual))
}

func TestSSMPagedCall(t *testing.T) {
	client := new(mockSSM)
	ctx := context.Background()
	firstOpts := getSSMOpts("/a/path/")
	client.On("GetParametersByPathWithContext", ctx, firstOpts).Return(
		&ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{},
			NextToken:  ptrString("/a/path/your_key"),
		}, nil)

	secondOpts := getSSMOpts("/a/path/")
	secondOpts.NextToken = ptrString("/a/path/your_key")
	client.On("GetParametersByPathWithContext", ctx, secondOpts).Return(
		&ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{Name: ptrString("/a/path/your_key"), Value: ptrString("shazam")},
				{Name: ptrString("/a/path/another_key"), Value: ptrString("kazam")},
			},
			NextToken: nil,
		}, nil)

	b := NewBackend(client, "/a/path/")
	actual, err := b.Get(context.Background(), "your_key")
	require.Nil(t, err)
	require.Equal(t, "shazam", string(actual))
	actual, err = b.Get(context.Background(), "another_key")
	require.Nil(t, err)
	require.Equal(t, "kazam", string(actual))
}

func getSSMOpts(path string) *ssm.GetParametersByPathInput {
	return &ssm.GetParametersByPathInput{
		Path:           &path,
		Recursive:      newBool(true),
		WithDecryption: newBool(true),
		MaxResults:     newInt64(10),
	}
}
