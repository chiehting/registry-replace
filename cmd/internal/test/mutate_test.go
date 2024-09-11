package test

import (
	"net/http"
	"testing"

	"github.com/chiehting/kubernetes-service/cmd/internal/logic"
	"github.com/chiehting/kubernetes-service/cmd/internal/svc"
	"github.com/chiehting/kubernetes-service/cmd/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMutatePod(t *testing.T) {
	r := http.Request{}
	svcCtx := &svc.ServiceContext{}
	l := logic.NewMutateLogic(r.Context(), svcCtx)

	req := &types.AdmissionReview{
		APIVersion: "admission.k8s.io/v1",
		Kind:       "AdmissionReview",
		Request: &types.AdmissionRequest{
			UID: "705ab4f5-6393-11e8-b7cc-42010a800002",
			Kind: &types.Object{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
		},
	}

	resp, err := l.Mutate(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Response.Allowed)
	assert.NotEmpty(t, resp.Response.Patch)
	assert.Equal(t, "JSONPatch", resp.Response.PatchType)
}
