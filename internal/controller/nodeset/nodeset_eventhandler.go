// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package nodeset

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	slurmclient "github.com/SlinkyProject/slurm-client/pkg/client"
	slurmtypes "github.com/SlinkyProject/slurm-client/pkg/types"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/utils"
	"github.com/SlinkyProject/slurm-operator/internal/utils/podinfo"
)

var _ handler.EventHandler = &podEventHandler{}

type podEventHandler struct {
	client.Reader
}

func enqueueNodeSet(q workqueue.TypedRateLimitingInterface[reconcile.Request], nodeset *slinkyv1alpha1.NodeSet) {
	q.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      nodeset.GetName(),
			Namespace: nodeset.GetNamespace(),
		},
	})
}

func (e *podEventHandler) Create(
	ctx context.Context,
	evt event.CreateEvent,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	logger := log.FromContext(ctx)
	pod := evt.Object.(*corev1.Pod)
	if utils.IsTerminating(pod) {
		// on a restart of the controller manager, it's possible a new pod shows up in a state that
		// is already pending deletion. Prevent the pod from being a creation observation.
		e.Delete(ctx, event.DeleteEvent{Object: evt.Object}, q)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(pod); controllerRef != nil {
		nodeset := e.resolveControllerRef(pod.Namespace, controllerRef)
		if nodeset == nil {
			return
		}
		logger.V(1).Info("Pod added", "Pod", klog.KObj(pod))
		enqueueNodeSet(q, nodeset)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching NodeSets and sync
	// them to see if anyone wants to adopt it.
	// DO NOT observe creation because no controller should be waiting for an
	// orphan.
	nodesetList := e.getPodNodeSets(ctx, pod)
	if len(nodesetList) == 0 {
		return
	}
	logger.V(1).Info("Orphan Pod created, matched Node owners",
		"Pod", klog.KObj(pod), "Nodes", nodesetList)
	for _, nodeset := range nodesetList {
		enqueueNodeSet(q, nodeset)
	}
}

func (e *podEventHandler) Update(
	ctx context.Context,
	evt event.UpdateEvent,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	logger := log.FromContext(ctx)
	oldPod := evt.ObjectOld.(*corev1.Pod)
	curPod := evt.ObjectNew.(*corev1.Pod)
	if curPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}

	curControllerRef := metav1.GetControllerOf(curPod)
	oldControllerRef := metav1.GetControllerOf(oldPod)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if nodeset := e.resolveControllerRef(oldPod.Namespace, oldControllerRef); nodeset != nil {
			enqueueNodeSet(q, nodeset)
		}
	}

	if curPod.DeletionTimestamp != nil {
		// when a pod is deleted gracefully its deletion timestamp is first modified to reflect a grace period,
		// and after such time has passed, the kubelet actually deletes it from the store. We receive an update
		// for modification of the deletion timestamp and expect a NodeSet to create more Pods asap, not wait
		// until the kubelet actually deletes the pod.
		e.deletePod(ctx, curPod, q, false)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		nodeset := e.resolveControllerRef(curPod.Namespace, curControllerRef)
		if nodeset == nil {
			return
		}
		logger.V(1).Info("Pod updated with NodeSet owner",
			"Pod", klog.KObj(curPod), "NodeSet", klog.KObj(nodeset))
		enqueueNodeSet(q, nodeset)
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	nodesetList := e.getPodNodeSets(ctx, curPod)
	if len(nodesetList) == 0 {
		return
	}
	logger.V(1).Info("Orphan Pod updated, matched Nodes",
		"Pod", klog.KObj(curPod), "Nodes", nodesetList)
	labelChanged := !reflect.DeepEqual(curPod.Labels, oldPod.Labels)
	if labelChanged || controllerRefChanged {
		for _, nodeset := range nodesetList {
			enqueueNodeSet(q, nodeset)
		}
	}
}

func (e *podEventHandler) Delete(
	ctx context.Context,
	evt event.DeleteEvent,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	logger := log.FromContext(ctx)
	pod, ok := evt.Object.(*corev1.Pod)
	if !ok {
		logger.Error(nil, "DeleteEvent parse pod failed",
			"DeleteStateUnknown", evt.DeleteStateUnknown,
			"Object", klog.KObj(evt.Object))
		return
	}
	e.deletePod(ctx, pod, q, true)
}

func (e *podEventHandler) deletePod(
	ctx context.Context,
	pod *corev1.Pod,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
	isDeleted bool,
) {
	logger := log.FromContext(ctx)
	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	nodeset := e.resolveControllerRef(pod.Namespace, controllerRef)
	if nodeset == nil {
		return
	}
	if isDeleted {
		logger.V(1).Info("NodeSet Pod deleted",
			"Pod", klog.KObj(pod), "NodeSet", klog.KObj(nodeset))
	} else {
		logger.V(1).Info("NodeSet Pod terminating",
			"Pod", klog.KObj(pod), "NodeSet", klog.KObj(nodeset))
	}
	enqueueNodeSet(q, nodeset)
}

func (e *podEventHandler) Generic(
	ctx context.Context,
	evt event.GenericEvent,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	pod := evt.Object.(*corev1.Pod)
	namespacedName := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}
	if err := e.Get(ctx, namespacedName, pod); err != nil {
		return
	}

	nodesetList := e.getPodNodeSets(ctx, pod)
	for _, nodeset := range nodesetList {
		if !isPodFromNodeSet(nodeset, pod) {
			continue
		}
		enqueueNodeSet(q, nodeset)
	}
}

