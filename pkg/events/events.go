package events

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// Emit emits event. Both k8s event and cloud event
func Emit(c client.Client, obj runtime.Object, evType, reason, message string) error {
	ref, err := getObjectReference(obj)
	if err != nil {
		return err
	}

	// Emit K8s event
	if err := emitK8sEvent(c, ref, evType, reason, message); err != nil {
		return err
	}

	// TODO - Emit cloud event

	return nil
}

func emitK8sEvent(c client.Client, ref *corev1.ObjectReference, evType, reason, message string) error {
	t := metav1.Time{Time: time.Now()}
	ev := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.%x", ref.Name, t.UnixNano()),
			Namespace: ref.Namespace,
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		FirstTimestamp: t,
		LastTimestamp:  t,
		Count:          1,
		Type:           evType,
		Source: corev1.EventSource{
			Component: ref.Kind,
		},
	}

	if err := c.Create(context.Background(), ev); err != nil {
		return err
	}

	return nil
}

func getObjectReference(obj runtime.Object) (*corev1.ObjectReference, error) {
	if obj == nil {
		return nil, fmt.Errorf("obj is nil")
	}

	if ref, ok := obj.(*corev1.ObjectReference); ok {
		return ref, nil
	}

	gvk := obj.GetObjectKind().GroupVersionKind()

	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}

	return &corev1.ObjectReference{
		APIVersion:      gvk.GroupVersion().String(),
		Kind:            gvk.Kind,
		Name:            objMeta.GetName(),
		Namespace:       objMeta.GetNamespace(),
		UID:             objMeta.GetUID(),
		ResourceVersion: objMeta.GetResourceVersion(),
	}, nil
}
