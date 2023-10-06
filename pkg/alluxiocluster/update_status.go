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
	"time"

	"github.com/jinzhu/copier"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	alluxiov1alpha1 "github.com/alluxio/k8s-operator/api/v1alpha1"
	"github.com/alluxio/k8s-operator/pkg/dataset"
	"github.com/alluxio/k8s-operator/pkg/logger"
	"github.com/alluxio/k8s-operator/pkg/utils"
)

func UpdateStatus(alluxioClusterCtx *AlluxioClusterReconcileReqCtx) (result ctrl.Result, err error) {
	alluxioClusterStatus := alluxioClusterCtx.AlluxioClusterer.GetStatus()
	datasetStatus := alluxioClusterCtx.Datasetter.GetStatus()
	alluxioClusterOldPhase := alluxioClusterStatus.Phase
	datasetOldPhase := datasetStatus.Phase

	if datasetOldPhase == alluxiov1alpha1.DatasetPhaseNotExist {
		alluxioClusterStatus.Phase = alluxiov1alpha1.ClusterPhasePending
	} else if alluxioClusterOldPhase == alluxiov1alpha1.ClusterPhaseNone {
		alluxioClusterStatus.Phase = alluxiov1alpha1.ClusterPhaseCreatingOrUpdating
		datasetStatus.Phase = alluxiov1alpha1.DatasetPhaseBounding
	} else {
		if ready, _ := IsClusterReady(alluxioClusterCtx); ready {
			alluxioClusterStatus.Phase = alluxiov1alpha1.ClusterPhaseReady
			datasetStatus.Phase = alluxiov1alpha1.DatasetPhaseReady
			alluxioClusterName := alluxioClusterCtx.AlluxioClusterer.GetName()
			datasetStatus.BoundedAlluxioCluster = &alluxioClusterName
		} else {
			alluxioClusterStatus.Phase = alluxiov1alpha1.ClusterPhaseCreatingOrUpdating
			datasetStatus.Phase = alluxiov1alpha1.DatasetPhaseBounding
		}
	}

	if alluxioClusterStatus.Phase != alluxioClusterOldPhase {
		if err = alluxioClusterCtx.Client.Status().Update(alluxioClusterCtx.Context, alluxioClusterCtx.AlluxioClusterer); err != nil {
			logger.Errorf("Error updating cluster status: %v", err)
			return
		}
	}
	if datasetStatus.Phase != datasetOldPhase {
		if err = updateDatasetStatus(alluxioClusterCtx); err != nil {
			return
		}
	}

	if alluxioClusterStatus.Phase != alluxiov1alpha1.ClusterPhaseReady {
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}
	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

func IsClusterReady(ctx *AlluxioClusterReconcileReqCtx) (bool, error) {
	componentStatusReqCtx := utils.ComponentStatusReqCtx{}
	if err := copier.Copy(&componentStatusReqCtx, &ctx); err != nil {
		logger.Errorf("Invalid copy from AlluxioClusterReconcileReqCtx to ComponentStatusReqCtx. %v", err)
		return false, err
	}
	master, err := utils.GetMasterStatus(componentStatusReqCtx)
	if err != nil || master.Status.AvailableReplicas != master.Status.Replicas {
		return false, err
	}
	worker, err := utils.GetWorkerStatus(componentStatusReqCtx)
	if err != nil || worker.Status.AvailableReplicas != worker.Status.Replicas {
		return false, err
	}
	if ctx.AlluxioClusterer.FuseSpec().Enabled != nil && *ctx.AlluxioClusterer.FuseSpec().Enabled {
		fuse, err := utils.GetFuseStatus(componentStatusReqCtx)
		if err != nil || fuse.Status.NumberAvailable != fuse.Status.DesiredNumberScheduled {
			return false, err
		}
	}
	if ctx.AlluxioClusterer.ProxySpec().Enabled != nil && *ctx.AlluxioClusterer.ProxySpec().Enabled {
		proxy, err := utils.GetProxyStatus(componentStatusReqCtx)
		if err != nil || proxy.Status.NumberAvailable != proxy.Status.DesiredNumberScheduled {
			return false, err
		}
	}
	return true, nil
}

func updateDatasetStatus(alluxioClusterCtx *AlluxioClusterReconcileReqCtx) error {
	datasetCtx := &dataset.DatasetReconcilerReqCtx{
		Datasetter: alluxioClusterCtx.Datasetter,
		Client:     alluxioClusterCtx.Client,
		Context:    alluxioClusterCtx.Context,
		NamespacedName: types.NamespacedName{
			Name:      alluxioClusterCtx.Datasetter.GetName(),
			Namespace: alluxioClusterCtx.Namespace,
		},
	}
	if _, err := dataset.UpdateDatasetStatus(datasetCtx); err != nil {
		return err
	}
	return nil
}
