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
	"reflect"
	"time"

	"github.com/jinzhu/copier"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	alluxiov1alpha1 "github.com/alluxio/k8s-operator/api/v1alpha1"
	"github.com/alluxio/k8s-operator/pkg/dataset"
	"github.com/alluxio/k8s-operator/pkg/logger"
	"github.com/alluxio/k8s-operator/pkg/utils"
)

func UpdateStatus(alluxioClusterCtx AlluxioClusterReconcileReqCtx) (ctrl.Result, error) {
	alluxioOriginalStatusCopy := alluxioClusterCtx.AlluxioClusterer.GetStatus().DeepCopy()
	datasetOriginalStatusCopy := alluxioClusterCtx.Datasetter.GetStatus().DeepCopy()

	alluxioClusterNewPhase := alluxioOriginalStatusCopy.Phase
	datasetNewPhase := alluxioClusterCtx.Datasetter.GetStatus().Phase

	if datasetOriginalStatusCopy.Phase == alluxiov1alpha1.DatasetPhaseNotExist {
		alluxioClusterNewPhase = alluxiov1alpha1.ClusterPhasePending
	} else if alluxioOriginalStatusCopy.Phase == alluxiov1alpha1.ClusterPhaseNone {
		alluxioClusterNewPhase = alluxiov1alpha1.ClusterPhaseCreatingOrUpdating
		datasetNewPhase = alluxiov1alpha1.DatasetPhaseBounding
	} else {
		if ClusterReady(alluxioClusterCtx) {
			alluxioClusterNewPhase = alluxiov1alpha1.ClusterPhaseReady
			datasetNewPhase = alluxiov1alpha1.DatasetPhaseReady
			alluxioClusterCtx.Datasetter.GetStatus().BoundedAlluxioCluster = alluxioClusterCtx.AlluxioClusterer.GetName()
		} else {
			alluxioClusterNewPhase = alluxiov1alpha1.ClusterPhaseCreatingOrUpdating
			datasetNewPhase = alluxiov1alpha1.DatasetPhaseBounding
		}
	}
	alluxioClusterCtx.AlluxioClusterer.GetStatus().Phase = alluxioClusterNewPhase
	alluxioClusterCtx.Datasetter.GetStatus().Phase = datasetNewPhase

	if !reflect.DeepEqual(alluxioOriginalStatusCopy, alluxioClusterCtx.AlluxioClusterer.GetStatus()) {
		if err := alluxioClusterCtx.Client.Status().Update(alluxioClusterCtx.Context, alluxioClusterCtx.AlluxioClusterer); err != nil {
			logger.Errorf("Error updating cluster status: %v", err)
			return ctrl.Result{}, err
		}
	}
	if !reflect.DeepEqual(*datasetOriginalStatusCopy, alluxioClusterCtx.Datasetter.GetStatus()) {
		if err := updateDatasetStatus(alluxioClusterCtx); err != nil {
			return ctrl.Result{}, err
		}
	}

	if alluxioClusterNewPhase != alluxiov1alpha1.ClusterPhaseReady {
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}
	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

func ClusterReady(ctx AlluxioClusterReconcileReqCtx) bool {
	componentStatusReqCtx := utils.ComponentStatusReqCtx{}
	if err := copier.Copy(&componentStatusReqCtx, &ctx); err != nil {
		logger.Errorf("Invalid copy from AlluxioClusterReconcileReqCtx to ComponentStatusReqCtx. %v", err)
		return false
	}
	master, err := utils.GetMasterStatus(componentStatusReqCtx)
	if err != nil || master.Status.AvailableReplicas != master.Status.Replicas {
		return false
	}
	worker, err := utils.GetWorkerStatus(componentStatusReqCtx)
	if err != nil || worker.Status.AvailableReplicas != worker.Status.Replicas {
		return false
	}
	if ctx.AlluxioClusterer.GetFuseSpec().Enabled != nil && *ctx.AlluxioClusterer.GetFuseSpec().Enabled {
		fuse, err := utils.GetFuseStatus(componentStatusReqCtx)
		if err != nil || fuse.Status.NumberAvailable != fuse.Status.DesiredNumberScheduled {
			return false
		}
	}
	if ctx.AlluxioClusterer.GetProxySpec().Enabled != nil && *ctx.AlluxioClusterer.GetProxySpec().Enabled {
		proxy, err := utils.GetProxyStatus(componentStatusReqCtx)
		if err != nil || proxy.Status.NumberAvailable != proxy.Status.DesiredNumberScheduled {
			return false
		}
	}
	return true
}

func updateDatasetStatus(alluxioClusterCtx AlluxioClusterReconcileReqCtx) error {
	datasetCtx := dataset.DatasetReconcilerReqCtx{
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
