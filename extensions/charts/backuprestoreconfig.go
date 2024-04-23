package charts

const (
	ConfigurationFileKey = "broInput"
)

type Config struct {
	BackupName                 string `json:"backupName" yaml:"backupName"`
	S3BucketName               string `json:"s3BucketName" yaml:"s3BucketName"`
	S3FolderName               string `json:"s3FolderName" yaml:"s3FolderName"`
	S3Region                   string `json:"s3Region" yaml:"s3Region"`
	S3Endpoint                 string `json:"s3Endpoint" yaml:"s3Endpoint"`
	VolumeName                 string `json:"volumeName" yaml:"volumeName"`
	CredentialSecretName       string `json:"credentialSecretName" yaml:"credentialSecretName"`
	CredentialSecretNamespace  string `json:"credentialSecretNamespace" yaml:"credentialSecretNamespace"`
	EndpointCA                 string `json:"endpointCA" yaml:"endpointCA"`
	ResourceSetName            string `json:"resourceSetName" yaml:"resourceSetName"`
	EncryptionConfigSecretName string `json:"encryptionConfigSecretName" yaml:"encryptionConfigSecretName"`
	Schedule                   string `json:"schedule" yaml:"schedule"`
	TlsSkipVerify              bool
	Prune                      bool
	RetentionCount             int64
	DeleteTimoutSeconds        int
	AccessKey                  string `json:"accessKey" yaml:"accessKey"`
	SecretKey                  string `json:"secretKey" yaml:"secretKey"`
}
