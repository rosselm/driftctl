package error

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/remote/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/sirupsen/logrus"
)

func HandleListAwsError(err error, alertr *alerter.Alerter) (handled bool) {
	listError, ok := err.(aws.ListError)
	if !ok {
		return false
	}

	reqerr, ok := listError.RootCause().(awserr.RequestFailure)
	if !ok {
		return false
	}

	if reqerr.StatusCode() == 403 {
		message := fmt.Sprintf("Ignoring %s from drift calculation: Listing %s is forbidden.", listError.SupplierType(), listError.ListedTypeError())
		logrus.Debugf(message)
		alertr.SendAlert(listError.SupplierType(), alerter.Alert{
			Message:              message,
			ShouldIgnoreResource: true,
		})
		return true
	}

	return false
}
