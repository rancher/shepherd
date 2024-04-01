package charts

const (
	BackupRestoreConfigurationFileKey = "backupRestoreInput"
)

type BackupRestoreConfig struct {
	BackupName                string `json:"backupName" yaml:"backupName"`
	S3BucketName              string `json:"s3BucketName" yaml:"s3BucketName"`
	S3FolderName              string `json:"s3FolderName" yaml:"s3FolderName"`
	S3Region                  string `json:"s3Region" yaml:"s3Region"`
	S3Endpoint                string `json:"s3Endpoint" yaml:"s3Endpoint"`
	VolumeName                string `json:"volumeName" yaml:"volumeName"`
	CredentialSecretNamespace string `json:"credentialSecretNamespace" yaml:"credentialSecretNamespace"`
	ResourceSetName           string `json:"resourceSetName" yaml:"resourceSetName"`
	Prune                     bool
	AccessKey                 string `json:"accessKey" yaml:"accessKey"`
	SecretKey                 string `json:"secretKey" yaml:"secretKey"`
	Rke1ClusterName           string `json:"rke1ClusterName" yaml:"rke1ClusterName"`
	Rke2ClusterName           string `json:"rke2ClusterName" yaml:"rke2ClusterName"`
	ClusterNamespace          string `json:"clusterNamepace" yaml:"clusterNamespace"`
}
