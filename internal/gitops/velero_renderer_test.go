package gitops

import (
	"strings"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestVeleroRendererOpenStackUsesProviderSafeValues(t *testing.T) {
	cfg := mustNewGitOpsTestConfig("velero-render", "openstack")
	cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Region = "DFW3"
	cfg.OpenCenter.Services["velero"] = &services.VeleroConfig{
		BackupBucket: "custom-backups",
		Region:       "DFW3",
		StorageType:  "swift",
	}

	spec, ok := newBuiltInRenderCatalog().Lookup("velero")
	require.True(t, ok)
	require.NotNil(t, spec.OverrideValuesRenderer)

	rendered, err := spec.OverrideValuesRenderer(cfg)
	require.NoError(t, err)

	var values struct {
		Configuration struct {
			BackupStorageLocation []struct {
				Name   string `yaml:"name"`
				Bucket string `yaml:"bucket"`
				Config struct {
					Region string `yaml:"region"`
				} `yaml:"config"`
			} `yaml:"backupStorageLocation"`
		} `yaml:"configuration"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(rendered), &values))
	require.Len(t, values.Configuration.BackupStorageLocation, 1)
	require.Equal(t, "default", values.Configuration.BackupStorageLocation[0].Name)
	require.Equal(t, "custom-backups", values.Configuration.BackupStorageLocation[0].Bucket)
	require.Equal(t, "DFW3", values.Configuration.BackupStorageLocation[0].Config.Region)

	require.Contains(t, rendered, "- name: default")
	require.Contains(t, rendered, "bucket: custom-backups")
	require.Contains(t, rendered, "region: DFW3")
	require.NotContains(t, strings.ToLower(rendered), "cloud-credentials")
	require.NotContains(t, rendered, "csi.vsphere.vmware.com/velero-vsphere-snapshot-class")
	require.NotContains(t, rendered, "driver: csi.vsphere.vmware.com")
}

func TestVeleroRendererS3UsesEndpointPathStyleAndExistingSecret(t *testing.T) {
	cfg := mustNewGitOpsTestConfig("velero-s3-render", "openstack")
	cfg.OpenCenter.Services["velero"] = &services.VeleroConfig{
		BackupBucket: "velero-bucket", Region: "RegionOne", StorageType: "s3",
		S3Endpoint: "https://s3.example", S3ForcePathStyle: true,
	}
	spec, ok := newBuiltInRenderCatalog().Lookup("velero")
	require.True(t, ok)
	rendered, err := spec.OverrideValuesRenderer(cfg)
	require.NoError(t, err)
	require.Contains(t, rendered, "s3Url: https://s3.example")
	require.Contains(t, rendered, "s3ForcePathStyle: true")
	require.Contains(t, rendered, "existingSecret: velero-cloud-credentials")
}

func TestVeleroRendererNoneOmitsBackupStorageLocation(t *testing.T) {
	cfg := mustNewGitOpsTestConfig("velero-none-render", "kind")
	cfg.OpenCenter.Services["velero"] = &services.VeleroConfig{
		BaseConfig:   services.BaseConfig{Enabled: false},
		StorageType:  "none",
		BackupBucket: "must-not-render",
		S3Endpoint:   "https://must-not-render.example",
	}

	spec, ok := newBuiltInRenderCatalog().Lookup("velero")
	require.True(t, ok)
	rendered, err := spec.OverrideValuesRenderer(cfg)
	require.NoError(t, err)
	require.NotContains(t, rendered, "backupStorageLocation")
	require.NotContains(t, rendered, "must-not-render")
	require.NotContains(t, rendered, "credentials:")
}

func TestLokiRendererNoneUsesFilesystemSingleBinaryPVC(t *testing.T) {
	cfg := mustNewGitOpsTestConfig("loki-none-render", "kind")
	cfg.OpenCenter.Services["loki"] = &services.LokiConfig{
		BaseConfig:   services.BaseConfig{Enabled: true},
		StorageType:  "none",
		BucketName:   "must-not-render",
		S3Endpoint:   "https://must-not-render.example",
		VolumeSize:   12,
		StorageClass: "local-path",
	}

	spec, ok := newBuiltInRenderCatalog().Lookup("loki")
	require.True(t, ok)
	rendered, err := spec.OverrideValuesRenderer(cfg)
	require.NoError(t, err)
	require.Contains(t, rendered, "deploymentMode: SingleBinary")
	require.Contains(t, rendered, "type: filesystem")
	require.Contains(t, rendered, "replicas: 1")
	require.Contains(t, rendered, "enabled: true")
	require.Contains(t, rendered, "size: 12Gi")
	require.Contains(t, rendered, "storageClass: local-path")
	require.Contains(t, rendered, "object_store: filesystem")
	require.NotContains(t, rendered, "bucketNames:")
	require.NotContains(t, rendered, "must-not-render")
	require.NotContains(t, rendered, "accessKeyId:")
	require.NotContains(t, rendered, "secretAccessKey:")
}
