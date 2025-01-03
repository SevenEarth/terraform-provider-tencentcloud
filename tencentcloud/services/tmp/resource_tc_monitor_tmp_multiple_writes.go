// Code generated by iacg; DO NOT EDIT.
package tmp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	monitorv20180724 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
	svcmonitor "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/services/monitor"
)

func ResourceTencentCloudMonitorTmpMultipleWrites() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudMonitorTmpMultipleWritesCreate,
		Read:   resourceTencentCloudMonitorTmpMultipleWritesRead,
		Update: resourceTencentCloudMonitorTmpMultipleWritesUpdate,
		Delete: resourceTencentCloudMonitorTmpMultipleWritesDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Instance id.",
			},

			"remote_writes": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Data multiple write configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Data multiple write url.",
						},
						"url_relabel_config": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "RelabelConfig.",
						},
						"basic_auth": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Authentication information.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"user_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "User name.",
									},
									"password": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Password.",
									},
								},
							},
						},
						"max_block_size": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Maximum block.",
						},
						"label": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Label.",
						},
						"headers": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "HTTP additional headers.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "HTTP header key.",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "HTTP header value.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceTencentCloudMonitorTmpMultipleWritesCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_monitor_tmp_multiple_writes.create")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	var (
		instanceId string
		url        string
	)
	var (
		request  = monitorv20180724.NewModifyRemoteURLsRequest()
		response = monitorv20180724.NewModifyRemoteURLsResponse()
	)

	if v, ok := d.GetOk("instance_id"); ok {
		instanceId = v.(string)
	}

	request.InstanceId = helper.String(instanceId)

	if v, ok := d.GetOk("remote_writes"); ok {
		for _, item := range v.([]interface{}) {
			remoteWritesMap := item.(map[string]interface{})
			remoteWrite := monitorv20180724.RemoteWrite{}
			if v, ok := remoteWritesMap["url"].(string); ok && v != "" {
				remoteWrite.URL = helper.String(v)
				url = v
			}
			if v, ok := remoteWritesMap["url_relabel_config"].(string); ok && v != "" {
				remoteWrite.URLRelabelConfig = helper.String(v)
			}
			if basicAuthMap, ok := helper.ConvertInterfacesHeadToMap(remoteWritesMap["basic_auth"]); ok {
				basicAuth := monitorv20180724.BasicAuth{}
				if v, ok := basicAuthMap["user_name"].(string); ok && v != "" {
					basicAuth.UserName = helper.String(v)
				}
				if v, ok := basicAuthMap["password"].(string); ok && v != "" {
					basicAuth.Password = helper.String(v)
				}
				remoteWrite.BasicAuth = &basicAuth
			}
			if v, ok := remoteWritesMap["max_block_size"].(string); ok && v != "" {
				remoteWrite.MaxBlockSize = helper.String(v)
			}
			if v, ok := remoteWritesMap["label"].(string); ok && v != "" {
				remoteWrite.Label = helper.String(v)
			}
			if v, ok := remoteWritesMap["headers"]; ok {
				for _, item := range v.([]interface{}) {
					headersMap := item.(map[string]interface{})
					remoteWriteHeader := monitorv20180724.RemoteWriteHeader{}
					if v, ok := headersMap["key"].(string); ok && v != "" {
						remoteWriteHeader.Key = helper.String(v)
					}
					if v, ok := headersMap["value"].(string); ok && v != "" {
						remoteWriteHeader.Value = helper.String(v)
					}
					remoteWrite.Headers = append(remoteWrite.Headers, &remoteWriteHeader)
				}
			}
			request.RemoteWrites = append(request.RemoteWrites, &remoteWrite)
		}
	}

	reqErr := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseMonitorV20180724Client().ModifyRemoteURLsWithContext(ctx, request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if reqErr != nil {
		log.Printf("[CRITAL]%s create monitor tmp multiple writes failed, reason:%+v", logId, reqErr)
		return reqErr
	}

	_ = response

	d.SetId(strings.Join([]string{instanceId, url}, tccommon.FILED_SP))

	return resourceTencentCloudMonitorTmpMultipleWritesRead(d, meta)
}

