package main

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bayserver/impl"
	"embed"
	"os"
)

//go:embed resources
var res embed.FS

func main() {
	impl.Init(res)
	bayserver.Main(os.Args)
}
