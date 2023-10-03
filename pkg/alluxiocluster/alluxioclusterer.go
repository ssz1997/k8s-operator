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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	alluxiov1alpha1 "github.com/alluxio/k8s-operator/api/v1alpha1"
)

type AlluxioClusterer interface {
	metav1.Object
	runtime.Object

	GetDatasetName() string
	GetFuseSpec() alluxiov1alpha1.FuseSpec
	GetImage() string
	GetImagePullSecrets() []string
	GetImageTag() string
	GetNameOverride() string
	GetPagestoreSpec() alluxiov1alpha1.PagestoreSpec
	GetProxySpec() alluxiov1alpha1.ProxySpec
	GetServiceAccountName() string
	GetSpecJson() ([]byte, error)
	GetStatus() *alluxiov1alpha1.AlluxioClusterStatus
	IsDeleted() bool
}
