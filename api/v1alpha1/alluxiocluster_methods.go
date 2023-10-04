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

package v1alpha1

import "sigs.k8s.io/yaml"

func (a *AlluxioCluster) DatasetName() *string {
	return a.Spec.Dataset
}

func (a *AlluxioCluster) FuseSpec() *FuseSpec {
	return a.Spec.Fuse
}

func (a *AlluxioCluster) Image() *string {
	return a.Spec.Image
}

func (a *AlluxioCluster) ImagePullSecrets() []string {
	return a.Spec.ImagePullSecrets
}

func (a *AlluxioCluster) ImageTag() *string {
	return a.Spec.ImageTag
}

func (a *AlluxioCluster) NameOverride() *string {
	return a.Spec.NameOverride
}

func (a *AlluxioCluster) PagestoreSpec() *PagestoreSpec {
	return a.Spec.Pagestore
}

func (a *AlluxioCluster) ProxySpec() *ProxySpec {
	return a.Spec.Proxy
}

func (a *AlluxioCluster) ServiceAccountName() *string {
	return a.Spec.ServiceAccountName
}

func (a *AlluxioCluster) SpecJson() ([]byte, error) {
	return yaml.Marshal(a.Spec)
}

func (a *AlluxioCluster) GetStatus() *AlluxioClusterStatus {
	return a.Status
}

func (a *AlluxioCluster) IsDeleted() bool {
	return a.DeletionTimestamp != nil
}
