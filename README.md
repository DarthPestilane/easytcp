# EasyTCP

[![gh-action](https://github.com/DarthPestilane/easytcp/actions/workflows/test.yml/badge.svg)](https://github.com/DarthPestilane/easytcp/actions/workflows/test.yml)
[![Go Report](https://goreportcard.com/badge/github.com/darthPestilane/easytcp)](https://goreportcard.com/report/github.com/darthPestilane/easytcp)
[![codecov](https://codecov.io/gh/DarthPestilane/easytcp/branch/master/graph/badge.svg?token=002KJ5IV4Z)](https://codecov.io/gh/DarthPestilane/easytcp)
[![Go Reference](https://pkg.go.dev/badge/github.com/DarthPestilane/easytcp.svg)](https://pkg.go.dev/github.com/DarthPestilane/easytcp)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#networking)

```
$ ./start

[EASYTCP ROUTE TABLE]:
+------------+-----------------------+
| Message ID |     Route Handler     |
+------------+-----------------------+
|       1000 | path/to/handler.Func1 |
+------------+-----------------------+
|       1002 | path/to/handler.Func2 |
+------------+-----------------------+
[EASYTCP] Serving at: tcp://[::]:10001
```

## Introduction

`EasyTCP` is a light-weight and less painful TCP server framework written in Go (Golang) based on the standard `net` package.

✨ Features:

- Non-invasive design
- Pipelined middlewares for route handler
- Customizable message packer and codec, and logger
- Handy functions to handle request data and send response
- Common hooks

`EasyTCP` helps you build a TCP server easily and fast.

This package, so far, has been tested with go1.15 ~ go1.17 on the latest Linux, Macos and Windows.

## Install

Use the below Go command to install EasyTCP.

```sh
$ go get -u github.com/DarthPestilane/easytcp
```

Note: EasyTCP uses **Go Modules** to manage dependencies.

## Quick start

```go
package main

import (
    "fmt"
    "github.com/DarthPestilane/easytcp"
    "github.com/DarthPestilane/easytcp/message"
)

func main() {
    // Create a new server with options.
    s := easytcp.NewServer(&easytcp.ServerOption{
        Packer: easytcp.NewDefaultPacker(), // use default packer
        Codec:  nil,                        // don't use codec
    })

    // Register a route with message's ID.
    // The `DefaultPacker` treats id as int,
    // so when we add routes or return response, we should use int.
    s.AddRoute(1001, func(c easytcp.Context) {
        // acquire request
        req := c.Request()

        // do things...
        fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", req.ID, len(req.Data), req.Data)

        // set response
        c.SetResponseMessage(&message.Entry{
            ID:   1002,
            Data: []byte("copy that"),
        })
    })

    // Set custom logger (optional).
    easytcp.SetLogger(lg)

    // Add global middlewares (optional).
    s.Use(recoverMiddleware)

    // Set hooks (optional).
    s.OnSessionCreate = func(session easytcp.Session) {...}
    s.OnSessionClose = func(session easytcp.Session) {...}

    // Set not-found route handler (optional).
    s.NotFoundHandler(handler)

    // Listen and serve.
    if err := s.Serve(":5896"); err != nil && err != server.ErrServerStopped {
        fmt.Println("serve error: ", err.Error())
    }
}
```

### If we setup with the codec

```go
// Create a new server with options.
s := easytcp.NewServer(&easytcp.ServerOption{
    Packer: easytcp.NewDefaultPacker(), // use default packer
    Codec:  &easytcp.JsonCodec{},       // use JsonCodec
})

// Register a route with message's ID.
// The `DefaultPacker` treats id as int,
// so when we add routes or return response, we should use int.
s.AddRoute(1001, func(c easytcp.Context) {
    // decode request data and bind to `reqData`
    var reqData map[string]interface{}
    if err := c.Bind(&reqData); err != nil {
        // handle err
    }

    // do things...
    respId := 1002
    respData := map[string]interface{}{
        "success": true,
        "feeling": "Great!",
    }

    // encode response data and set to `c`
    if err := c.SetResponse(respId, respData); err != nil {
        // handle err
    }
})
```

Above is the server side example. There are client and more detailed examples including:

- [broadcasting](./examples/tcp/broadcast)
- [custom packet](./examples/tcp/custom_packet)
- [communicating with protobuf](./examples/tcp/proto_packet)

in [examples/tcp](./examples/tcp).

## Benchmark

```sh
go test -bench=. -run=none -benchmem -benchtime=250000x
goos: darwin
goarch: amd64
pkg: github.com/DarthPestilane/easytcp
Benchmark_NoHandler-8              	  250000	      5891 ns/op	      96 B/op	       2 allocs/op
Benchmark_OneHandler-8             	  250000	      5772 ns/op	      95 B/op	       2 allocs/op
Benchmark_DefaultPacker_Pack-8     	  250000	        36.7 ns/op	      16 B/op	       1 allocs/op
Benchmark_DefaultPacker_Unpack-8   	  250000	      2239 ns/op	      95 B/op	       3 allocs/op
```

*since easytcp is built on the top of golang `net` library, the benchmark of networks does not make much sense.*

## Architecture

```
accepting connection:

+------------+    +-------------------+    +----------------+
|            |    |                   |    |                |
|            |    |                   |    |                |
| tcp server |--->| accept connection |--->| create session |
|            |    |                   |    |                |
|            |    |                   |    |                |
+------------+    +-------------------+    +----------------+

in session:

+------------------+    +-----------------------+    +----------------------------------+
| read connection  |--->| unpack packet payload |--->|                                  |
+------------------+    +-----------------------+    |                                  |
                                                     | router (middlewares and handler) |
+------------------+    +-----------------------+    |                                  |
| write connection |<---| pack packet payload   |<---|                                  |
+------------------+    +-----------------------+    +----------------------------------+

in route handler:

+----------------------------+    +------------+
| codec decode request data  |--->|            |
+----------------------------+    |            |
                                  | user logic |
+----------------------------+    |            |
| codec encode response data |<---|            |
+----------------------------+    +------------+
```

## Conception

### Routing

EasyTCP considers every message has a `ID` segment to distinguish one another.
A message will be routed, according to it's id, to the handler through middlewares.

```
request flow:

+----------+    +--------------+    +--------------+    +---------+
| request  |--->|              |--->|              |--->|         |
+----------+    |              |    |              |    |         |
                | middleware 1 |    | middleware 2 |    | handler |
+----------+    |              |    |              |    |         |
| response |<---|              |<---|              |<---|         |
+----------+    +--------------+    +--------------+    +---------+
```

#### Register a route

```go
s.AddRoute(reqID, func(c easytcp.Context) {
    // acquire request
    req := c.Request()

    // do things...
    fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", req.ID, len(req.Data), req.Data)

    // set response
    c.SetResponseMessage(&message.Entry{
        ID:   respID,
        Data: []byte("copy that"),
    })
})
```

#### Using middleware

```go
// register global middlewares.
// global middlewares are prior than per-route middlewares, they will be invoked first
s.Use(recoverMiddleware, logMiddleware, ...)

// register middlewares for one route
s.AddRoute(reqID, handler, middleware1, middleware2)

// a middleware looks like:
var exampleMiddleware easytcp.MiddlewareFunc = func(next easytcp.HandlerFunc) easytcp.HandlerFunc {
    return func(c easytcp.Context) {
        // do things before...
        next(c)
        // do things after...
    }
}
```

### Packer

A packer is to pack and unpack packets' payload. We can set the Packer when creating the server.

```go
s := easytcp.NewServer(&easytcp.ServerOption{
    Packer: new(MyPacker), // this is optional, the default one is DefaultPacker
})
```

We can set our own Packer or EasyTCP uses [`DefaultPacker`](./packer.go).

The `DefaultPacker` considers packet's payload as a `Size(4)|ID(4)|Data(n)` format. (`Size` only represents the length of `Data` instead of the whole payload length)

This may not covery some particular cases, but fortunately, we can create our own Packer.

```go
// CustomPacker is a custom packer, implements Packer interafce.
// Treats Packet format as `size(2)id(2)data(n)`
type CustomPacker struct{}

func (p *CustomPacker) bytesOrder() binary.ByteOrder {
    return binary.BigEndian
}

func (p *CustomPacker) Pack(entry *message.Entry) ([]byte, error) {
    size := len(entry.Data) // only the size of data.
    buffer := make([]byte, 2+2+size)
    p.bytesOrder().PutUint16(buffer[:2], uint16(size))
    p.bytesOrder().PutUint16(buffer[2:4], entry.ID.(uint16))
    copy(buffer[4:], entry.Data)
    return buffer, nil
}

func (p *CustomPacker) Unpack(reader io.Reader) (*message.Entry, error) {
    headerBuffer := make([]byte, 2+2)
    if _, err := io.ReadFull(reader, headerBuffer); err != nil {
        return nil, fmt.Errorf("read size and id err: %s", err)
    }
    size := p.bytesOrder().Uint16(headerBuffer[:2])
    id := p.bytesOrder().Uint16(headerBuffer[2:])

    data := make([]byte, size)
    if _, err := io.ReadFull(reader, data); err != nil {
        return nil, fmt.Errorf("read data err: %s", err)
    }

    entry := &message.Entry{
        // since entry.ID is type of uint16, we need to use uint16 as well when adding routes.
        // eg: server.AddRoute(uint16(123), ...)
        ID:   id,
        Data: data,
    }
    entry.Set("theWholeLength", 2+2+size) // we can set our custom kv data here.
    // c.Request().Get("theWholeLength")  // and get them in route handler.
    return entry, nil
}
```

And see more custom packers:
- [custom_packet](./examples/tcp/custom_packet/common/packer.go)
- [proto_packet](./examples/tcp/proto_packet/common/packer.go)

### Codec

A Codec is to encode and decode message data. The Codec is optional, EasyTCP won't encode or decode message data if the Codec is not set.

We can set Codec when creating the server.

```go
s := easytcp.NewServer(&easytcp.ServerOption{
    Codec: &easytcp.JsonCodec{}, // this is optional. The JsonCodec is a built-in codec
})
```

Since we set the codec, we may want to decode the request data in route handler.

```go
s.AddRoute(reqID, func(c easytcp.Context) {
    var reqData map[string]interface{}
    if err := c.Bind(&reqData); err != nil { // here we decode message data and bind to reqData
        // handle error...
    }
    req := c.Request()
    fmt.Printf("[server] request received | id: %d; size: %d; data-decoded: %+v\n", req.ID, len(req.Data), reqData)
    respData := map[string]string{"key": "value"}
    if err := c.SetResponse(respID, respData); err != nil {
        // handle error...
    }
})
```

Codec's encoding will be invoked before message packed,
and decoding should be invoked in the route handler which is after message unpacked.

#### JSON Codec

`JsonCodec` is an EasyTCP's built-in codec, which uses `encoding/json` as the default implementation.
Can be changed by build from other tags.

[jsoniter](https://github.com/json-iterator/go) :

```sh
go build -tags=jsoniter .
```

#### Protobuf Codec

`ProtobufCodec` is an EasyTCP's built-in codec, which uses `google.golang.org/protobuf` as the implementation.

#### Msgpack Codec

`MsgpackCodec` is an EasyTCP's built-in codec, which uses `github.com/vmihailenco/msgpack` as the implementation.

## Contribute

Check out a new branch for the job, and make sure github action passed.

Use issues for everything

- For a small change, just send a PR.
- For bigger changes open an issue for discussion before sending a PR.
- PR should have:
  - Test case
  - Documentation
  - Example (If it makes sense)
- You can also contribute by:
  - Reporting issues
  - Suggesting new features or enhancements
  - Improve/fix documentation

## Stargazers over time

[![Stargazers over time](https://starchart.cc/DarthPestilane/easytcp.svg)](https://starchart.cc/DarthPestilane/easytcp)
