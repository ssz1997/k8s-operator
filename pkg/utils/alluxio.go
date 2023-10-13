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

package utils

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/alluxio/k8s-operator/pkg/logger"
)

type ComponentStatusReqCtx struct {
	client.Client
	context.Context
	ClusterNameOverride string
	types.NamespacedName
}

func GetMasterStatus(ctx ComponentStatusReqCtx) (*v1.StatefulSet, error) {
	master := &v1.StatefulSet{}
	if err := ctx.Get(ctx.Context, GetMasterStatefulSetNamespacedName(ctx.ClusterNameOverride, ctx.NamespacedName), master); err != nil {
		logger.Errorf("Error getting Alluxio master StatefulSet from k8s api server: %v", err)
		return nil, err
	}
	return master, nil
}

func GetWorkerStatus(ctx ComponentStatusReqCtx) (*v1.Deployment, error) {
	worker := &v1.Deployment{}
	if err := ctx.Get(ctx.Context, GetWorkerDeploymentNamespacedName(ctx.ClusterNameOverride, ctx.NamespacedName), worker); err != nil {
		logger.Errorf("Error getting Alluxio worker Deployment from k8s api server: %v", err)
		return nil, err
	}
	return worker, nil
}

func GetFuseStatus(ctx ComponentStatusReqCtx) (*v1.DaemonSet, error) {
	fuse := &v1.DaemonSet{}
	if err := ctx.Get(ctx.Context, GetFuseDaemonSetNamespacedName(ctx.ClusterNameOverride, ctx.NamespacedName), fuse); err != nil {
		logger.Errorf("Error getting Alluxio fuse DaemonSet from k8s api server: %v", err)
		return nil, err
	}
	return fuse, nil
}

func GetProxyStatus(ctx ComponentStatusReqCtx) (*v1.DaemonSet, error) {
	proxy := &v1.DaemonSet{}
	if err := ctx.Get(ctx.Context, GetProxyDaemonSetNamespacedName(ctx.ClusterNameOverride, ctx.NamespacedName), proxy); err != nil {
		logger.Errorf("Error getting Alluxio proxy DaemonSet from k8s api server: %v", err)
		return nil, err
	}
	return proxy, nil
}
