package ssm

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/heetch/confita/backend"
)

type Backend struct {
	client  ssmiface.SSMAPI
	ssmPath string
	cache   map[string][]byte
}

func NewBackend(ssm ssmiface.SSMAPI, path string) *Backend {
	return &Backend{client: ssm, ssmPath: path}
}

func (b *Backend) Get(ctx context.Context, key string) ([]byte, error) {
	if b.cache == nil {
		err := b.fetchParams(ctx)
		if err != nil {
			return nil, err
		}
	}

	return b.fromCache(ctx, key)
}

func (b *Backend) Name() string {
	return "ssm"
}

func (b *Backend) fetchParams(ctx context.Context) error {
	b.cache = make(map[string][]byte)

	ssmInput := &ssm.GetParametersByPathInput{
		Path:           &b.ssmPath,
		Recursive:      newBool(true),
		WithDecryption: newBool(true),
		MaxResults:     newInt64(10),
	}

	for {
		res, err := b.client.GetParametersByPathWithContext(ctx, ssmInput)
		if err != nil {
			return err
		}

		for _, p := range res.Parameters {
			path := strings.Split(*p.Name, "/")
			key := path[len(path)-1]
			b.cache[key] = []byte(*p.Value)
		}

		if res.NextToken == nil {
			break
		}

		ssmInput.NextToken = res.NextToken
	}

	return nil
}

func (b *Backend) fromCache(ctx context.Context, key string) ([]byte, error) {
	v, ok := b.cache[key]
	if !ok {
		return nil, backend.ErrNotFound
	}

	return v, nil
}

func newBool(b bool) *bool {
	return &b
}

func newInt64(i int64) *int64 {
	return &i
}
