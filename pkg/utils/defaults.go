/*
Copyright 2019 IBM Corporation
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"io/ioutil"

	kamv1 "github.com/kappnav/operator/pkg/apis/actions/v1"
	kappnavv1 "github.com/kappnav/operator/pkg/apis/kappnav/v1"
	"sigs.k8s.io/yaml"
)

// SetKappnavDefaults sets default values on the CR instance
func SetKappnavDefaults(instance *kappnavv1.Kappnav) error {
	defaults, err := getDefaults()
	if err != nil {
		return err
	}
	setAPIContainerDefaults(instance, defaults)
	setUIContainerDefaults(instance, defaults)
	setControllerContainerDefaults(instance, defaults)
	setExtensionContainerDefaults(instance, defaults)
	setImageDefaults(instance, defaults)
	setEnvironmentDefaults(instance, defaults)
	setLoggingDefaults(instance, defaults)
	return nil
}

func getDefaults() (*kappnavv1.Kappnav, error) {
	// Read default values file
	fData, err := ioutil.ReadFile("deploy/default_values.yaml")
	if err != nil {
		return nil, err
	}
	defaults := &kappnavv1.Kappnav{}
	err = yaml.Unmarshal(fData, defaults)
	if err != nil {
		return nil, err
	}
	return defaults, nil
}

func setAPIContainerDefaults(instance *kappnavv1.Kappnav, defaults *kappnavv1.Kappnav) {
	apiConfig := instance.Spec.AppNavAPI
	if apiConfig == nil {
		instance.Spec.AppNavAPI = defaults.Spec.AppNavAPI
	} else {
		setContainerDefaults(apiConfig, defaults.Spec.AppNavAPI)
	}
}

func setUIContainerDefaults(instance *kappnavv1.Kappnav, defaults *kappnavv1.Kappnav) {
	uiConfig := instance.Spec.AppNavUI
	if uiConfig == nil {
		instance.Spec.AppNavUI = defaults.Spec.AppNavUI
	} else {
		setContainerDefaults(uiConfig, defaults.Spec.AppNavUI)
	}
}

func setControllerContainerDefaults(instance *kappnavv1.Kappnav, defaults *kappnavv1.Kappnav) {
	controllerConfig := instance.Spec.AppNavController
	if controllerConfig == nil {
		instance.Spec.AppNavController = defaults.Spec.AppNavController
	} else {
		setContainerDefaults(controllerConfig, defaults.Spec.AppNavController)
	}
}

func setExtensionContainerDefaults(instance *kappnavv1.Kappnav, defaults *kappnavv1.Kappnav) {
	extensionContainerConfig := instance.Spec.ExtensionContainers
	if extensionContainerConfig == nil {
		instance.Spec.ExtensionContainers = defaults.Spec.ExtensionContainers
	} else {
		defaultExtensionContainerConfig := defaults.Spec.ExtensionContainers
		if defaultExtensionContainerConfig != nil {
			for defaultConfigName, defaultConfig := range defaultExtensionContainerConfig {
				extConfig := extensionContainerConfig[defaultConfigName]
				if extConfig != nil {
					setContainerDefaults(extConfig, defaultConfig)
				} else {
					extensionContainerConfig[defaultConfigName] = defaultConfig
				}
			}
		}
	}
}

func setContainerDefaults(containerConfig *kappnavv1.KappnavContainerConfiguration,
	defaultContainerConfig *kappnavv1.KappnavContainerConfiguration) {
	if len(containerConfig.Repository) == 0 {
		containerConfig.Repository = defaultContainerConfig.Repository
	}
	if len(containerConfig.Tag) == 0 {
		containerConfig.Tag = defaultContainerConfig.Tag
	}
	if containerConfig.Resources == nil {
		containerConfig.Resources = defaultContainerConfig.Resources
	} else {
		if containerConfig.Resources.Enabled {
			if containerConfig.Resources.Requests == nil {
				containerConfig.Resources.Requests = defaultContainerConfig.Resources.Requests
			} else {
				setResourceDefaults(containerConfig.Resources.Requests, defaultContainerConfig.Resources.Requests)
			}
			if containerConfig.Resources.Limits == nil {
				containerConfig.Resources.Limits = defaultContainerConfig.Resources.Limits
			} else {
				setResourceDefaults(containerConfig.Resources.Limits, defaultContainerConfig.Resources.Limits)
			}
		}
	}
}

func setResourceDefaults(resources *kappnavv1.Resources, defaultResources *kappnavv1.Resources) {
	if len(resources.CPU) == 0 {
		resources.CPU = defaultResources.CPU
	}
	if len(resources.Memory) == 0 {
		resources.Memory = defaultResources.Memory
	}
}

func setImageDefaults(instance *kappnavv1.Kappnav, defaults *kappnavv1.Kappnav) {
	image := instance.Spec.Image
	if image == nil {
		instance.Spec.Image = defaults.Spec.Image
	} else {
		if len(image.PullPolicy) == 0 {
			image.PullPolicy = defaults.Spec.Image.PullPolicy
		}
		if image.PullSecrets == nil {
			image.PullSecrets = defaults.Spec.Image.PullSecrets
		}
	}
}

func setEnvironmentDefaults(instance *kappnavv1.Kappnav, defaults *kappnavv1.Kappnav) {
	env := instance.Spec.Env
	if env == nil {
		instance.Spec.Env = defaults.Spec.Env
	} else {
		if len(env.KubeEnv) == 0 {
			env.KubeEnv = defaults.Spec.Env.KubeEnv
		}
	}
}

// SetKAMDefaults sets default kam values on the CR instance
func SetKAMDefaults(instance_kam *kamv1.KindActionMapping) error {
	err := getKAMDefaults(instance_kam)
	if err != nil {
		return err
	}
	return nil
}

func getKAMDefaults(instance_kam *kamv1.KindActionMapping) error {
	// Read default kam values file
	fData, err := ioutil.ReadFile("deploy/default_kam.yaml")
	if err != nil {
		return err
	}
	defaults := &kamv1.KindActionMapping{}

	err = yaml.Unmarshal(fData, defaults)
	if err != nil {
		return err
	}

	instance_kam.Spec = defaults.Spec
	instance_kam.Status = defaults.Status
	return nil
}

// set default logging values
func setLoggingDefaults(instance *kappnavv1.Kappnav, defaults *kappnavv1.Kappnav) {
	logging := instance.Spec.Logging

	if logging == nil {
		instance.Spec.Logging = defaults.Spec.Logging
	} else {
		defaultLogging := defaults.Spec.Logging
		for key, value := range logging {
			// set default values when no logging values specified in kappnav instance
			if len(key) != 0 && len(value) == 0 {
				if defaultLogging != nil {
					logging[key] = defaultLogging[key]
				}
			}
		}
	}
}
