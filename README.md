# go-grpcpool

<p align="left"> 
<a href="https://hits.seeyoufarm.com"><img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fgjbae1212%2Fgo-grpcpool&count_bg=%2379C83D&title_bg=%23555555&icon=go.svg&icon_color=%2308BEB8&title=hits&edge_flat=false"/></a>               
   <a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-GREEN.svg" alt="license"/></a>
</p>

**go-gprcpool** is a pool for GRPC connections.  
It's a library for **golang**.

## Getting Started
### Install
```bash
$ go get -u github.com/gjbae1212/go-grpcpool
```

### Usage
```go
// pseudo code
package main
import (
    "context"  
    grpcpool "github.com/gjbae1212/go-grpcpool"    
)

func main() {
	pool, _ := grpcpool.New(func() (*grpc.ClientConn, error){ 
                  return nil, nil // a function which creates GRPC connection. 
              })
    conn, _ := pool.GetConn()
}  
```

## LICENSE
This project is following The MIT.
