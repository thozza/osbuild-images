package awscloud

import (
	"context"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type EC2 interface {
	DescribeRegions(context.Context, *ec2.DescribeRegionsInput, ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)

	// Images
	CopyImage(context.Context, *ec2.CopyImageInput, ...func(*ec2.Options)) (*ec2.CopyImageOutput, error)
	RegisterImage(context.Context, *ec2.RegisterImageInput, ...func(*ec2.Options)) (*ec2.RegisterImageOutput, error)
	DeregisterImage(context.Context, *ec2.DeregisterImageInput, ...func(*ec2.Options)) (*ec2.DeregisterImageOutput, error)
	DescribeImages(context.Context, *ec2.DescribeImagesInput, ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	ModifyImageAttribute(context.Context, *ec2.ModifyImageAttributeInput, ...func(*ec2.Options)) (*ec2.ModifyImageAttributeOutput, error)

	// Snapshots
	DeleteSnapshot(context.Context, *ec2.DeleteSnapshotInput, ...func(*ec2.Options)) (*ec2.DeleteSnapshotOutput, error)
	DescribeImportSnapshotTasks(context.Context, *ec2.DescribeImportSnapshotTasksInput, ...func(*ec2.Options)) (*ec2.DescribeImportSnapshotTasksOutput, error)
	ImportSnapshot(context.Context, *ec2.ImportSnapshotInput, ...func(*ec2.Options)) (*ec2.ImportSnapshotOutput, error)
	ModifySnapshotAttribute(context.Context, *ec2.ModifySnapshotAttributeInput, ...func(*ec2.Options)) (*ec2.ModifySnapshotAttributeOutput, error)

	// Tags
	CreateTags(context.Context, *ec2.CreateTagsInput, ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

type S3 interface {
	GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListBuckets(context.Context, *s3.ListBucketsInput, ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	PutObjectAcl(context.Context, *s3.PutObjectAclInput, ...func(*s3.Options)) (*s3.PutObjectAclOutput, error)
}

type S3Uploader interface {
	Upload(context.Context, *s3.PutObjectInput, ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

type S3Presign interface {
	PresignGetObject(context.Context, *s3.GetObjectInput, ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}
