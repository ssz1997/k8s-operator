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
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/alluxio/k8s-operator/pkg/logger"
	"github.com/alluxio/k8s-operator/pkg/utils"
)

const (
	chartPath                = "/opt/charts/alluxio"
	etcdHelmReleaseName      = "etcd-alluxio"
	etcdHelmReleaseNamespace = "alluxio-etcd"
	etcdHelmRepoName         = "etcd"
	etcdHelmRepoUrl          = "https://charts.bitnami.com/bitnami"
	etcdHelmRepoVersion      = "9.0.5"
)

func CreateAlluxioClusterIfNotExist(ctx AlluxioClusterReconcileReqCtx) error {
	// if the release has already been deployed, requeue without further actions
	helmCtx := utils.HelmContext{
		Namespace:   ctx.Namespace,
		ReleaseName: ctx.Name,
	}
	exists, err := utils.IfHelmReleaseExists(helmCtx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	logger.Infof("Creating Alluxio cluster %s.", ctx.NamespacedName.String())

	if err := SetUpEtcdClusterIfNeeded(ctx); err != nil {
		return err
	}

	// Construct alluxio config file
	clusterYaml, err := yaml.Marshal(ctx.AlluxioCluster.Spec)
	if err != nil {
		return err
	}
	confYamlFilePath := utils.GetConfYamlPath(ctx.NamespacedName)
	confYamlFile, err := os.OpenFile(confYamlFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	defer confYamlFile.Close()
	if err != nil {
		logger.Errorf("failed to create empty Alluxio config file: %v", err)
		return err
	}
	if _, err := confYamlFile.Write(clusterYaml); err != nil {
		logger.Errorf("Error writing to Alluxio config file: %v", err)
		return err
	}
	datasetYalm, err := yaml.Marshal(ctx.Dataset.Spec)
	if err != nil {
		return err
	}
	if _, err = confYamlFile.WriteString(string(datasetYalm)); err != nil {
		logger.Errorf("Error writing to Alluxio config file: %v", err)
		return err
	}
	// helm install release with the constructed config file
	helmCtx = utils.HelmContext{
		HelmChartPath:  chartPath,
		ConfigFilePath: confYamlFilePath,
		Namespace:      ctx.Namespace,
		ReleaseName:    ctx.Name,
	}
	if err := utils.HelmInstall(helmCtx); err != nil {
		logger.Errorf("error installing helm release. Uninstalling...")
		if err := DeleteAlluxioClusterIfExist(ctx.NamespacedName); err != nil {
			logger.Errorf("failed to delete failed helm release %s: %v", ctx.NamespacedName.String(), err)
			return err
		}
	}
	return nil
}

func SetUpEtcdClusterIfNeeded(ctx AlluxioClusterReconcileReqCtx) error {
	if ctx.AlluxioCluster.Spec.Etcd.Enabled {
		tweakEtcdPropertiesInAlluxio(ctx)
		return installEtcdClusterIfNeeded(ctx)
	}
	return nil
}

func tweakEtcdPropertiesInAlluxio(ctx AlluxioClusterReconcileReqCtx) {
	ctx.AlluxioCluster.Spec.Properties["alluxio.worker.membership.manager.type"] = "ETCD"
	ctx.AlluxioCluster.Spec.Properties["alluxio.etcd.endpoints"] = fmt.Sprintf("http://%v.%v.svc.cluster.local:2379", etcdHelmReleaseName, etcdHelmReleaseNamespace)
	// We don't start etcd cluster with Alluxio helm chart
	ctx.AlluxioCluster.Spec.Etcd.Enabled = false
}

func installEtcdClusterIfNeeded(ctx AlluxioClusterReconcileReqCtx) error {
	etcdClusterCtx := utils.HelmContext{
		Namespace:     etcdHelmReleaseNamespace,
		ReleaseName:   etcdHelmReleaseName,
		Version:       etcdHelmRepoVersion,
		HelmChartPath: fmt.Sprintf("%s/%s", etcdHelmRepoName, etcdHelmRepoName),
	}
	exist, err := utils.IfHelmReleaseExists(etcdClusterCtx)
	if err != nil {
		return err
	}
	if !exist {
		if err := createEtcdNamespaceIfNeeded(); err != nil {
			return err
		}
		etcdYamlFilePath := utils.GetConfYamlPath(types.NamespacedName{Name: etcdHelmReleaseName, Namespace: etcdHelmReleaseNamespace})
		if err := constructEtcdHelmConfYaml(ctx, etcdYamlFilePath); err != nil {
			return err
		}
		if err := utils.HelmAddRepo(etcdHelmRepoName, etcdHelmRepoUrl); err != nil {
			return err
		}
		etcdClusterCtx.ConfigFilePath = etcdYamlFilePath
		if err := utils.HelmInstall(etcdClusterCtx); err != nil {
			logger.Errorf("error installing helm release %s. Uninstalling...", etcdHelmReleaseName)
			if err := utils.HelmDeleteIfExist(etcdClusterCtx); err != nil {
				logger.Errorf("failed to delete failed helm release %s: %v", etcdHelmReleaseName, err.Error())
				return err
			}
		}
	}

	return nil
}

func constructEtcdHelmConfYaml(ctx AlluxioClusterReconcileReqCtx, etcdYamlFilePath string) error {
	etcdYaml, err := yaml.Marshal(ctx.AlluxioCluster.Spec.Etcd)
	if err != nil {
		return err
	}
	etcdYamlFile, err := os.OpenFile(etcdYamlFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	defer etcdYamlFile.Close()
	if err != nil {
		logger.Errorf("failed to create empty etcd config file: %v", err)
		return err
	}
	if _, err := etcdYamlFile.Write(etcdYaml); err != nil {
		logger.Errorf("Error writing to etcd config file: %v", err)
		return err
	}
	return nil
}

func createEtcdNamespaceIfNeeded() error {
	clientSet, err := utils.GetK8sClient()
	if err != nil {
		logger.Errorf("Failed to get clientSet while creating namespace %s: %v", etcdHelmReleaseNamespace, err)
		return err
	}
	etcdNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: etcdHelmReleaseNamespace,
		},
	}
	_, err = clientSet.CoreV1().Namespaces().Get(context.TODO(), etcdHelmReleaseNamespace, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		if _, err := clientSet.CoreV1().Namespaces().Create(context.TODO(), etcdNamespace, metav1.CreateOptions{}); err != nil {
			logger.Errorf("error creating namespace %s: %v", err.Error())
			return err
		}
	} else {
		logger.Errorf("error checking whether namespace %s exists: %v", etcdHelmReleaseNamespace, err.Error())
	}
	return nil
}
