package rds

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// ProxyEndpoint NotFound
	proxyEndpointStatusNotFound = "NotFound"

	// ProxyEndpoint Unknown
	proxyEndpointStatusUnknown = "Unknown"
)

func statusEventSubscription(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindEventSubscriptionByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// statusDBProxyEndpoint fetches the ProxyEndpoint and its Status
func statusDBProxyEndpoint(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindDBProxyEndpoint(conn, id)

		if err != nil {
			return nil, proxyEndpointStatusUnknown, err
		}

		if output == nil {
			return nil, proxyEndpointStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusDBClusterRole(conn *rds.RDS, dbClusterID, roleARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindDBClusterRoleByDBClusterIDAndRoleARN(conn, dbClusterID, roleARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}