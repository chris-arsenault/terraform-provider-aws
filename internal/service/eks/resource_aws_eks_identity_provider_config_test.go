package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfeks "github.com/hashicorp/terraform-provider-aws/aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/eks/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	resource.AddTestSweepers("aws_eks_identity_provider_config", &resource.Sweeper{
		Name: "aws_eks_identity_provider_config",
		F:    testSweepEksIdentityProviderConfigs,
	})
}

func testSweepEksIdentityProviderConfigs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	ctx := context.TODO()
	conn := client.(*conns.AWSClient).EKSConn
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListClustersPagesWithContext(ctx, input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListIdentityProviderConfigsInput{
				ClusterName: cluster,
			}

			err := conn.ListIdentityProviderConfigsPagesWithContext(ctx, input, func(page *eks.ListIdentityProviderConfigsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, identityProviderConfig := range page.IdentityProviderConfigs {
					r := ResourceIdentityProviderConfig()
					d := r.Data(nil)
					d.SetId(tfeks.IdentityProviderConfigCreateResourceID(aws.StringValue(cluster), aws.StringValue(identityProviderConfig.Name)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Identity Provider Configs (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping EKS Identity Provider Configs sweep for %s: %w", region, err))
		return sweeperErrs // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Identity Provider Configs (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSEksIdentityProviderConfig_basic(t *testing.T) {
	var config eks.OidcIdentityProviderConfig
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	eksClusterResourceName := "aws_eks_cluster.test"
	resourceName := "aws_eks_identity_provider_config.test"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAWSEks(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSEksIdentityProviderConfigConfigIssuerUrl(rName, "http://example.com"),
				ExpectError: regexp.MustCompile(`expected .* to have a url with schema of: "https", got http://example.com`),
			},
			{
				Config: testAccAWSEksIdentityProviderConfigConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(ctx, resourceName, &config),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("identityproviderconfig/%[1]s/oidc/%[1]s/.+", rName))),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", eksClusterResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "oidc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.client_id", "example.net"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.groups_claim", ""),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.groups_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.identity_provider_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.issuer_url", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.required_claims.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.username_claim", ""),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.username_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEksIdentityProviderConfig_disappears(t *testing.T) {
	var config eks.OidcIdentityProviderConfig
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_identity_provider_config.test"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAWSEks(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksIdentityProviderConfigConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(ctx, resourceName, &config),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceIdentityProviderConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEksIdentityProviderConfig_AllOidcOptions(t *testing.T) {
	var config eks.OidcIdentityProviderConfig
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_identity_provider_config.test"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAWSEks(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksIdentityProviderConfigAllOidcOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "oidc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.client_id", "example.net"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.groups_claim", "groups"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.groups_prefix", "oidc:"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.identity_provider_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.issuer_url", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.required_claims.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.required_claims.keyOne", "valueOne"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.required_claims.keyTwo", "valueTwo"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.username_claim", "email"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.username_prefix", "-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEksIdentityProviderConfig_Tags(t *testing.T) {
	var config eks.OidcIdentityProviderConfig
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_identity_provider_config.test"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAWSEks(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksIdentityProviderConfigConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksIdentityProviderConfigConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEksIdentityProviderConfigConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSEksIdentityProviderConfigExists(ctx context.Context, resourceName string, config *eks.OidcIdentityProviderConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EKS Identity Profile Config ID is set")
		}

		clusterName, configName, err := tfeks.IdentityProviderConfigParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn

		output, err := finder.FindOIDCIdentityProviderConfigByClusterNameAndConfigName(ctx, conn, clusterName, configName)

		if err != nil {
			return err
		}

		*config = *output

		return nil
	}
}

func testAccCheckAWSEksIdentityProviderConfigDestroy(s *terraform.State) error {
	ctx := context.TODO()
	conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_identity_provider_config" {
			continue
		}

		clusterName, configName, err := tfeks.IdentityProviderConfigParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.FindOIDCIdentityProviderConfigByClusterNameAndConfigName(ctx, conn, clusterName, configName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EKS Identity Profile Config %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSEksIdentityProviderConfigConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "eks.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccAWSEksIdentityProviderConfigConfigName(rName string) string {
	return acctest.ConfigCompose(testAccAWSEksIdentityProviderConfigConfigBase(rName), fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name

  oidc {
    client_id                     = "example.net"
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://example.com"
  }
}
`, rName))
}

func testAccAWSEksIdentityProviderConfigConfigIssuerUrl(rName, issuerUrl string) string {
	return acctest.ConfigCompose(testAccAWSEksIdentityProviderConfigConfigBase(rName), fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name

  oidc {
    client_id                     = "example.net"
    identity_provider_config_name = %[1]q
    issuer_url                    = %[2]q
  }
}
`, rName, issuerUrl))
}

func testAccAWSEksIdentityProviderConfigAllOidcOptions(rName string) string {
	return acctest.ConfigCompose(testAccAWSEksIdentityProviderConfigConfigBase(rName), fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name

  oidc {
    client_id                     = "example.net"
    groups_claim                  = "groups"
    groups_prefix                 = "oidc:"
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://example.com"
    username_claim                = "email"
    username_prefix               = "-"

    required_claims = {
      keyOne = "valueOne"
      keyTwo = "valueTwo"
    }
  }
}
`, rName))
}

func testAccAWSEksIdentityProviderConfigConfigTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAWSEksIdentityProviderConfigConfigBase(rName), fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name

  oidc {
    client_id                     = "example.net"
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://example.com"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSEksIdentityProviderConfigConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAWSEksIdentityProviderConfigConfigBase(rName), fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name

  oidc {
    client_id                     = "example.net"
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://example.com"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
