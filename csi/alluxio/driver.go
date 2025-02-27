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

package alluxio

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"k8s.io/client-go/kubernetes"
)

const (
	driverName = "alluxio"
	version    = "1.1.0-SNAPSHOT"
)

type driver struct {
	csiDriver *csicommon.CSIDriver
	endpoint  string
	client    kubernetes.Clientset
	nodeId    string
}

func NewDriver(nodeID, endpoint string, client kubernetes.Clientset) *driver {
	glog.Infof("Driver: %v version: %v", driverName, version)
	csiDriver := csicommon.NewCSIDriver(driverName, version, nodeID)
	csiDriver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	csiDriver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	return &driver{
		nodeId:    nodeID,
		endpoint:  endpoint,
		csiDriver: csiDriver,
		client:    client,
	}
}

func (d *driver) newControllerServer() *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.csiDriver),
	}
}
func (d *driver) newNodeServer() *nodeServer {
	return &nodeServer{
		nodeId:            d.nodeId,
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d.csiDriver),
		client:            d.client,
	}
}

func (d *driver) Run() {
	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(
		d.endpoint,
		csicommon.NewDefaultIdentityServer(d.csiDriver),
		d.newControllerServer(),
		d.newNodeServer(),
	)
	s.Wait()
}
