/*
 * The Alluxio Open Foundation licenses this work under the Apache License, version 2.0
 * (the "License"). You may not use this work except in compliance with the License, which is
 * available at www.apache.org/licenses/LICENSE-2.0
 *
 * This software is distributed on an "AS IS" basis, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied, as more fully set forth in the License.
 *
 * See the NOTICE file distributed with this work for information regarding copyright ownership.
 */

package dataset

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	alluxiov1alpha1 "github.com/alluxio/k8s-operator/api/v1alpha1"
	"github.com/alluxio/k8s-operator/pkg/logger"
)

// DatasetReconciler reconciles a Dataset object
type DatasetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type DatasetReconcilerReqCtx struct {
	Datasetter
	client.Client
	context.Context
	types.NamespacedName
}

func (r *DatasetReconciler) Reconcile(context context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	ctx := &DatasetReconcilerReqCtx{
		Client:         r.Client,
		Context:        context,
		NamespacedName: req.NamespacedName,
	}

	dataset := &alluxiov1alpha1.Dataset{}
	ctx.Datasetter = dataset
	if err = FetchDatasetFromK8sApiServer(r, req.NamespacedName, dataset); err != nil {
		return
	}

	if dataset.ObjectMeta.UID == "" {
		return DeleteDatasetIfExist(req)
	}
	if dataset.Status.Phase == alluxiov1alpha1.DatasetPhaseNone {
		dataset.Status.Phase = alluxiov1alpha1.DatasetPhasePending
		return UpdateDatasetStatus(ctx)
	}
	return
}

func FetchDatasetFromK8sApiServer(r client.Reader, namespacedName types.NamespacedName, dataset *alluxiov1alpha1.Dataset) error {
	if err := r.Get(context.TODO(), namespacedName, dataset); err != nil {
		if errors.IsNotFound(err) {
			dataset.Status.Phase = alluxiov1alpha1.DatasetPhaseNotExist
		}
		if !errors.IsNotFound(err) {
			logger.Errorf("Failed to get Dataset %s: %v", namespacedName.String(), err)
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatasetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alluxiov1alpha1.Dataset{}).
		Complete(r)
}
