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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlluxioClusterSpec defines the desired state of AlluxioCluster
type AlluxioClusterSpec struct {
	DatasetName *string `json:"datasetName" yaml:"datasetName"`

	NameOverride       *string             `json:"nameOverride,omitempty" yaml:"nameOverride,omitempty"`
	Image              *string             `json:"image,omitempty" yaml:"image,omitempty"`
	ImageTag           *string             `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`
	ImagePullPolicy    *string             `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	ImagePullSecrets   []string            `json:"imagePullSecrets,omitempty" yaml:"imagePullSecrets,omitempty"`
	User               *int                `json:"user,omitempty" yaml:"user,omitempty"`
	Group              *int                `json:"group,omitempty" yaml:"group,omitempty"`
	FsGroup            *string             `json:"fsGroup,omitempty" yaml:"fsGroup,omitempty"`
	HostNetwork        *bool               `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"`
	DnsPolicy          *string             `json:"dnsPolicy,omitempty" yaml:"dnsPolicy,omitempty"`
	ServiceAccountName *string             `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	HostAliases        []*HostAlias        `json:"hostAliases,omitempty" yaml:"hostAliases,omitempty"`
	NodeSelector       map[string]string   `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	Tolerations        []*Toleration       `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	Properties         map[string]string   `json:"properties,omitempty" yaml:"properties,omitempty"`
	JvmOptions         []string            `json:"jvmOptions,omitempty" yaml:"jvmOptions,omitempty"`
	PvcMounts          *MountSpec          `json:"pvcMounts,omitempty" yaml:"pvcMounts,omitempty"`
	ConfigMaps         *MountSpec          `json:"configMaps,omitempty" yaml:"configMaps,omitempty"`
	Secrets            *MountSpec          `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Master             *MasterSpec         `json:"master,omitempty" yaml:"master,omitempty"`
	Journal            *JournalSpec        `json:"journal,omitempty" yaml:"journal,omitempty"`
	Worker             *WorkerSpec         `json:"worker,omitempty" yaml:"worker,omitempty"`
	Pagestore          *PagestoreSpec      `json:"pagestore,omitempty" yaml:"pagestore,omitempty"`
	Metastore          *MetastoreSpec      `json:"metastore,omitempty" yaml:"metastore,omitempty"`
	Proxy              *ProxySpec          `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	Fuse               *FuseSpec           `json:"fuse,omitempty" yaml:"fuse,omitempty"`
	Metrics            *MetricsSpec        `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	AlluxioMonitor     *AlluxioMonitorSpec `json:"alluxio-monitor,omitempty" yaml:"alluxio-monitor,omitempty"`
	Etcd               *EtcdSpec           `json:"etcd,omitempty" yaml:"etcd,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ClusterPhase",type="string",JSONPath=`.status.phase`,priority=0
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,priority=0

// AlluxioCluster is the Schema for the alluxioclusters API
type AlluxioCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec          AlluxioClusterSpec          `json:"spec,omitempty"`
	Status        AlluxioClusterStatus        `json:"status,omitempty"`
	HelmChartSpec AlluxioClusterHelmChartSpec `json:"helmChartSpec,omitempty"`
}

// +kubebuilder:object:root=true

// AlluxioClusterList contains a list of AlluxioCluster. Operator wouldn't work without this list.
type AlluxioClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlluxioCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlluxioCluster{}, &AlluxioClusterList{})
}
