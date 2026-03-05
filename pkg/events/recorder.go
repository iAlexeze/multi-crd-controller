package events

import (
	"context"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

type Recorder struct {
	name        string
	kube        *kubeclient.Kubeclient
	scheme      *runtime.Scheme
	broadcaster record.EventBroadcaster
	recorder    record.EventRecorder
	opts        Options
}

type Options struct {
	Component string
}

var _ domain.Component = (*Recorder)(nil)

func NewRecorder(kube *kubeclient.Kubeclient, scheme *runtime.Scheme, opts Options) *Recorder {
	if opts.Component == "" {
		opts.Component = "project-controller"
	}

	return &Recorder{
		name:   "event handler",
		kube:   kube,
		scheme: scheme,
		opts:   opts,
	}
}

func (r *Recorder) Start(ctx context.Context) error {
	r.broadcaster = record.NewBroadcaster(record.WithContext(ctx))
	r.broadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: r.kube.Clientset().CoreV1().Events(""),
		})

	r.recorder = r.broadcaster.NewRecorder(
		r.scheme,
		corev1.EventSource{
			Component: r.opts.Component,
		})
	return nil
}

func (r *Recorder) Shutdown(ctx context.Context) {
	if r.broadcaster != nil {
		r.broadcaster.Shutdown()
	}
}

func (r *Recorder) Name() string {
	return r.name
}

func (r *Recorder) Broadcaster() record.EventBroadcaster {
	return r.broadcaster
}

func (r *Recorder) Recorder() record.EventRecorder {
	return r.recorder
}
