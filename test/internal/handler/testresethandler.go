package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"main/test/internal/logic"
	"main/test/internal/svc"
	"main/test/internal/types"
)

func TestResetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RequestReset
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewTestResetLogic(r.Context(), svcCtx)
		resp, err := l.TestReset(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
