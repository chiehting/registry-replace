package logic

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/chiehting/kubernetes-service/cmd/internal/svc"
	"github.com/chiehting/kubernetes-service/cmd/internal/types"
	"github.com/containers/image/docker/reference"
	"github.com/zeromicro/go-zero/core/logx"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

type MutateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMutateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MutateLogic {
	return &MutateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MutateLogic) Mutate(req *types.AdmissionReview) (*types.AdmissionReview, error) {
	doNothing := &types.AdmissionReview{
		Kind:       req.Kind,
		APIVersion: req.APIVersion,
		Response: &types.AdmissionResponse{
			UID:     req.Request.UID,
			Allowed: true,
		},
	}

	if req.Request.Kind.Kind != "Pod" {
		return doNothing, nil
	}

	pod := &corev1.Pod{}
	jsonData, err := json.Marshal(req.Request.Object)
	if err != nil {
		logx.Error(err.Error())
		return doNothing, nil
	}

	if err = json.Unmarshal(jsonData, pod); err != nil {
		logx.Error(err.Error())
		return doNothing, nil
	}

	podName := pod.Name
	if podName == "" {
		podName = pod.GenerateName
	}
	logx.Infof("podName: %s", podName)
	l.replaceImageUrl(pod)

	patch := []map[string]interface{}{
		{
			"op":    "replace",
			"path":  "/spec/initContainers",
			"value": pod.Spec.InitContainers,
		},
		{
			"op":    "replace",
			"path":  "/spec/containers",
			"value": pod.Spec.Containers,
		},
		{
			"op":    "replace",
			"path":  "/spec/ephemeralContainers",
			"value": pod.Spec.EphemeralContainers,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		logx.Error(err.Error())
		return doNothing, nil
	}

	admissionReview := &types.AdmissionReview{
		Kind:       req.Kind,
		APIVersion: req.APIVersion,
		Response: &types.AdmissionResponse{
			UID:       req.Request.UID,
			Allowed:   true,
			Patch:     patchBytes,
			PatchType: string(admissionv1.PatchTypeJSONPatch),
		},
	}
	return admissionReview, nil
}

func (l *MutateLogic) replaceImageUrl(pod *corev1.Pod) {
	for i := range pod.Spec.InitContainers {
		pod.Spec.InitContainers[i].Image = l.getMirrorHost(pod.Spec.InitContainers[i].Image)
	}

	for i := range pod.Spec.Containers {
		pod.Spec.Containers[i].Image = l.getMirrorHost(pod.Spec.Containers[i].Image)
	}

	for i := range pod.Spec.EphemeralContainers {
		pod.Spec.EphemeralContainers[i].EphemeralContainerCommon.Image = l.getMirrorHost(pod.Spec.EphemeralContainers[i].EphemeralContainerCommon.Image)
	}
}

func (l *MutateLogic) getMirrorHost(originalImage string) string {
	named, err := reference.ParseDockerRef(originalImage)
	if err != nil {
		logx.Error(err)
		return originalImage
	}

	originalRegistry := reference.Domain(named)
	mirrorRegistry, _ := l.svcCtx.Kubernetes.HostMap.Data[originalRegistry]
	if mirrorRegistry == "" {
		logx.Debugf("mirror registry not found: %s", originalRegistry)
		return originalImage
	}
	path := strings.TrimPrefix(reference.TagNameOnly(named).String(), originalRegistry+"/")
	return mirrorRegistry + "/" + path
}
