package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/chiehting/kubernetes-service/cmd/internal/config"
	"github.com/chiehting/kubernetes-service/cmd/internal/handler"
	"github.com/chiehting/kubernetes-service/cmd/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var (
	// configFile is the path to the config file.
	configFile = flag.String("f", "etc/registry_replace.yaml", "the config file")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	defer func() {
		svc.DestroyKubernetesSetting(ctx)
		server.Stop()
	}()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	// server.Use(LogRequestBody)
	server.Start()
}

func LogRequestBody(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyString := string(bodyBytes)

		logx.Info(bodyString)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next(w, r)
	}
}
