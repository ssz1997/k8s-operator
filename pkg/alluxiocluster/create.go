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
	"bytes"
	"os"

	"go.uber.org/config"

	"github.com/alluxio/k8s-operator/pkg/logger"
	"github.com/alluxio/k8s-operator/pkg/utils"
)

const (
	chartPath                      = "/opt/charts/alluxio"
	alluxioHelmChartValuesFilePath = "/opt/charts/alluxio/values.yaml"
)

func CreateAlluxioClusterIfNotExist(ctx *AlluxioClusterReconcileReqCtx) error {
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
	PropagateAndWriteConfFile(ctx)

	// helm install release with the constructed config.yaml
	helmCtx = utils.HelmContext{
		HelmChartPath:  chartPath,
		ConfigFilePath: utils.GetConfYamlPath(ctx.NamespacedName),
		Namespace:      ctx.Namespace,
		ReleaseName:    ctx.Name,
	}
	if err := utils.HelmInstall(helmCtx); err != nil {
		logger.Errorf("error installing helm release. Uninstalling...")
		if err := DeleteAlluxioClusterIfExists(ctx.NamespacedName); err != nil {
			logger.Errorf("failed to delete failed helm release %s: %v", ctx.NamespacedName.String(), err.Error())
			return err
		}
	}
	return nil
}

func PropagateAndWriteConfFile(ctx *AlluxioClusterReconcileReqCtx) error {
	defaultValues, err := os.Open(alluxioHelmChartValuesFilePath)
	if err != nil {
		logger.Errorf("Error reading alluxio helm chart default values at %v: %v", alluxioHelmChartValuesFilePath, err.Error())
		return err
	}
	clusterYaml, err := ctx.AlluxioClusterer.SpecYaml()
	if err != nil {
		logger.Errorf("Error marshalling Alluxio config to yaml file. %v", err.Error())
		return err
	}
	datasetYaml, err := ctx.Datasetter.SpecYaml()
	if err != nil {
		logger.Errorf("Error marshalling dataset config to yaml file. %v", err.Error())
		return err
	}
	mergedConfigYAML, err := config.NewYAML(config.Source(defaultValues), config.Source(bytes.NewReader(clusterYaml)), config.Source(bytes.NewReader(datasetYaml)))
	if err != nil {
		logger.Errorf("Error merging user config into default config: %v", err.Error())
		return err
	}
	mergedConfigBytes, err := ctx.AlluxioClusterer.HelmChartValues().YAMLToYaml(mergedConfigYAML)
	if err != nil {
		logger.Errorf("Error marshalling merged config into byte array. %v", err.Error())
		return err
	}

	confYamlFile, err := os.OpenFile(utils.GetConfYamlPath(ctx.NamespacedName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	defer confYamlFile.Close()
	if err != nil {
		logger.Errorf("failed to create empty config file: %v", err)
		return err
	}
	if _, err := confYamlFile.Write(mergedConfigBytes); err != nil {
		logger.Errorf("Error writing to config file: %v", err)
		return err
	}
	return nil
}

