package vpc

import (
	"context"
	"fmt"
	"log"
	"strings"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceTencentCloudEniAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudEniAttachmentCreate,
		Read:   resourceTencentCloudEniAttachmentRead,
		Delete: resourceTencentCloudEniAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"eni_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the ENI.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the instance which bind the ENI.",
			},
		},
	}
}

func resourceTencentCloudEniAttachmentCreate(d *schema.ResourceData, m interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_eni_attachment.create")()

	var (
		logId   = tccommon.GetLogId(tccommon.ContextNil)
		ctx     = context.WithValue(context.TODO(), tccommon.LogIdKey, logId)
		service = VpcService{client: m.(tccommon.ProviderMeta).GetAPIV3Conn()}
	)

	eniId := d.Get("eni_id").(string)
	cvmId := d.Get("instance_id").(string)

	if err := service.AttachEniToCvm(ctx, eniId, cvmId); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s+%s", eniId, cvmId))

	return resourceTencentCloudEniAttachmentRead(d, m)
}

func resourceTencentCloudEniAttachmentRead(d *schema.ResourceData, m interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_eni_attachment.read")()
	defer tccommon.InconsistentCheck(d, m)()

	var (
		logId   = tccommon.GetLogId(tccommon.ContextNil)
		ctx     = context.WithValue(context.TODO(), tccommon.LogIdKey, logId)
		service = VpcService{client: m.(tccommon.ProviderMeta).GetAPIV3Conn()}
	)

	id := d.Id()
	split := strings.Split(id, "+")
	if len(split) != 2 {
		log.Printf("[CRITAL]%s id %s is invalid", logId, id)
		d.SetId("")
		return nil
	}

	eniId := split[0]
	enis, err := service.DescribeEniById(ctx, []string{eniId})
	if err != nil {
		return err
	}

	if len(enis) < 1 {
		d.SetId("")
		return nil
	}

	eni := enis[0]
	if eni.Attachment == nil {
		d.SetId("")
		return nil
	}

	_ = d.Set("eni_id", eni.NetworkInterfaceId)
	_ = d.Set("instance_id", eni.Attachment.InstanceId)

	return nil
}

func resourceTencentCloudEniAttachmentDelete(d *schema.ResourceData, m interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_eni_attachment.delete")()

	var (
		logId   = tccommon.GetLogId(tccommon.ContextNil)
		ctx     = context.WithValue(context.TODO(), tccommon.LogIdKey, logId)
		service = VpcService{client: m.(tccommon.ProviderMeta).GetAPIV3Conn()}
	)

	id := d.Id()
	split := strings.Split(id, "+")
	if len(split) != 2 {
		log.Printf("[CRITAL]%s id %s is invalid", logId, id)
		d.SetId("")
		return nil
	}

	eniId, cvmId := split[0], split[1]
	return service.DetachEniFromCvm(ctx, eniId, cvmId)
}
