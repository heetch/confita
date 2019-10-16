package ssm

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"

	"github.com/heetch/confita/backend"
)

type ssmBackend struct {
	client  ssmiface.SSMAPI
	ssmPath string
	cache   map[string][]byte
}

// NewBackend returns a backend instance that uses the given SSMAPI implementation
// to retrieve keys from the parameter store at the given path.
func NewBackend(ssm ssmiface.SSMAPI, path string) backend.Backend {
	return &ssmBackend{client: ssm, ssmPath: path}
}

// Get implements backend.Backend.Get by fetching the key from SSM params.
func (b *ssmBackend) Get(ctx context.Context, key string) ([]byte, error) {
	if b.cache == nil {
		err := b.fetchParams(ctx)
		if err != nil {
			return nil, err
		}
	}

	return b.fromCache(ctx, key)
}

// Name implements backend.Backend.Name.
func (b *ssmBackend) Name() string {
	return "ssm"
}

func (b *ssmBackend) fetchParams(ctx context.Context) error {
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
			if p.Name == nil || p.Value == nil {
				continue
			}
			path := strings.Split(*p.Name, "/")
			if key := path[len(path)-1]; key != "" {
				b.cache[key] = []byte(*p.Value)
			}
		}
		if res.NextToken == nil {
			break
		}
		ssmInput.NextToken = res.NextToken
	}
	return nil
}

func (b *ssmBackend) fromCache(ctx context.Context, key string) ([]byte, error) {
	if v, ok := b.cache[key]; ok {
		return v, nil
	}
	return nil, backend.ErrNotFound
}

func newBool(b bool) *bool {
	return &b
}

func newInt64(i int64) *int64 {
	return &i
}
