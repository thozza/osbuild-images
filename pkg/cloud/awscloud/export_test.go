package awscloud

type AwsClient = awsClient

func NewAWSForTest(s3cli S3, upldr S3Uploader, sign S3Presign) *AWS {
	return &AWS{
		s3:         s3cli,
		s3uploader: upldr,
		s3presign:  sign,
	}
}

func MockNewAwsClient(f func(string) (awsClient, error)) (restore func()) {
	saved := newAwsClient
	newAwsClient = f
	return func() {
		newAwsClient = saved
	}
}
