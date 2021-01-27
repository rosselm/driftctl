package aws

import (
	"context"
	"testing"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
)

func TestIamPolicySupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.FakeIAM)
		err     error
	}{
		{
			test:    "no iam custom policies",
			dirName: "iam_policy_empty",
			mocks: func(client *mocks.FakeIAM) {
				client.On(
					"ListPoliciesPages",
					&iam.ListPoliciesInput{Scope: aws.String("Local")},
					mock.Anything,
				).Once().Return(nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples custom policies",
			dirName: "iam_policy_multiple",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListPoliciesPages",
					&iam.ListPoliciesInput{Scope: aws.String(iam.PolicyScopeTypeLocal)},
					mock.MatchedBy(func(callback func(res *iam.ListPoliciesOutput, lastPage bool) bool) bool {
						callback(&iam.ListPoliciesOutput{Policies: []*iam.Policy{
							{
								Arn: aws.String("arn:aws:iam::929327065333:policy/policy-0"),
							},
							{
								Arn: aws.String("arn:aws:iam::929327065333:policy/policy-1"),
							},
						}}, false)
						callback(&iam.ListPoliciesOutput{Policies: []*iam.Policy{
							{
								Arn: aws.String("arn:aws:iam::929327065333:policy/policy-2"),
							},
						}}, true)
						return true
					})).Once().Return(nil)
			},
			err: nil,
		},
		{
			test:    "cannot list iam custom policies",
			dirName: "iam_policy_empty",
			mocks: func(client *mocks.FakeIAM) {
				client.On(
					"ListPoliciesPages",
					&iam.ListPoliciesInput{Scope: aws.String("Local")},
					mock.Anything,
				).Once().Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: NewBaseListError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamPolicyResourceType, resourceaws.AwsIamPolicyResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update
		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			terraform.AddProvider(terraform.AWS, provider)
			resource.AddSupplier(NewIamPolicySupplier(provider.Runner(), iam.New(provider.session)))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := mocks.FakeIAM{}
			c.mocks(&fakeIam)

			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewIamPolicyDeserializer()
			s := &IamPolicySupplier{
				provider,
				deserializer,
				&fakeIam,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
