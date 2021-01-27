package aws

import (
	"github.com/cloudskiff/driftctl/pkg/parallel"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2EbsSnapshotSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewEC2EbsSnapshotSupplier(runner *parallel.ParallelRunner, client ec2iface.EC2API) *EC2EbsSnapshotSupplier {
	return &EC2EbsSnapshotSupplier{
		terraform.Provider(terraform.AWS),
		awsdeserializer.NewEC2EbsSnapshotDeserializer(),
		client,
		terraform.NewParallelResourceReader(runner),
	}
}

func (s EC2EbsSnapshotSupplier) Resources() ([]resource.Resource, error) {
	snapshots, err := listSnapshots(s.client)
	if err != nil {
		return nil, NewBaseListError(err, resourceaws.AwsEbsSnapshotResourceType, resourceaws.AwsEbsSnapshotResourceType)
	}
	results := make([]cty.Value, 0)
	if len(snapshots) > 0 {
		for _, snapshot := range snapshots {
			snap := *snapshot
			s.runner.Run(func() (cty.Value, error) {
				return s.readEbsSnapshot(snap)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s EC2EbsSnapshotSupplier) readEbsSnapshot(snapshot ec2.Snapshot) (cty.Value, error) {
	id := aws.StringValue(snapshot.SnapshotId)
	resSnapshot, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsEbsSnapshotResourceType,
		ID: id,
	})
	if err != nil {
		logrus.Warnf("Error reading snapshot %s[%s]: %+v", id, resourceaws.AwsEbsSnapshotResourceType, err)
		return cty.NilVal, err
	}
	return *resSnapshot, nil
}

func listSnapshots(client ec2iface.EC2API) ([]*ec2.Snapshot, error) {
	var snapshots []*ec2.Snapshot
	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: []*string{
			aws.String("self"),
		},
	}
	err := client.DescribeSnapshotsPages(input, func(res *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
		snapshots = append(snapshots, res.Snapshots...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return snapshots, err
}