func (e *podEventHandler) resolveControllerRef(
	namespace string,
	controllerRef *metav1.OwnerReference,
) *slinkyv1alpha1.NodeSet {
	if controllerRef.Kind != controllerKind.Kind || controllerRef.APIVersion != controllerKind.GroupVersion().String() {
		return nil
	}

	nodeset := &slinkyv1alpha1.NodeSet{}
	if err := e.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: controllerRef.Name}, nodeset); err != nil {
		return nil
	}
	if nodeset.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return nodeset
}

func (e *podEventHandler) getPodNodeSets(ctx context.Context, pod *corev1.Pod) []*slinkyv1alpha1.NodeSet {
	logger := log.FromContext(ctx)
	nodesetList := slinkyv1alpha1.NodeSetList{}
	if err := e.List(context.TODO(), &nodesetList, client.InNamespace(pod.Namespace)); err != nil {
		return nil
	}

	var nsMatched []*slinkyv1alpha1.NodeSet
	for i := range nodesetList.Items {
		nodeset := &nodesetList.Items[i]
		selector, err := metav1.LabelSelectorAsSelector(nodeset.Spec.Selector)
		if err != nil || selector.Empty() || !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}

		nsMatched = append(nsMatched, nodeset)
	}

	if len(nsMatched) > 1 {
		// ControllerRef will ensure we do not do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		logger.Info("More than one NodeSet is selecting Pod",
			"Pod", klog.KObj(pod), "NodeSets", nsMatched)
	}
	return nsMatched
}

var _ handler.EventHandler = &nodeEventHandler{}

type nodeEventHandler struct {
	reader client.Reader
}

func (e *nodeEventHandler) Create(
	ctx context.Context,
	evt event.CreateEvent,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	logger := log.FromContext(ctx)
	nodesetList := &slinkyv1alpha1.NodeSetList{}
	err := e.reader.List(context.TODO(), nodesetList)
	if err != nil {
		logger.V(1).Error(err, "Error enqueueing NodeSets")
		return
	}

	node := evt.Object.(*corev1.Node)
	for i := range nodesetList.Items {
		nodeset := &nodesetList.Items[i]
		if shouldSchedule, _ := nodeShouldRunNodeSetPod(node, nodeset); shouldSchedule {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      nodeset.GetName(),
				Namespace: nodeset.GetNamespace(),
			}})
		}
	}
}

