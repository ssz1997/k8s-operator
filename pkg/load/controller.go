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

package load

import (
	"context"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	alluxiov1alpha1 "github.com/alluxio/k8s-operator/api/v1alpha1"
	alluxioClusterPkg "github.com/alluxio/k8s-operator/pkg/alluxiocluster"
	datasetPkg "github.com/alluxio/k8s-operator/pkg/dataset"
	"github.com/alluxio/k8s-operator/pkg/logger"
	"github.com/alluxio/k8s-operator/pkg/utils"
)

type LoadReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type LoadReconcilerReqCtx struct {
	alluxioClusterPkg.AlluxioClusterer
	Loader
	client.Client
	context.Context
	types.NamespacedName
}

func (r *LoadReconciler) Reconcile(context context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	ctx := &LoadReconcilerReqCtx{
		Client:         r.Client,
		Context:        context,
		NamespacedName: req.NamespacedName,
	}
	load := &alluxiov1alpha1.Load{}
	ctx.Loader = load
	if err = FetchLoadFromK8sApiServer(r, req.NamespacedName, load); err != nil {
		return
	}

	if load.ObjectMeta.UID == "" {
		// TODO: shall we stop the load if still loading?
		return DeleteJob(ctx)
	}

	dataset := &alluxiov1alpha1.Dataset{}
	datasetNamespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      *load.Spec.Dataset,
	}
	if err = datasetPkg.FetchDatasetFromK8sApiServer(r, datasetNamespacedName, dataset); err != nil {
		return
	}

	if dataset.Status.Phase != alluxiov1alpha1.DatasetPhaseReady {
		load.Status.Phase = alluxiov1alpha1.LoadPhaseWaiting
		return UpdateLoadStatus(ctx)
	}

	alluxioCluster := &alluxiov1alpha1.AlluxioCluster{}
	ctx.AlluxioClusterer = alluxioCluster
	alluxioNamespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      *dataset.Status.BoundedAlluxioCluster,
	}
	if err = alluxioClusterPkg.FetchAlluxioClusterFromK8sApiServer(r, alluxioNamespacedName, alluxioCluster); err != nil {
		return
	}

	switch load.Status.Phase {
	case alluxiov1alpha1.LoadPhaseNone, alluxiov1alpha1.LoadPhaseWaiting:
		return CreateLoadJob(ctx)
	case alluxiov1alpha1.LoadPhaseLoading:
		return WaitLoadJobFinish(ctx)
	default:
		return
	}
}

func FetchLoadFromK8sApiServer(r client.Reader, namespacedName types.NamespacedName, load *alluxiov1alpha1.Load) error {
	if err := r.Get(context.TODO(), namespacedName, load); err != nil {
		if errors.IsNotFound(err) {
			logger.Infof("Load object %v not found. It is being deleted or already deleted.", namespacedName.String())
		} else {
			logger.Errorf("Failed to get load job %v: %v", namespacedName.String(), err)
			return err
		}
	}
	return nil
}

func WaitLoadJobFinish(ctx *LoadReconcilerReqCtx) (ctrl.Result, error) {
	loadJob, err := getLoadJob(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if loadJob.Status.Succeeded == 1 {
		ctx.Loader.GetStatus().Phase = alluxiov1alpha1.LoadPhaseLoaded
		if _, err := UpdateLoadStatus(ctx); err != nil {
			logger.Errorf("Data is loaded but failed to update status. %v", err)
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	} else if loadJob.Status.Failed == 1 {
		ctx.Loader.GetStatus().Phase = alluxiov1alpha1.LoadPhaseFailed
		if _, err := UpdateLoadStatus(ctx); err != nil {
			logger.Errorf("Failed to update status. %v", err)
			return ctrl.Result{Requeue: true}, err
		}
		logger.Errorf("Load data job failed. Please check the log of the pod for errors.")
		return ctrl.Result{}, nil
	} else {
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}
}

func getLoadJob(ctx *LoadReconcilerReqCtx) (*batchv1.Job, error) {
	loadJob := &batchv1.Job{}
	loadJobNamespacedName := types.NamespacedName{
		Name:      utils.GetLoadJobName(ctx.Name),
		Namespace: ctx.Namespace,
	}
	if err := ctx.Get(ctx.Context, loadJobNamespacedName, loadJob); err != nil {
		logger.Errorf("Error getting load job %s: %v", ctx.NamespacedName.String(), err)
		return nil, err
	}
	return loadJob, nil
}

func UpdateLoadStatus(ctx *LoadReconcilerReqCtx) (result ctrl.Result, err error) {
	if err = ctx.Update(ctx.Context, ctx.Loader); err != nil {
		logger.Errorf("Failed updating load job status: %v", err)
		return
	}
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alluxiov1alpha1.Load{}).
		Complete(r)
}
