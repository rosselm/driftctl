package aws

import (
	"context"
	"testing"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEC2EipAssociationSupplier_Resources(t *testing.T) {
	tests := []struct {
		test      string
		dirName   string
		addresses []*ec2.Address
		listError error
		err       error
	}{
		{
			test:      "no eip associations",
			dirName:   "ec2_eip_association_empty",
			addresses: []*ec2.Address{},
			err:       nil,
		},
		{
			test:    "with eip associations",
			dirName: "ec2_eip_association_single",
			addresses: []*ec2.Address{
				{
					AssociationId: aws.String("eipassoc-0e9a7356e30f0c3d1"),
				},
			},
			err: nil,
		},
		{
			test:      "Cannot list eip associations",
			dirName:   "ec2_eip_association_empty",
			listError: awserr.NewRequestFailure(nil, 403, ""),
			err:       NewBaseListError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEipAssociationResourceType, resourceaws.AwsEipAssociationResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update
		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			terraform.AddProvider(terraform.AWS, provider)
			resource.AddSupplier(NewEC2EipAssociationSupplier(provider.Runner(), ec2.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2EipAssociationDeserializer()
			client := mocks.NewMockAWSEC2EipClient(tt.addresses)
			if tt.listError != nil {
				client = mocks.NewMockAWSEC2ErrorClient(tt.listError)
			}
			s := &EC2EipAssociationSupplier{
				provider,
				deserializer,
				client,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(t, tt.err, err)

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
