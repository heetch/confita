# confita

confita is a library that loads configuration from multiple backends and stores it in a struct.

## Install

```sh
go get -u github.com/heetch/confita
```

## Usage

confita scans a struct for `config` tags and calls all the backends one after another until the key is found.
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

By default, all fields are optional. With the `required` option, if a key is not found confita will return an error.

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
```

As a special case, if the field tag is "-", the field is always omitted.

```go
type Config struct {
  // Field is ignored by this package.
  Field float64 `config:"-"`

  // confita scans any structure recursively, the "-" value prevents that.
  Client http.Client `config:"-"`
}
```

### Loading configuration

Creating a loader:

```go
loader := confita.NewLoader()
```

By default, a confita loader loads all the keys from the environment.
A loader can take other configured backends as parameters. For now, only [etcd](https://github.com/coreos/etcd) is supported.

```go
loader := confita.NewLoader(
  env.NewBackend(),
  etcd.NewBackend(etcdClientv3, "namespace"),
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

If a key is not found, confita won't change the respective struct field. With that in mind, default values can simply be implemented by filling the structure before passing it to confita.

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

err := config.NewLoader().Load(context.Background(), &cfg)
```
