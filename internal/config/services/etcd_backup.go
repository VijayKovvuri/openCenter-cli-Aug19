package services

import (
	"github.com/opencenter-cloud/opencenter-cli/internal/config/registry"
)

// EtcdBackupConfig extends BaseConfig with etcd backup configuration.
type EtcdBackupConfig struct {
	BaseConfig `yaml:",inline"`

	S3Host         string `yaml:"s3_host,omitempty" json:"s3_host,omitempty" jsonschema:"description=S3-compatible endpoint host"`
	S3Endpoint     string `yaml:"s3_endpoint,omitempty" json:"s3_endpoint,omitempty" jsonschema:"description=S3-compatible endpoint URL"`
	S3BucketName   string `yaml:"s3_bucket_name,omitempty" json:"s3_bucket_name,omitempty" jsonschema:"description=S3 bucket name"`
	S3CredentialID string `yaml:"s3_credential_id,omitempty" json:"s3_credential_id,omitempty" jsonschema:"description=OpenStack EC2 credential ID"`
	S3Region       string `yaml:"s3_region,omitempty" json:"s3_region,omitempty" jsonschema:"description=S3 region"`
	StorageType    string `yaml:"storage_type,omitempty" json:"storage_type,omitempty" jsonschema:"description=etcd backup storage backend type,enum=s3,enum=none,default=s3"`
}

func init() {
	registry.RegisterServiceConfig("etcd-backup", EtcdBackupConfig{})
}
