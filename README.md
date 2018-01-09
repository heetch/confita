# confita

confita is a library that loads configuration from multiple backends and stores it in a struct.

## Install

```sh
go get -u github.com/heetch/confita
```

## Usage

```go
package main

import (
  "log"
  "time"
  "context"

  "github.com/coreos/etcd/clientv3"
  "github.com/heetch/confita"
  "github.com/heetch/confita/etcd"
)

type Config struct {
  Host        string `config:"host"`
  Port        int    `config:"port"`
  Database struct {
    URI string            `config:"database-uri,required"`
    Timeout time.Duration `config:"database-timeout"`
    Password string       `config:"-"`
  }
}

func main() {
  var cfg Config
  ctx := context.Background()

  // By default, the loader loads keys from the environment.
  loader := confita.NewLoader()
  err := loader.Load(ctx, &cfg)
  if err != nil {
    log.Fatal(err)
  }

  // From the environment and etcd, with a timeout of 5 seconds.
  client, err := clientv3.New(clientv3.Config{
    Endpoints: endpoints,
  })
  if err != nil {
    log.Fatal(err)
  }
  defer client.Close()

  loader = confita.NewLoader(
    env.NewBackend(),
    etcd.NewBackend(client, "prefix"),
  )
  err = loader.Load(ctx, &cfg)
  if err != nil {
    log.Fatal(err)
  }
}
```