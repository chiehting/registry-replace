package svc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/chiehting/kubernetes-service/cmd/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	certuitl "k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ServiceContext struct {
	Config     config.Config
	Kubernetes struct {
		Client    *kubernetes.Clientset
		Name      string
		Namespace string
		HostMap   *corev1.ConfigMap
	}
}

var (
	alternateDNS string
	cert         []byte
	key          []byte
)

func NewServiceContext(c config.Config) *ServiceContext {
	namespace := getNamespace()
	if len(c.IncludeNamespaces) == 0 {
		c.IncludeNamespaces = config.IncludeNamespaces
	}
	if len(c.ExcludeNamespaces) == 0 {
		c.ExcludeNamespaces = config.ExcludeNamespaces
		c.ExcludeNamespaces = append(c.ExcludeNamespaces, namespace)
	}
	svcCtx := &ServiceContext{
		Config: c,
		Kubernetes: struct {
			Client    *kubernetes.Clientset
			Name      string
			Namespace string
			HostMap   *corev1.ConfigMap
		}{
			getClient(),
			strings.ToLower(c.Name),
			namespace,
			&corev1.ConfigMap{},
		},
	}
	// Destory k8s setting
	DestroyKubernetesSetting(svcCtx)
	// Init tls
	createTlsSecret(svcCtx)
	writeTLSFiles(svcCtx)
	// Init MutatingWebhookConfiguration
	createMutatingWebhookConfiguration(svcCtx)
	// Init configmap
	watchConfigMap(svcCtx)
	createConfigMap(svcCtx)
	// Init service
	createService(svcCtx)
	return svcCtx
}

func DestroyKubernetesSetting(sc *ServiceContext) {
	k := &sc.Kubernetes
	admissionregistration, _ := k.Client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), k.Name, metav1.GetOptions{})
	if admissionregistration != nil {
		k.Client.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.TODO(), k.Name, metav1.DeleteOptions{})
	}
	secret, _ := k.Client.CoreV1().Secrets(k.Namespace).Get(context.TODO(), k.Name, metav1.GetOptions{})
	if secret != nil {
		k.Client.CoreV1().Secrets(k.Namespace).Delete(context.TODO(), k.Name, metav1.DeleteOptions{})
	}
}

func createService(sc *ServiceContext) {
	k := &sc.Kubernetes
	podName, _ := os.ReadFile("/etc/hostname")
	pod, err := k.Client.CoreV1().Pods(k.Namespace).Get(context.TODO(), strings.TrimSpace(string(podName)), metav1.GetOptions{})
	if err != nil {
		logx.Errorf(err.Error())
		return
	}
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels["app"] = k.Name
	_, err = k.Client.CoreV1().Pods(k.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name,
			Namespace: k.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": k.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "webhook",
					Port: 443,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8443,
					},
				},
			},
		},
	}
	_, err = k.Client.CoreV1().Services(k.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		logx.Errorf("Create Service failed: %v", err)
	}
}

func createMutatingWebhookConfiguration(sc *ServiceContext) {
	k := &sc.Kubernetes
	failurePolicy := admissionregistrationv1.Fail
	sideEffects := admissionregistrationv1.SideEffectClassNone
	timeoutSeconds := int32(5)
	rulesScope := admissionregistrationv1.NamespacedScope
	matchPloicy := admissionregistrationv1.Exact
	webhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: k.Name,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name: alternateDNS,
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Name:      k.Name,
					Namespace: k.Namespace,
					Path:      &config.Path,
				},
				CABundle: cert,
			},
			Rules: []admissionregistrationv1.RuleWithOperations{{
				Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"pods"},
					Scope:       &rulesScope,
				},
			}},
			NamespaceSelector:       &metav1.LabelSelector{},
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &sideEffects,
			TimeoutSeconds:          &timeoutSeconds,
			FailurePolicy:           &failurePolicy,
			MatchPolicy:             &matchPloicy,
		}},
	}
	const metadataKey = "kubernetes.io/metadata.name"
	if v := sc.Config.ExcludeNamespaces; len(v) > 0 {
		var selector metav1.LabelSelectorRequirement
		if slices.Contains(v, "*") {
			selector = metav1.LabelSelectorRequirement{
				Key:      metadataKey,
				Operator: metav1.LabelSelectorOpDoesNotExist,
			}
		} else {
			selector = metav1.LabelSelectorRequirement{
				Key:      metadataKey,
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   v,
			}
		}
		webhookConfig.Webhooks[0].NamespaceSelector.MatchExpressions = append(webhookConfig.Webhooks[0].NamespaceSelector.MatchExpressions, selector)
	}
	if v := sc.Config.IncludeNamespaces; len(v) > 0 {
		var selector metav1.LabelSelectorRequirement
		if slices.Contains(v, "*") {
			selector = metav1.LabelSelectorRequirement{
				Key:      metadataKey,
				Operator: metav1.LabelSelectorOpExists,
			}
		} else {
			selector = metav1.LabelSelectorRequirement{
				Key:      metadataKey,
				Operator: metav1.LabelSelectorOpIn,
				Values:   v,
			}
		}
		webhookConfig.Webhooks[0].NamespaceSelector.MatchExpressions = append(webhookConfig.Webhooks[0].NamespaceSelector.MatchExpressions, selector)
	}
	_, err := k.Client.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(
		context.TODO(),
		webhookConfig,
		metav1.CreateOptions{},
	)
	if err != nil {
		logx.Errorf("%s", err.Error())
	}
}