func (e *nodeEventHandler) Update(
	ctx context.Context,
	evt event.UpdateEvent,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	logger := log.FromContext(ctx)
	oldNode := evt.ObjectOld.(*corev1.Node)
	curNode := evt.ObjectNew.(*corev1.Node)
	if shouldIgnoreNodeUpdate(*oldNode, *curNode) {
		return
	}

	nodesetList := &slinkyv1alpha1.NodeSetList{}
	err := e.reader.List(context.TODO(), nodesetList)
	if err != nil {
		logger.Error(err, "Error listing NodeSets")
		return
	}
	// TODO: it'd be nice to pass a hint with these enqueues, so that each nodeset would only examine the added node (unless it has other work to do, too).
	for i := range nodesetList.Items {
		nodeset := &nodesetList.Items[i]
		oldShouldRun, oldShouldContinueRunning := nodeShouldRunNodeSetPod(oldNode, nodeset)
		currentShouldRun, currentShouldContinueRunning := nodeShouldRunNodeSetPod(curNode, nodeset)
		if (oldShouldRun != currentShouldRun) || (oldShouldContinueRunning != currentShouldContinueRunning) {
			logger.V(1).Info("Node update triggers NodeSet to reconcile.",
				"Node", klog.KObj(curNode), "NodeSet", klog.KObj(nodeset))
			q.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      nodeset.GetName(),
					Namespace: nodeset.GetNamespace(),
				},
			})
		}
	}
}

func (e *nodeEventHandler) Delete(
	ctx context.Context,
	evt event.DeleteEvent,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	// Intentionally empty
}

func (e *nodeEventHandler) Generic(
	ctx context.Context,
	evt event.GenericEvent,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	// Intentionally empty
}

func shouldIgnoreNodeUpdate(oldNode, curNode corev1.Node) bool {
	if !nodeInSameCondition(oldNode.Status.Conditions, curNode.Status.Conditions) {
		return false
	}
	oldNode.ResourceVersion = curNode.ResourceVersion
	oldNode.Status.Conditions = curNode.Status.Conditions
	return apiequality.Semantic.DeepEqual(oldNode, curNode)
}

// SetEventHandler is a helper function to make slurm node updates propagate to
// the nodeset controller via configured event channel.
func SetEventHandler(client slurmclient.Client, eventCh chan event.GenericEvent) {
	informer := client.GetInformer(slurmtypes.ObjectTypeV0041Node)
	informer.SetEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*slurmtypes.V0041Node)
			if !ok {
				return
			}
			podInfo := podinfo.PodInfo{}
			_ = podinfo.ParseIntoPodInfo(node.Comment, &podInfo)
			eventCh <- podEvent(podInfo)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldNode, ok := oldObj.(*slurmtypes.V0041Node)
			if !ok {
				return
			}
			newNode, ok := newObj.(*slurmtypes.V0041Node)
			if !ok {
				return
			}
			if apiequality.Semantic.DeepEqual(newNode.State, oldNode.State) {
				return
			}
			podInfo := podinfo.PodInfo{}
			_ = podinfo.ParseIntoPodInfo(newNode.Comment, &podInfo)
			eventCh <- podEvent(podInfo)
		},
		DeleteFunc: func(obj interface{}) {
			node, ok := obj.(*slurmtypes.V0041Node)
			if !ok {
				return
			}
			podInfo := podinfo.PodInfo{}
			_ = podinfo.ParseIntoPodInfo(node.Comment, &podInfo)
			eventCh <- podEvent(podInfo)
		},
	})
}

func podEvent(podInfo podinfo.PodInfo) event.GenericEvent {
	return event.GenericEvent{
		Object: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: podInfo.Namespace,
				Name:      podInfo.PodName,
			},
		},
	}
}
