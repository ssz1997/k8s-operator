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

package alluxiocluster

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	alluxiov1alpha1 "github.com/alluxio/k8s-operator/api/v1alpha1"
	datasetter "github.com/alluxio/k8s-operator/pkg/dataset"
	"github.com/alluxio/k8s-operator/pkg/finalizer"
	"github.com/alluxio/k8s-operator/pkg/logger"
)

// AlluxioClusterReconciler reconciles a AlluxioCluster object
type AlluxioClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type AlluxioClusterReconcileReqCtx struct {
	AlluxioClusterer
	client.Client
	context.Context
	datasetter.Datasetter
	types.NamespacedName
}

func (r *AlluxioClusterReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger.Infof("Reconciling AlluxioCluster %s", req.NamespacedName.String())
	ctx := &AlluxioClusterReconcileReqCtx{
		Client:         r.Client,
		Context:        context,
		NamespacedName: req.NamespacedName,
	}

	alluxioCluster := &alluxiov1alpha1.AlluxioCluster{}
	ctx.AlluxioClusterer = alluxioCluster
	if err := GetAlluxioClusterFromK8sApiServer(r, req.NamespacedName, alluxioCluster); err != nil {
		return ctrl.Result{}, err
	}

	dataset := &alluxiov1alpha1.Dataset{}
	ctx.Datasetter = dataset
	datasetNamespacedName := types.NamespacedName{
		Namespace: ctx.Namespace,
		Name:      alluxioCluster.Spec.Dataset,
	}
	if err := datasetter.GetDatasetFromK8sApiServer(r, datasetNamespacedName, dataset); err != nil {
		return ctrl.Result{}, err
	}

	if err := DeleteClusterIfNeeded(ctx); err != nil {
		return ctrl.Result{}, err
	}

	if err := CreateAlluxioClusterIfNeeded(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return UpdateStatus(ctx)
}

func GetAlluxioClusterFromK8sApiServer(r client.Reader, namespacedName types.NamespacedName, alluxioCluster AlluxioClusterer) error {
	if err := r.Get(context.TODO(), namespacedName, alluxioCluster); err != nil {
		if errors.IsNotFound(err) {
			logger.Infof("Alluxio cluster %s not found. It is being deleted or already deleted.", namespacedName.String())
		} else {
			logger.Errorf("Failed to get Alluxio cluster %s: %v", namespacedName.String(), err)
			return err
		}
	}
	return nil
}

func DeleteClusterIfNeeded(ctx *AlluxioClusterReconcileReqCtx) error {
	if ctx.AlluxioClusterer.IsDeleted() {
		if err := DeleteConfYamlFileIfExist(ctx.NamespacedName); err != nil {
			return err
		}
		if err := DeleteAlluxioClusterIfExist(ctx.NamespacedName); err != nil {
			return err
		}
		if ctx.Datasetter.GetStatus().Phase != alluxiov1alpha1.DatasetPhaseNotExist {
			ctx.Datasetter.GetStatus().Phase = alluxiov1alpha1.DatasetPhasePending
			ctx.Datasetter.GetStatus().BoundedAlluxioCluster = ""
			if err := updateDatasetStatus(ctx); err != nil {
				return err
			}
		}
		if err := finalizer.RemoveDummyFinalizerIfExist(ctx.Client, ctx.AlluxioClusterer, ctx.Context); err != nil {
			return err
		}
	}
	return nil
}

func CreateAlluxioClusterIfNeeded(ctx *AlluxioClusterReconcileReqCtx) error {
	if ctx.AlluxioClusterer.GetStatus().Phase == alluxiov1alpha1.ClusterPhaseNone || ctx.AlluxioClusterer.GetStatus().Phase == alluxiov1alpha1.ClusterPhasePending {
		if err := finalizer.AddDummyFinalizerIfNotExist(ctx.Client, ctx.AlluxioClusterer, ctx.Context); err != nil {
			return err
		}
		if err := CreateAlluxioClusterIfNotExist(ctx); err != nil {
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlluxioClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alluxiov1alpha1.AlluxioCluster{}).
		Complete(r)
}
