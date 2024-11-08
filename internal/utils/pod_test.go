// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"
	"time"

	"github.com/SlinkyProject/slurm-operator/internal/annotations"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodByCreationTimestampAndPhase_Len(t *testing.T) {
	tests := []struct {
		name string
		o    PodByCreationTimestampAndPhase
		want int
	}{
		{
			name: "empty list",
			o:    []*corev1.Pod{},
			want: 0,
		},
		{
			name: "single list",
			o:    append([]*corev1.Pod{}, &corev1.Pod{}),
			want: 1,
		},
		{
			name: "double list",
			o:    append([]*corev1.Pod{}, &corev1.Pod{}, &corev1.Pod{}),
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.Len(); got != tt.want {
				t.Errorf("PodByCreationTimestampAndPhase.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodByCreationTimestampAndPhase_Swap(t *testing.T) {
	podA := corev1.Pod{}
	podA.SetName("A")
	podB := corev1.Pod{}
	podB.SetName("B")
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		o    PodByCreationTimestampAndPhase
		args args
		want string
	}{
		{
			name: "swap",
			o:    append([]*corev1.Pod{}, &podA, &podB),
			args: args{
				i: 0,
				j: 1,
			},
			want: podB.GetName(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.o.Swap(tt.args.i, tt.args.j)
			if tt.o[0].GetName() != tt.want {
				t.Errorf("o[0].GetName = %v, wanted %v", tt.o[0].GetName(), tt.want)
			}
		})
	}
}

func TestPodByCreationTimestampAndPhase_Less(t *testing.T) {
	var podA, podB, podC, podD corev1.Pod
	podA.SetCreationTimestamp(metav1.Now())
	podA.Spec.NodeName = "nodeA"
	podB.SetCreationTimestamp(metav1.Now())
	podA.Spec.NodeName = ""
	podC.SetCreationTimestamp(metav1.Now())
	podA.Spec.NodeName = "nodeC"
	podD.SetCreationTimestamp(podC.GetCreationTimestamp())
	podA.Spec.NodeName = "nodeD"
	pods := append([]*corev1.Pod{}, &podA, &podB, &podC, &podD)
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		o    PodByCreationTimestampAndPhase
		args args
		want bool
	}{
		{
			name: "Is 1 less than 99",
			o:    pods,
			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "Is 99 less than 1",
			o:    pods,
			args: args{
				i: 1,
				j: 0,
			},
			want: false,
		},
		{
			name: "Is 99 less than 99 (by CreationTimestamp)",
			o:    pods,
			args: args{
				i: 1,
				j: 2,
			},
			want: true,
		},
		{
			name: "Is 99 less than 99 (by CreationTimestamp)",
			o:    pods,
			args: args{
				i: 2,
				j: 1,
			},
			want: false,
		},
		{
			name: "Is 99 less than 99 (by Name)",
			o:    pods,
			args: args{
				i: 2,
				j: 3,
			},
			want: false,
		},
		{
			name: "Is 99 less than 99 (by Name)",
			o:    pods,
			args: args{
				i: 3,
				j: 2,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("PodByCreationTimestampAndPhase.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodByCost_Len(t *testing.T) {
	tests := []struct {
		name string
		o    PodByCost
		want int
	}{
		{
			name: "empty list",
			o:    []*corev1.Pod{},
			want: 0,
		},
		{
			name: "single list",
			o:    append([]*corev1.Pod{}, &corev1.Pod{}),
			want: 1,
		},
		{
			name: "double list",
			o:    append([]*corev1.Pod{}, &corev1.Pod{}, &corev1.Pod{}),
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.Len(); got != tt.want {
				t.Errorf("PodByCost.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodByCost_Swap(t *testing.T) {
	podA := corev1.Pod{}
	podA.SetName("A")
	podB := corev1.Pod{}
	podB.SetName("B")
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		o    PodByCost
		args args
		want string
	}{
		{
			name: "swap",
			o:    append([]*corev1.Pod{}, &podA, &podB),
			args: args{
				i: 0,
				j: 1,
			},
			want: podB.GetName(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.o.Swap(tt.args.i, tt.args.j)
		})
	}
}

func TestPodByCost_Less(t *testing.T) {
	var podA, podB, podC, podD corev1.Pod
	podA.SetCreationTimestamp(metav1.Now())
	podA.SetName("A")
	podA.SetAnnotations(map[string]string{
		annotations.PodDeletionCost: "1",
	})
	podB.SetCreationTimestamp(metav1.Now())
	podA.SetName("B")
	podB.SetAnnotations(map[string]string{
		annotations.PodDeletionCost: "99",
	})
	podC.SetCreationTimestamp(metav1.Now())
	podA.SetName("C")
	podC.SetAnnotations(map[string]string{
		annotations.PodDeletionCost: "99",
	})
	podD.SetCreationTimestamp(podC.GetCreationTimestamp())
	podD.SetName("D")
	podD.SetAnnotations(map[string]string{
		annotations.PodDeletionCost: "99",
	})
	pods := append([]*corev1.Pod{}, &podA, &podB, &podC, &podD)
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		o    PodByCost
		args args
		want bool
	}{
		{
			name: "Is 1 less than 99",
			o:    pods,
			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "Is 99 less than 1",
			o:    pods,
			args: args{
				i: 1,
				j: 0,
			},
			want: false,
		},
		{
			name: "Is 99 less than 99 (by CreationTimestamp)",
			o:    pods,
			args: args{
				i: 1,
				j: 2,
			},
			want: true,
		},
		{
			name: "Is 99 less than 99 (by CreationTimestamp)",
			o:    pods,
			args: args{
				i: 2,
				j: 1,
			},
			want: false,
		},
		{
			name: "Is 99 less than 99 (by Name)",
			o:    pods,
			args: args{
				i: 2,
				j: 3,
			},
			want: true,
		},
		{
			name: "Is 99 less than 99 (by Name)",
			o:    pods,
			args: args{
				i: 3,
				j: 2,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("PodByCost.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRunningAndReady(t *testing.T) {
	var podA, podB corev1.Pod
	podA.Status.Phase = corev1.PodRunning
	podA.Status.Conditions = append(podA.Status.Conditions, corev1.PodCondition{
		Type:   corev1.PodReady,
		Status: corev1.ConditionTrue,
	})
	podB.Status.Phase = corev1.PodFailed
	podB.Status.Conditions = append(podA.Status.Conditions, corev1.PodCondition{
		Type:   corev1.PodReady,
		Status: corev1.ConditionFalse,
	})
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "podA should be Running and Ready",
			args: args{
				pod: &podA,
			},
			want: true,
		},
		{
			name: "podB should not be Running and Ready",
			args: args{
				pod: &podB,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRunningAndReady(tt.args.pod); got != tt.want {
				t.Errorf("IsRunningAndReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newPod(now metav1.Time, ready bool, beforeSec int) *corev1.Pod {
	conditionStatus := corev1.ConditionFalse
	if ready {
		conditionStatus = corev1.ConditionTrue
	}
	return &corev1.Pod{
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{
				{
					Type:               corev1.PodReady,
					LastTransitionTime: metav1.NewTime(now.Time.Add(-1 * time.Duration(beforeSec) * time.Second)),
					Status:             conditionStatus,
				},
			},
		},
	}
}
func TestIsRunningAndAvailable(t *testing.T) {
	type args struct {
		pod             *corev1.Pod
		minReadySeconds int32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Not ready before 0",
			args: args{
				pod:             newPod(metav1.Now(), false, 0),
				minReadySeconds: 0,
			},
			want: false,
		},
		{
			name: "Ready before 0",
			args: args{
				pod:             newPod(metav1.Now(), true, 0),
				minReadySeconds: 1,
			},
			want: false,
		},
		{
			name: "Ready 0",
			args: args{
				pod:             newPod(metav1.Now(), true, 0),
				minReadySeconds: 0,
			},
			want: true,
		},
		{
			name: "Ready after 50",
			args: args{
				pod:             newPod(metav1.Now(), true, 51),
				minReadySeconds: 50,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRunningAndAvailable(tt.args.pod, tt.args.minReadySeconds); got != tt.want {
				t.Errorf("IsRunningAndAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCreated(t *testing.T) {
	var podA, podB corev1.Pod
	podA.Status.Phase = corev1.PodRunning
	podB.Status.Phase = ""
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "podA should not be created",
			args: args{
				pod: &podA,
			},
			want: true,
		},
		{
			name: "podB should not be created",
			args: args{
				pod: &podB,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCreated(tt.args.pod); got != tt.want {
				t.Errorf("IsCreated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPending(t *testing.T) {
	var podA, podB corev1.Pod
	podA.Status.Phase = corev1.PodPending
	podB.Status.Phase = ""
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "podA should be Pending",
			args: args{
				pod: &podA,
			},
			want: true,
		},
		{
			name: "podB should not be Pending",
			args: args{
				pod: &podB,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPending(tt.args.pod); got != tt.want {
				t.Errorf("IsPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFailed(t *testing.T) {
	var podA, podB corev1.Pod
	podA.Status.Phase = corev1.PodFailed
	podB.Status.Phase = ""
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "podA should be Failed",
			args: args{
				pod: &podA,
			},
			want: true,
		},
		{
			name: "podB should not be Failed",
			args: args{
				pod: &podB,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFailed(tt.args.pod); got != tt.want {
				t.Errorf("IsFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSucceeded(t *testing.T) {
	var podA, podB corev1.Pod
	podA.Status.Phase = corev1.PodSucceeded
	podB.Status.Phase = ""
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "podA should be Succeeded",
			args: args{
				pod: &podA,
			},
			want: true,
		},
		{
			name: "podB should not be Succeeded",
			args: args{
				pod: &podB,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSucceeded(tt.args.pod); got != tt.want {
				t.Errorf("IsSucceeded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTerminating(t *testing.T) {
	var podA, podB corev1.Pod
	timestamp := metav1.Now()
	podA.SetDeletionTimestamp(&timestamp)
	podB.DeletionTimestamp = nil
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "podA should be Terminating",
			args: args{
				pod: &podA,
			},
			want: true,
		},
		{
			name: "podB should not be Terminating",
			args: args{
				pod: &podB,
			},
			want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTerminating(tt.args.pod); got != tt.want {
				t.Errorf("IsTerminating() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHealthy(t *testing.T) {
	var podA, podB, podC corev1.Pod
	podA.Status.Phase = corev1.PodRunning
	podA.Status.Conditions = append(podA.Status.Conditions, corev1.PodCondition{
		Type:   corev1.PodReady,
		Status: corev1.ConditionTrue,
	})
	podA.DeletionTimestamp = nil
	podB.Status.Phase = corev1.PodFailed
	podB.Status.Conditions = append(podB.Status.Conditions, corev1.PodCondition{
		Type:   corev1.PodReady,
		Status: corev1.ConditionTrue,
	})
	podC.Status.Phase = corev1.PodFailed
	podC.Status.Conditions = append(podC.Status.Conditions, corev1.PodCondition{
		Type:   corev1.PodReady,
		Status: corev1.ConditionTrue,
	})
	timestamp := metav1.Now()
	podC.SetDeletionTimestamp(&timestamp)
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "podA should be Healthy",
			args: args{
				pod: &podA,
			},
			want: true,
		},
		{
			name: "podB should not be Healthy",
			args: args{
				pod: &podB,
			},
			want: false,
		},
		{
			name: "podC should not be Healthy",
			args: args{
				pod: &podC,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHealthy(tt.args.pod); got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}
