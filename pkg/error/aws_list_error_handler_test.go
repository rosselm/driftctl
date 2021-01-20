package error

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/alerter"
)

func TestHandleListAwsError(t *testing.T) {

	tests := []struct {
		name        string
		err         error
		wantAlerts  alerter.Alerts
		wantHandled bool
	}{
		{
			name:        "Handled error",
			err:         aws.NewBaseListError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsVpcResourceType, resourceaws.AwsVpcResourceType),
			wantAlerts:  alerter.Alerts{"aws_vpc": []alerter.Alert{alerter.Alert{Message: "Ignoring aws_vpc from drift calculation: Listing aws_vpc is forbidden.", ShouldIgnoreResource: true}}},
			wantHandled: true,
		},
		{
			name:        "Not Handled error code",
			err:         aws.NewBaseListError(awserr.NewRequestFailure(nil, 404, ""), resourceaws.AwsVpcResourceType, resourceaws.AwsVpcResourceType),
			wantAlerts:  map[string][]alerter.Alert{},
			wantHandled: false,
		},
		{
			name:        "Not Handled error",
			err:         aws.NewBaseSupplierError(awserr.NewRequestFailure(nil, 404, ""), map[string]string{}, resourceaws.AwsVpcResourceType),
			wantAlerts:  map[string][]alerter.Alert{},
			wantHandled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertr := alerter.NewAlerter()
			gotHandled := HandleListAwsError(tt.err, alertr)
			assert.Equal(t, tt.wantHandled, gotHandled)

			retrieve := alertr.Retrieve()
			assert.Equal(t, tt.wantAlerts, retrieve)

		})
	}
}