func createTlsSecret(sc *ServiceContext) {
	k := &sc.Kubernetes
	alternateDNS = fmt.Sprintf("%s.%s.svc", k.Name, k.Namespace)
	cert, key, _ = certuitl.GenerateSelfSignedCertKey(alternateDNS, nil, []string{alternateDNS})
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sc.Kubernetes.Name,
			Namespace: sc.Kubernetes.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       cert,
			corev1.TLSPrivateKeyKey: key,
		},
	}
	_, err := sc.Kubernetes.Client.CoreV1().Secrets(sc.Kubernetes.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logx.Errorf("%s", err.Error())
	}
}

func writeTLSFiles(sc *ServiceContext) {
	certDir, _ := filepath.Split(sc.Config.CertFile)
	keyDir, _ := filepath.Split(sc.Config.KeyFile)
	os.Remove(sc.Config.CertFile)
	os.Remove(sc.Config.KeyFile)
	os.MkdirAll(certDir, 0755)
	os.MkdirAll(keyDir, 0755)
	err := os.WriteFile(sc.Config.CertFile, cert, 0644)
	if err != nil {
		logx.Errorf("failed to write cert file: %v", err)
	}
	err = os.WriteFile(sc.Config.KeyFile, key, 0600)
	if err != nil {
		logx.Errorf("failed to write key file: %v", err)
	}
}

func getClient() *kubernetes.Clientset {
	config := ctrl.GetConfigOrDie()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logx.Errorf("%s", err.Error())
	}
	return clientset
}

func getNamespace() string {
	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		logx.Errorf("%v. Using \"default\"", err.Error())
		return config.DefaultNamespace
	}
	ns := strings.TrimSpace(string(namespace))
	if ns == "" {
		logx.Errorf("Namespace is empty. Using \"default\"")
		return config.DefaultNamespace
	}
	return ns
}

func createConfigMap(sc *ServiceContext) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sc.Kubernetes.Name,
			Namespace: sc.Kubernetes.Namespace,
		},
		Data: config.HostMap,
	}
	_, err := sc.Kubernetes.Client.CoreV1().ConfigMaps(sc.Kubernetes.Namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		logx.Errorf("%s", err.Error())
	}
}

func watchConfigMap(sc *ServiceContext) {
	listWatch := cache.NewListWatchFromClient(
		sc.Kubernetes.Client.CoreV1().RESTClient(),
		"configmaps",
		sc.Kubernetes.Namespace,
		fields.OneTermEqualSelector("metadata.name", sc.Kubernetes.Name),
	)
	informer := cache.NewSharedIndexInformer(
		listWatch,
		&corev1.ConfigMap{},
		0,
		cache.Indexers{},
	)
	const message = "ConfigMap changed"
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logx.Info(message)
			sc.Kubernetes.HostMap, _ = sc.Kubernetes.Client.CoreV1().ConfigMaps(sc.Kubernetes.Namespace).Get(context.TODO(), sc.Kubernetes.Name, metav1.GetOptions{})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			logx.Info(message)
			sc.Kubernetes.HostMap, _ = sc.Kubernetes.Client.CoreV1().ConfigMaps(sc.Kubernetes.Namespace).Get(context.TODO(), sc.Kubernetes.Name, metav1.GetOptions{})
		},
		DeleteFunc: func(obj interface{}) {
			logx.Info(message)
			sc.Kubernetes.HostMap, _ = sc.Kubernetes.Client.CoreV1().ConfigMaps(sc.Kubernetes.Namespace).Get(context.TODO(), sc.Kubernetes.Name, metav1.GetOptions{})
		},
	})
	go informer.Run(wait.NeverStop)
}
