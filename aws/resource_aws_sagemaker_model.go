package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSagemakerModel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerModelCreate,
		Read:   resourceAwsSagemakerModelRead,
		Update: resourceAwsSagemakerModelUpdate,
		Delete: resourceAwsSagemakerModelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_hostname": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateSagemakerName,
						},
						"environment": {
							Type:         schema.TypeMap,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateSagemakerEnvironment,
							Elem:         &schema.Schema{Type: schema.TypeString},
						},
						"image": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateSagemakerImage,
						},
						"image_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"repository_access_mode": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.RepositoryAccessMode_Values(), false),
									},
								},
							},
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      sagemaker.ContainerModeSingleModel,
							ValidateFunc: validation.StringInSlice(sagemaker.ContainerMode_Values(), false),
						},
						"model_data_url": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateSagemakerModelDataUrl,
						},
					},
				},
			},
			"enable_network_isolation": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"inference_execution_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.InferenceExecutionMode_Values(), false),
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateSagemakerName,
			},
			"primary_container": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_hostname": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateSagemakerName,
						},
						"environment": {
							Type:         schema.TypeMap,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateSagemakerEnvironment,
							Elem:         &schema.Schema{Type: schema.TypeString},
						},
						"image": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateSagemakerImage,
						},
						"image_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"repository_access_mode": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.RepositoryAccessMode_Values(), false),
									},
								},
							},
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      sagemaker.ContainerModeSingleModel,
							ValidateFunc: validation.StringInSlice(sagemaker.ContainerMode_Values(), false),
						},
						"model_data_url": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateSagemakerModelDataUrl,
						},
						"model_package_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
						},
					},
				},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 16,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsSagemakerModelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateModelInput{
		ModelName: aws.String(name),
	}

	if v, ok := d.GetOk("primary_container"); ok {
		createOpts.PrimaryContainer = expandContainer(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("container"); ok {
		createOpts.Containers = expandContainers(v.([]interface{}))
	}

	if v, ok := d.GetOk("execution_role_arn"); ok {
		createOpts.ExecutionRoleArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		createOpts.Tags = tags.IgnoreAws().SagemakerTags()
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		createOpts.VpcConfig = expandSageMakerVpcConfigRequest(v.([]interface{}))
	}

	if v, ok := d.GetOk("enable_network_isolation"); ok {
		createOpts.EnableNetworkIsolation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("inference_execution_config"); ok {
		createOpts.InferenceExecutionConfig = expandSagemakerModelInferenceExecutionConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] Sagemaker model create config: %#v", *createOpts)
	_, err := retryOnAwsCode("ValidationException", func() (interface{}, error) {
		return conn.CreateModel(createOpts)
	})

	if err != nil {
		return fmt.Errorf("error creating Sagemaker model: %w", err)
	}
	d.SetId(name)

	return resourceAwsSagemakerModelRead(d, meta)
}

func expandSageMakerVpcConfigRequest(l []interface{}) *sagemaker.VpcConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &sagemaker.VpcConfig{
		SecurityGroupIds: expandStringSet(m["security_group_ids"].(*schema.Set)),
		Subnets:          expandStringSet(m["subnets"].(*schema.Set)),
	}
}

func resourceAwsSagemakerModelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	request := &sagemaker.DescribeModelInput{
		ModelName: aws.String(d.Id()),
	}

	model, err := conn.DescribeModel(request)
	if err != nil {
		if isAWSErr(err, "ValidationException", "") {
			log.Printf("[INFO] unable to find the sagemaker model resource and therefore it is removed from the state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Sagemaker model %s: %w", d.Id(), err)
	}

	arn := aws.StringValue(model.ModelArn)
	d.Set("arn", arn)
	d.Set("name", model.ModelName)
	d.Set("execution_role_arn", model.ExecutionRoleArn)
	d.Set("enable_network_isolation", model.EnableNetworkIsolation)

	if err := d.Set("primary_container", flattenContainer(model.PrimaryContainer)); err != nil {
		return fmt.Errorf("error setting primary_container: %w", err)
	}

	if err := d.Set("container", flattenContainers(model.Containers)); err != nil {
		return fmt.Errorf("error setting container: %w", err)
	}

	if err := d.Set("vpc_config", flattenSageMakerVpcConfigResponse(model.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc_config: %w", err)
	}

	if err := d.Set("inference_execution_config", flattenSagemakerModelInferenceExecutionConfig(model.InferenceExecutionConfig)); err != nil {
		return fmt.Errorf("error setting inference_execution_config: %w", err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Sagemaker Model (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func flattenSageMakerVpcConfigResponse(vpcConfig *sagemaker.VpcConfig) []map[string]interface{} {
	if vpcConfig == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"security_group_ids": flattenStringSet(vpcConfig.SecurityGroupIds),
		"subnets":            flattenStringSet(vpcConfig.Subnets),
	}

	return []map[string]interface{}{m}
}

func resourceAwsSagemakerModelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Sagemaker Model (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsSagemakerModelRead(d, meta)
}

func resourceAwsSagemakerModelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	deleteOpts := &sagemaker.DeleteModelInput{
		ModelName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting Sagemaker model: %s", d.Id())

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteModel(deleteOpts)
		if err == nil {
			return nil
		}

		if isAWSErr(err, "ResourceNotFound", "") {
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteModel(deleteOpts)
	}
	if err != nil {
		return fmt.Errorf("Error deleting sagemaker model: %w", err)
	}
	return nil
}

func expandContainer(m map[string]interface{}) *sagemaker.ContainerDefinition {
	container := sagemaker.ContainerDefinition{}

	//make image optional
	if v, ok := m["image"]; ok && v.(string) != "" {
		container.Image = aws.String(v.(string))
	}

	//make model package name available
	if v, ok := m["model_package_name"]; ok && v.(string) != "" {
		container.ModelPackageName = aws.String(v.(string))
	}

	if v, ok := m["mode"]; ok && v.(string) != "" {
		container.Mode = aws.String(v.(string))
	}

	if v, ok := m["container_hostname"]; ok && v.(string) != "" {
		container.ContainerHostname = aws.String(v.(string))
	}
	if v, ok := m["model_data_url"]; ok && v.(string) != "" {
		container.ModelDataUrl = aws.String(v.(string))
	}
	if v, ok := m["environment"]; ok {
		container.Environment = expandStringMap(v.(map[string]interface{}))
	}

	if v, ok := m["image_config"]; ok {
		container.ImageConfig = expandSagemakerModelImageConfig(v.([]interface{}))
	}

	return &container
}

func expandSagemakerModelImageConfig(l []interface{}) *sagemaker.ImageConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	imageConfig := &sagemaker.ImageConfig{
		RepositoryAccessMode: aws.String(m["repository_access_mode"].(string)),
	}

	return imageConfig
}

func expandContainers(a []interface{}) []*sagemaker.ContainerDefinition {
	containers := make([]*sagemaker.ContainerDefinition, 0, len(a))

	for _, m := range a {
		containers = append(containers, expandContainer(m.(map[string]interface{})))
	}

	return containers
}

func flattenContainer(container *sagemaker.ContainerDefinition) []interface{} {
	if container == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})

	//Make image optional
	if container.Image != nil {
		cfg["image"] = aws.StringValue(container.Image)
	}

	//Make model package name available
	if container.ModelPackageName != nil {
		cfg["model_package_name"] = aws.StringValue(container.ModelPackageName)
	}

	if container.Mode != nil {
		cfg["mode"] = aws.StringValue(container.Mode)
	}

	if container.ContainerHostname != nil {
		cfg["container_hostname"] = aws.StringValue(container.ContainerHostname)
	}
	if container.ModelDataUrl != nil {
		cfg["model_data_url"] = aws.StringValue(container.ModelDataUrl)
	}
	if container.Environment != nil {
		cfg["environment"] = aws.StringValueMap(container.Environment)
	}

	if container.ImageConfig != nil {
		cfg["image_config"] = flattenSagemakerImageConfig(container.ImageConfig)
	}

	return []interface{}{cfg}
}

func flattenSagemakerImageConfig(imageConfig *sagemaker.ImageConfig) []interface{} {
	if imageConfig == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})

	cfg["repository_access_mode"] = aws.StringValue(imageConfig.RepositoryAccessMode)

	return []interface{}{cfg}
}

func flattenContainers(containers []*sagemaker.ContainerDefinition) []interface{} {
	fContainers := make([]interface{}, 0, len(containers))
	for _, container := range containers {
		fContainers = append(fContainers, flattenContainer(container)[0].(map[string]interface{}))
	}
	return fContainers
}

func expandSagemakerModelInferenceExecutionConfig(l []interface{}) *sagemaker.InferenceExecutionConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.InferenceExecutionConfig{
		Mode: aws.String(m["mode"].(string)),
	}

	return config
}

func flattenSagemakerModelInferenceExecutionConfig(config *sagemaker.InferenceExecutionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})

	cfg["mode"] = aws.StringValue(config.Mode)

	return []interface{}{cfg}
}
