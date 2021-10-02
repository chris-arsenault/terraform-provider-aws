package appmesh_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func testAccVirtualService_virtualNode(t *testing.T) {
	var vs appmesh.VirtualServiceData
	resourceName := "aws_appmesh_virtual_service.test"
	meshName := sdkacctest.RandomWithPrefix("tf-acc-test")
	vnName1 := sdkacctest.RandomWithPrefix("tf-acc-test")
	vnName2 := sdkacctest.RandomWithPrefix("tf-acc-test")
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", sdkacctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appmesh.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppmeshVirtualServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualNode(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(resourceName, "name", vsName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_node.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_node.0.virtual_node_name", vnName1),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualService/%s", meshName, vsName)),
				),
			},
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualNode(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(resourceName, "name", vsName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_node.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_node.0.virtual_node_name", vnName2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vsName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualService_virtualRouter(t *testing.T) {
	var vs appmesh.VirtualServiceData
	resourceName := "aws_appmesh_virtual_service.test"
	meshName := sdkacctest.RandomWithPrefix("tf-acc-test")
	vrName1 := sdkacctest.RandomWithPrefix("tf-acc-test")
	vrName2 := sdkacctest.RandomWithPrefix("tf-acc-test")
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", sdkacctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appmesh.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppmeshVirtualServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualRouter(meshName, vrName1, vrName2, vsName, "aws_appmesh_virtual_router.foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(resourceName, "name", vsName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_router.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_router.0.virtual_router_name", vrName1),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualService/%s", meshName, vsName))),
			},
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualRouter(meshName, vrName1, vrName2, vsName, "aws_appmesh_virtual_router.bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(resourceName, "name", vsName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_router.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_router.0.virtual_router_name", vrName2),
				),
			},
		},
	})
}

func testAccVirtualService_tags(t *testing.T) {
	var vs appmesh.VirtualServiceData
	resourceName := "aws_appmesh_virtual_service.test"
	meshName := sdkacctest.RandomWithPrefix("tf-acc-test")
	vnName1 := sdkacctest.RandomWithPrefix("tf-acc-test")
	vnName2 := sdkacctest.RandomWithPrefix("tf-acc-test")
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", sdkacctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, appmesh.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppmeshVirtualServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualServiceConfig_tags(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.foo", "foo", "bar", "good", "bad"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad"),
				),
			},
			{
				Config: testAccAppmeshVirtualServiceConfig_tags(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.foo", "foo2", "bar", "good", "bad2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo2", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad2"),
				),
			},
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualNode(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vsName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAppmeshVirtualServiceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_virtual_service" {
			continue
		}

		_, err := conn.DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
			MeshName:           aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualServiceName: aws.String(rs.Primary.Attributes["name"]),
		})
		if tfawserr.ErrMessageContains(err, appmesh.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("still exist.")
	}

	return nil
}

func testAccCheckAppmeshVirtualServiceExists(name string, v *appmesh.VirtualServiceData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
			MeshName:           aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualServiceName: aws.String(rs.Primary.Attributes["name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.VirtualService

		return nil
	}
}

func testAccAppmeshVirtualServiceConfig_virtualNode(meshName, vnName1, vnName2, vsName, rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_node" "bar" {
  name      = %[3]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_service" "test" {
  name      = %[4]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = %[5]s.name
      }
    }
  }
}
`, meshName, vnName1, vnName2, vsName, rName)
}

func testAccAppmeshVirtualServiceConfig_virtualRouter(meshName, vrName1, vrName2, vsName, rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "foo" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_virtual_router" "bar" {
  name      = %[3]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_virtual_service" "test" {
  name      = %[4]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_router {
        virtual_router_name = %[5]s.name
      }
    }
  }
}
`, meshName, vrName1, vrName2, vsName, rName)
}

func testAccAppmeshVirtualServiceConfig_tags(meshName, vnName1, vnName2, vsName, rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_node" "bar" {
  name      = %[3]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_service" "test" {
  name      = %[4]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = %[5]s.name
      }
    }
  }

  tags = {
    %[6]s = %[7]q
    %[8]s = %[9]q
  }
}
`, meshName, vnName1, vnName2, vsName, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
