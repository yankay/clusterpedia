package clustersynchro

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	clusterv1alpha2 "github.com/clusterpedia-io/api/cluster/v1alpha2"
)

func (synchro *ClusterSynchro) monitor() {
	klog.V(2).InfoS("Cluster Synchro Monitor Running...", "cluster", synchro.name)

	wait.JitterUntil(synchro.checkClusterHealthy, 5*time.Second, 0.5, false, synchro.closer)

	healthyCondition := metav1.Condition{
		Type:               clusterv1alpha2.ClusterHealthyCondition,
		Status:             metav1.ConditionUnknown,
		Reason:             clusterv1alpha2.ClusterMonitorStopReason,
		Message:            "cluster synchro is shutdown",
		LastTransitionTime: metav1.Now().Rfc3339Copy(),
	}
	if lastReadyCondition := synchro.healthyCondition.Load().(metav1.Condition); lastReadyCondition.Status == metav1.ConditionFalse {
		healthyCondition.Message = fmt.Sprintf("Last Condition Reason: %s, Message: %s", lastReadyCondition.Reason, lastReadyCondition.Message)
	}
	synchro.healthyCondition.Store(healthyCondition)
}

func (synchro *ClusterSynchro) checkClusterHealthy() {
	defer synchro.updateStatus()
	lastReadyCondition := synchro.healthyCondition.Load().(metav1.Condition)

	if ready, err := checkKubeHealthy(synchro.clusterclient); !ready {
		// if the last status was not ConditionTrue, stop resource synchros
		if lastReadyCondition.Status != metav1.ConditionTrue {
			synchro.stopResourceSynchro()
		}

		condition := metav1.Condition{
			Type:    clusterv1alpha2.ClusterHealthyCondition,
			Status:  metav1.ConditionFalse,
			Reason:  clusterv1alpha2.ClusterUnhealthyReason,
			Message: "cluster health responded without ok",
		}
		if err != nil {
			condition.Reason = clusterv1alpha2.ClusterNotReachableReason
			condition.Message = err.Error()
		}

		if lastReadyCondition.Status != condition.Status || lastReadyCondition.Reason != condition.Reason || lastReadyCondition.Message != condition.Message {
			condition.LastTransitionTime = metav1.Now().Rfc3339Copy()
			synchro.healthyCondition.Store(condition)
		}
		return
	}

	synchro.startResourceSynchro()
	if lastReadyCondition.Status == metav1.ConditionTrue {
		// TODO: if lastReadyCondition.Message != "", need process
		return
	}

	condition := metav1.Condition{
		Type:               clusterv1alpha2.ClusterHealthyCondition,
		Status:             metav1.ConditionTrue,
		Reason:             clusterv1alpha2.ClusterHealthyReason,
		Message:            "cluster health responded with ok",
		LastTransitionTime: metav1.Now().Rfc3339Copy(),
	}
	defer func() {
		synchro.healthyCondition.Store(condition)
	}()

	if _, err := synchro.dynamicDiscoveryManager.GetAndFetchServerVersion(); err != nil {
		condition.Message = fmt.Sprintf("cluster health responded with ok, but get server version: %v", err)
		klog.ErrorS(err, "Failed to get cluster version", "cluster", synchro.name)
	}
}

// TODO(iceber): resolve for more detailed error
func checkKubeHealthy(client kubernetes.Interface) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	_, err := client.Discovery().RESTClient().Get().AbsPath("/readyz").DoRaw(ctx)
	if apierrors.IsNotFound(err) {
		_, err = client.Discovery().RESTClient().Get().AbsPath("/healthz").DoRaw(ctx)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
