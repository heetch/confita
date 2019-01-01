![](https://user-images.githubusercontent.com/4570448/40795774-57ad2ae2-6503-11e8-8448-328094f4091f.png)

[![Build Status](https://travis-ci.org/heetch/confita.svg?branch=master)](https://travis-ci.org/heetch/confita)
[![GoDoc](https://godoc.org/github.com/heetch/confita?status.svg)](https://godoc.org/github.com/heetch/confita)
[![Go Report Card](https://goreportcard.com/badge/github.com/heetch/confita)](https://goreportcard.com/report/github.com/heetch/confita)

Confita is a library that loads configuration from multiple backends and stores it in a struct.

## Supported backends

- Environment variables
- JSON files
- Yaml files
- Toml files
- Command line flags
- [etcd](https://github.com/coreos/etcd)
- [Consul](https://www.consul.io/)
- [Vault](https://www.vaultproject.io/)

## Install

```sh
go get -u github.com/karantin2020/confita
```

## Usage

Confita scans a struct for `config` tags and calls all the backends one after another until the key is found.
The value is then converted into the type of the field.

### Struct layout

Go primitives are supported:

```go
type Config struct {
  Host        string        `config:"host"`
  Port        uint32        `config:"port"`
  Timeout     time.Duration `config:"timeout"`
}
```

By default, all fields are optional. With the required option, if a key is not found then Confita will return an error.

```go
type Config struct {
  Addr        string        `config:"addr,required"`
  Timeout     time.Duration `config:"timeout"`
}
```

Nested structs are supported too:

```go
type Config struct {
  Host        string        `config:"host"`
  Port        uint32        `config:"port"`
  Timeout time.Duration     `config:"timeout"`
  Database struct {
    URI string              `config:"database-uri,required"`
  }
}
```

If a field is a slice, Confita will automatically split the config value by commas and fill the slice with each sub value.

```go
type Config struct {
  Endpoints []string `config:"endpoints"`
}
```

As a special case, if the field tag is "-", the field is always omitted. This is useful if you want to populate this field on your own.

```go
type Config struct {
  // Field is ignored by this package.
  Field float64 `config:"-"`

  // Confita scans any structure recursively, the "-" value prevents that.
  Client http.Client `config:"-"`
}
```

### Loading configuration

Creating a loader:

```go
loader := confita.NewLoader()
```

By default, a Confita loader loads all the keys from the environment.
A loader can take other configured backends as parameters.

```go
loader := confita.NewLoader(
  env.NewBackend(
    env.WithPrefix("PREFIX"),
		env.ToUpper(),
  ),
  file.NewBackend("/path/to/config.json"),
  file.NewBackend("/path/to/config.yaml"),
  flags.NewBackend(),
  etcd.NewBackend(etcdClientv3),
  consul.NewBackend(consulClient),
  vault.NewBackend(vaultClient),
)
```

Loading configuration:

```go
err := loader.Load(context.Background(), &cfg)
```

Since loading configuration can take time when used with multiple remote backends, context can be used for timeout and cancelation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
defer cancel()
err := loader.Load(ctx, &cfg)
```

### Default values

If a key is not found, Confita won't change the respective struct field. With that in mind, default values can simply be implemented by filling the structure before passing it to Confita.

```go
type Config struct {
  Host        string        `config:"host"`
  Port        uint32        `config:"port"`
  Timeout     time.Duration `config:"timeout"`
  Password    string        `config:"password,required"`
}

// default values
cfg := Config{
  Host: "127.0.0.1",
  Port: "5656",
  Timeout: 5 * time.Second,
}

err := confita.NewLoader().Load(context.Background(), &cfg)
```

### Backend option

By default, Confita queries each backend one after another until a key is found. However, in order to avoid some useless processing the `backend` option can be specified to describe in which backend this key is expected to be found.
This is especially useful when the location of the key is known beforehand.

```go
type Config struct {
  Host        string        `config:"host,backend=env"`
  Port        uint32        `config:"port,required,backend=etcd"`
  Timeout     time.Duration `config:"timeout"`
}
```

### Command line flags

The `flags` backend allows to load individual configuration keys from the command line. The default values are extracted from the struct fields values.

```sh
./bin -h

Usage of ./bin:
  -host string
       (default "127.0.0.1")
  -port int
       (default 5656)
  -timeout duration
       (default 10s)
```

## License

The library is released under the MIT license. See LICENSE file.
