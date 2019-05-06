package component

import (
	"strings"
	"testing"

	"github.com/devspace-cloud/devspace/pkg/devspace/config/configutil"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/generated"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/devspace/helm"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/devspace-cloud/devspace/pkg/util/ptr"

	"k8s.io/client-go/kubernetes/fake"
)

func TestComponentDeployment(t *testing.T) {
	deployConfig := &latest.DeploymentConfig{
		Name: ptr.String("test-deployment"),
		Component: &latest.ComponentConfig{
			Containers: &[]*latest.ContainerConfig{
				{
					Image: ptr.String("nginx"),
				},
			},
			Service: &latest.ServiceConfig{
				Ports: &[]*latest.ServicePortConfig{
					{
						Port: ptr.Int(3000),
					},
				},
			},
		},
	}

	// Create fake devspace config
	testConfig := &latest.Config{
		Deployments: &[]*latest.DeploymentConfig{
			deployConfig,
		},
		// The images config will tell the deployment method to override the image name used in the component above with the tag defined in the generated config below
		Images: &map[string]*latest.ImageConfig{
			"default": &latest.ImageConfig{
				Image: ptr.String("nginx"),
			},
		},
	}
	configutil.SetFakeConfig(testConfig)

	// Create fake generated config
	generatedConfig := &generated.Config{
		ActiveConfig: "default",
		Configs: map[string]*generated.DevSpaceConfig{
			"default": &generated.DevSpaceConfig{
				Deploy: generated.CacheConfig{
					ImageTags: map[string]string{
						"default": "1.15", // This will be appended to nginx during deploy
					},
				},
			},
		},
	}
	generated.InitDevSpaceConfig(generatedConfig, "default")

	// Create the fake client.
	kubeClient := fake.NewSimpleClientset()
	helmClient := helm.NewFakeClient(kubeClient, configutil.TestNamespace)

	// Init handler
	deployHandler, err := New(kubeClient, deployConfig, log.GetInstance())

	// Use fake helm client
	deployHandler.HelmConfig.Helm = helmClient

	// Deploy
	err = deployHandler.Deploy(&generatedConfig.GetActive().Deploy, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Status
	status, err := deployHandler.Status()
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(status.Status, "Deployed") == false {
		t.Fatalf("Unexpected deployment status: %s != Deployed", status.Status)
	}

	err = deployHandler.Delete()
	if err != nil {
		t.Fatal(err)
	}
}
