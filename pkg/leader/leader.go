package leader

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/events"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type leaderElection struct {
	name       string
	kube       *kubeclient.Kubeclient
	events     *events.Recorder
	cancelFunc context.CancelFunc
	run        func(context.Context)

	opts Options
}

type Options struct {
	LeaseDuration time.Duration
	RetryPeriod   time.Duration
	RenewDeadline time.Duration

	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

var _ domain.Component = (*leaderElection)(nil)

func NewLeaderElection(
	kube *kubeclient.Kubeclient,
	events *events.Recorder,
	run func(context.Context),
	opts Options,
) *leaderElection {
	if opts.Namespace == "" {
		opts.Namespace = "default"
	}

	return &leaderElection{
		name:   "resource-leader",
		events: events,
		kube:   kube,
		run:    run,
		opts:   opts,
	}
}

func (le *leaderElection) Start(ctx context.Context) error {
	// Create a cancellable context for the leader election
	leaderCtx, cancel := context.WithCancel(ctx)
	le.cancelFunc = cancel

	go func() {
		leaderelection.RunOrDie(leaderCtx, le.leaseConfig())
	}()
	return nil
}

func (le *leaderElection) Shutdown(ctx context.Context) {
	logger.Info().Msg("🛑 Shutting down leader election...")

	// Cancel the leader election context
	if le.cancelFunc != nil {
		le.cancelFunc()
	}

	// Give it a moment to release the lease
	utils.Sleep(2)
	logger.Info().Msg("✅ Leader election shut down")
}

func (le *leaderElection) Name() string {
	return le.name
}

func (le *leaderElection) kind() string {
	return "Lease"
}

// Helpers
// Lease configuration
func (le *leaderElection) leaseConfig() leaderelection.LeaderElectionConfig {
	return leaderelection.LeaderElectionConfig{
		Name:            le.Name(),
		Lock:            le.leaseLock(),
		LeaseDuration:   le.opts.LeaseDuration,
		RenewDeadline:   le.opts.RenewDeadline,
		RetryPeriod:     le.opts.RetryPeriod,
		ReleaseOnCancel: true,
		Callbacks:       le.callbacks(),
	}
}

// Lease lock
func (le *leaderElection) leaseLock() *resourcelock.LeaseLock {
	opts := le.opts
	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:        le.name,
			Namespace:   opts.Namespace,
			Annotations: opts.Annotations,
			Labels:      opts.Labels,
		},
		Client: le.kube.Clientset().CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      hostname(),
			EventRecorder: le.events.Recorder(),
		},
	}
}

// Build callbacks
func (le *leaderElection) callbacks() leaderelection.LeaderCallbacks {
	return leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			le.events.Recorder().Eventf(
				&corev1.ObjectReference{
					Name:      le.name,
					Namespace: le.opts.Namespace,
					Kind:      le.kind(),
				}, corev1.EventTypeNormal, "LeaderElected", "%s became leader", hostname(),
			)

			logger.Info().Msgf("%s 🏆 Became leader, starting controller...", hostname())
			le.run(ctx)
		},
		OnStoppedLeading: func() {
			le.events.Recorder().Eventf(
				&corev1.ObjectReference{
					Name:      le.name,
					Namespace: le.opts.Namespace,
					Kind:      le.kind(),
				}, corev1.EventTypeWarning, "LeaderLost", "%s lost leadership", hostname(),
			)
			logger.Info().Msgf("%s👋 Stopped leading - lease released", hostname())
		},
		OnNewLeader: func(identity string) {
			le.events.Recorder().Eventf(
				&corev1.ObjectReference{
					Name:      le.name,
					Namespace: le.opts.Namespace,
					Kind:      le.kind(),
				}, corev1.EventTypeNormal, "NewLeaderElected", "%s elected as leader", hostname(),
			)
			logger.Info().Msgf("👑 New leader elected: %s", identity)
		},
	}
}

// Get hostname
func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = uuid.New().String()
	}
	return hostname
}
