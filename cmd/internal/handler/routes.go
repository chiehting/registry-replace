// Code generated by goctl. DO NOT EDIT.
// goctl 1.7.2

package handler

import (
	"net/http"

	"github.com/chiehting/kubernetes-service/cmd/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/mutate",
				Handler: mutateHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/ping",
				Handler: pingHandler(serverCtx),
			},
		},
	)
}
