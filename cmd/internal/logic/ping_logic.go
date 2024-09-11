package logic

import (
	"context"

	"github.com/chiehting/kubernetes-service/cmd/internal/svc"
	"github.com/chiehting/kubernetes-service/cmd/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PingLogic) Ping(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
