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

package unload

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	alluxiov1alpha1 "github.com/alluxio/k8s-operator/api/v1alpha1"
	alluxioClusterPkg "github.com/alluxio/k8s-operator/pkg/alluxiocluster"
	datasetPkg "github.com/alluxio/k8s-operator/pkg/dataset"
	"github.com/alluxio/k8s-operator/pkg/logger"
)

type UnloadReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type UnloadReconcilerReqCtx struct {
	alluxioClusterPkg.AlluxioClusterer
	Unloader
	client.Client
	context.Context
	types.NamespacedName
}

func (r *UnloadReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx := &UnloadReconcilerReqCtx{
		Client:         r.Client,
		Context:        context,
		NamespacedName: req.NamespacedName,
	}
	unload := &alluxiov1alpha1.Unload{}
	ctx.Unloader = unload
	if err := FetchUnLoadFromK8sApiServer(r, req.NamespacedName, unload); err != nil {
		return ctrl.Result{}, err
	}

	if unload.ObjectMeta.UID == "" {
		return ctrl.Result{}, nil
	}

	dataset := &alluxiov1alpha1.Dataset{}
	datasetNamespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      *unload.Spec.Dataset,
	}
	if err := datasetPkg.FetchDatasetFromK8sApiServer(r, datasetNamespacedName, dataset); err != nil {
		return ctrl.Result{}, err
	}

	if dataset.Status.Phase != alluxiov1alpha1.DatasetPhaseReady {
		unload.Status.Phase = alluxiov1alpha1.UnloadPhaseNotExist
		return UpdateUnloadStatus(ctx)
	}

	alluxioCluster := &alluxiov1alpha1.AlluxioCluster{}
	ctx.AlluxioClusterer = alluxioCluster
	alluxioNamespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      *dataset.Status.BoundedAlluxioCluster,
	}
	if err := alluxioClusterPkg.FetchAlluxioClusterFromK8sApiServer(r, alluxioNamespacedName, alluxioCluster); err != nil {
		return ctrl.Result{}, err
	}

	switch unload.Status.Phase {
	case alluxiov1alpha1.UnloadPhaseNone, alluxiov1alpha1.UnloadPhaseNotExist:
		return Unload(ctx)
	default:
		return ctrl.Result{}, nil
	}
}

func FetchUnLoadFromK8sApiServer(r client.Reader, namespacedName types.NamespacedName, unload *alluxiov1alpha1.Unload) error {
	if err := r.Get(context.TODO(), namespacedName, unload); err != nil {
		if errors.IsNotFound(err) {
			logger.Infof("Unload object %v not found. It is being deleted or already deleted.", namespacedName.String())
		} else {
			logger.Errorf("Failed to get unload job %v: %v", namespacedName.String(), err)
			return err
		}
	}
	return nil
}

func UpdateUnloadStatus(ctx *UnloadReconcilerReqCtx) (ctrl.Result, error) {
	if err := ctx.Update(ctx.Context, ctx.Unloader); err != nil {
		logger.Errorf("Failed updating unload job status: %v", err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UnloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alluxiov1alpha1.Unload{}).
		Complete(r)
}
