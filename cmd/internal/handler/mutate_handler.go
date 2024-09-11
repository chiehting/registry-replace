package handler

import (
	"net/http"

	"github.com/chiehting/kubernetes-service/cmd/internal/logic"
	"github.com/chiehting/kubernetes-service/cmd/internal/svc"
	"github.com/chiehting/kubernetes-service/cmd/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func mutateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdmissionReview
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewMutateLogic(r.Context(), svcCtx)
		resp, err := l.Mutate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