func resourceTencentCloudMonitorTmpMultipleWritesRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_monitor_tmp_multiple_writes.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	service := svcmonitor.NewMonitorService(meta.(tccommon.ProviderMeta).GetAPIV3Conn())

	idSplit := strings.SplitN(d.Id(), tccommon.FILED_SP, 2)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	instanceId := idSplit[0]
	url := idSplit[1]

	_ = d.Set("instance_id", instanceId)

	respData, err := service.DescribeMonitorTmpMultipleWritesById(ctx, instanceId, url)
	if err != nil {
		return err
	}

	if respData == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `monitor_tmp_multiple_writes` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}
	remoteWritesList := make([]map[string]interface{}, 0, len(respData.RemoteWrites))
	if respData.RemoteWrites != nil {
		for _, remoteWrites := range respData.RemoteWrites {
			remoteWritesMap := map[string]interface{}{}

			if remoteWrites.URL != nil {
				remoteWritesMap["url"] = remoteWrites.URL
				url = *remoteWrites.URL
			}

			if remoteWrites.URLRelabelConfig != nil {
				remoteWritesMap["url_relabel_config"] = remoteWrites.URLRelabelConfig
			}

			basicAuthMap := map[string]interface{}{}

			if remoteWrites.BasicAuth != nil {
				if remoteWrites.BasicAuth.UserName != nil {
					basicAuthMap["user_name"] = remoteWrites.BasicAuth.UserName
				}

				if remoteWrites.BasicAuth.Password != nil {
					basicAuthMap["password"] = remoteWrites.BasicAuth.Password
				}

				remoteWritesMap["basic_auth"] = []interface{}{basicAuthMap}
			}

			if remoteWrites.MaxBlockSize != nil {
				remoteWritesMap["max_block_size"] = remoteWrites.MaxBlockSize
			}

			if remoteWrites.Label != nil {
				remoteWritesMap["label"] = remoteWrites.Label
			}

			headersList := make([]map[string]interface{}, 0, len(remoteWrites.Headers))
			if remoteWrites.Headers != nil {
				for _, headers := range remoteWrites.Headers {
					headersMap := map[string]interface{}{}

					if headers.Key != nil {
						headersMap["key"] = headers.Key
					}

					if headers.Value != nil {
						headersMap["value"] = headers.Value
					}

					headersList = append(headersList, headersMap)
				}

				remoteWritesMap["headers"] = headersList
			}
			remoteWritesList = append(remoteWritesList, remoteWritesMap)
		}

		_ = d.Set("remote_writes", remoteWritesList)
	}

	return nil
}

func resourceTencentCloudMonitorTmpMultipleWritesUpdate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_monitor_tmp_multiple_writes.update")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	immutableArgs := []string{"instance_id"}
	for _, v := range immutableArgs {
		if d.HasChange(v) {
			return fmt.Errorf("argument `%s` cannot be changed", v)
		}
	}
	idSplit := strings.SplitN(d.Id(), tccommon.FILED_SP, 2)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	instanceId := idSplit[0]
	url := idSplit[1]

	needChange := false
	mutableArgs := []string{"remote_writes"}
	for _, v := range mutableArgs {
		if d.HasChange(v) {
			needChange = true
			break
		}
	}

	if needChange {
		request := monitorv20180724.NewModifyRemoteURLsRequest()

		request.InstanceId = helper.String(instanceId)

		if v, ok := d.GetOk("remote_writes"); ok {
			for _, item := range v.([]interface{}) {
				remoteWritesMap := item.(map[string]interface{})
				remoteWrite := monitorv20180724.RemoteWrite{}
				remoteWrite.URL = helper.String(url)
				if v, ok := remoteWritesMap["url_relabel_config"].(string); ok && v != "" {
					remoteWrite.URLRelabelConfig = helper.String(v)
				}
				if basicAuthMap, ok := helper.ConvertInterfacesHeadToMap(remoteWritesMap["basic_auth"]); ok {
					basicAuth := monitorv20180724.BasicAuth{}
					if v, ok := basicAuthMap["user_name"].(string); ok && v != "" {
						basicAuth.UserName = helper.String(v)
					}
					if v, ok := basicAuthMap["password"].(string); ok && v != "" {
						basicAuth.Password = helper.String(v)
					}
					remoteWrite.BasicAuth = &basicAuth
				}
				if v, ok := remoteWritesMap["max_block_size"].(string); ok && v != "" {
					remoteWrite.MaxBlockSize = helper.String(v)
				}
				if v, ok := remoteWritesMap["label"].(string); ok && v != "" {
					remoteWrite.Label = helper.String(v)
				}
				if v, ok := remoteWritesMap["headers"]; ok {
					for _, item := range v.([]interface{}) {
						headersMap := item.(map[string]interface{})
						remoteWriteHeader := monitorv20180724.RemoteWriteHeader{}
						if v, ok := headersMap["key"].(string); ok && v != "" {
							remoteWriteHeader.Key = helper.String(v)
						}
						if v, ok := headersMap["value"].(string); ok && v != "" {
							remoteWriteHeader.Value = helper.String(v)
						}
						remoteWrite.Headers = append(remoteWrite.Headers, &remoteWriteHeader)
					}
				}
				request.RemoteWrites = append(request.RemoteWrites, &remoteWrite)
			}
		}

		reqErr := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
			result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseMonitorV20180724Client().ModifyRemoteURLsWithContext(ctx, request)
			if e != nil {
				return tccommon.RetryError(e)
			} else {
				log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
			}
			return nil
		})
		if reqErr != nil {
			log.Printf("[CRITAL]%s update monitor tmp multiple writes failed, reason:%+v", logId, reqErr)
			return reqErr
		}
	}

	return resourceTencentCloudMonitorTmpMultipleWritesRead(d, meta)
}

func resourceTencentCloudMonitorTmpMultipleWritesDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_monitor_tmp_multiple_writes.delete")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	idSplit := strings.SplitN(d.Id(), tccommon.FILED_SP, 2)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	instanceId := idSplit[0]
	url := idSplit[1]

	var (
		request  = monitorv20180724.NewModifyRemoteURLsRequest()
		response = monitorv20180724.NewModifyRemoteURLsResponse()
	)

	request.InstanceId = helper.String(instanceId)

	reqErr := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseMonitorV20180724Client().ModifyRemoteURLsWithContext(ctx, request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if reqErr != nil {
		log.Printf("[CRITAL]%s delete monitor tmp multiple writes failed, reason:%+v", logId, reqErr)
		return reqErr
	}

	_ = response
	_ = url
	return nil
}