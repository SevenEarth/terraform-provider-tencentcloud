// Code generated by iacg; DO NOT EDIT.
package cls

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	clsv20201016 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cls/v20201016"

	tccommon "github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/common"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func ResourceTencentCloudClsNoticeContent() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudClsNoticeContentCreate,
		Read:   resourceTencentCloudClsNoticeContentRead,
		Update: resourceTencentCloudClsNoticeContentUpdate,
		Delete: resourceTencentCloudClsNoticeContentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Notice content name.",
			},

			"type": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Template content language. 0: Chinese 1: English.",
			},

			"notice_contents": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Template detailed configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Channel type. Email: Email; Sms: SMS; WeChat: WeChat; Phone: Telephone; WeCom: Enterprise WeChat; DingTalk: DingTalk; Lark: Feishu; Http: Custom callback.",
						},
						"trigger_content": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Alarm triggered notification content template.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"title": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Notification content template title information. Some notification channel types do not support 'title', please refer to the Tencent Cloud Console page.",
									},
									"content": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Notification content template body information.",
									},
									"headers": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "Request headers: In HTTP requests, request headers contain additional information sent by the client to the server, such as user agent, authorization credentials, expected response format, etc. Only `custom callback` supports this configuration.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"recovery_content": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Template for Alarm Recovery Notification Content.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"title": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Notification content template title information. Some notification channel types do not support 'title', please refer to the Tencent Cloud Console page.",
									},
									"content": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Notification content template body information.",
									},
									"headers": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "Request headers: In HTTP requests, request headers contain additional information sent by the client to the server, such as user agent, authorization credentials, expected response format, etc. Only `custom callback` supports this configuration.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
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

func resourceTencentCloudClsNoticeContentCreate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cls_notice_content.create")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	var (
		noticeContentId string
	)
	var (
		request  = clsv20201016.NewCreateNoticeContentRequest()
		response = clsv20201016.NewCreateNoticeContentResponse()
	)

	if v, ok := d.GetOk("name"); ok {
		request.Name = helper.String(v.(string))
	}

	if v, ok := d.GetOkExists("type"); ok {
		request.Type = helper.IntUint64(v.(int))
	}

	if v, ok := d.GetOk("notice_contents"); ok {
		for _, item := range v.([]interface{}) {
			noticeContentsMap := item.(map[string]interface{})
			noticeContent := clsv20201016.NoticeContent{}
			if v, ok := noticeContentsMap["type"]; ok {
				noticeContent.Type = helper.String(v.(string))
			}
			if triggerContentMap, ok := helper.ConvertInterfacesHeadToMap(noticeContentsMap["trigger_content"]); ok {
				noticeContentInfo := clsv20201016.NoticeContentInfo{}
				if v, ok := triggerContentMap["title"]; ok {
					noticeContentInfo.Title = helper.String(v.(string))
				}
				if v, ok := triggerContentMap["content"]; ok {
					noticeContentInfo.Content = helper.String(v.(string))
				}
				if v, ok := triggerContentMap["headers"]; ok {
					headersSet := v.(*schema.Set).List()
					for i := range headersSet {
						headers := headersSet[i].(string)
						noticeContentInfo.Headers = append(noticeContentInfo.Headers, helper.String(headers))
					}
				}
				noticeContent.TriggerContent = &noticeContentInfo
			}
			if recoveryContentMap, ok := helper.ConvertInterfacesHeadToMap(noticeContentsMap["recovery_content"]); ok {
				noticeContentInfo2 := clsv20201016.NoticeContentInfo{}
				if v, ok := recoveryContentMap["title"]; ok {
					noticeContentInfo2.Title = helper.String(v.(string))
				}
				if v, ok := recoveryContentMap["content"]; ok {
					noticeContentInfo2.Content = helper.String(v.(string))
				}
				if v, ok := recoveryContentMap["headers"]; ok {
					headersSet := v.(*schema.Set).List()
					for i := range headersSet {
						headers := headersSet[i].(string)
						noticeContentInfo2.Headers = append(noticeContentInfo2.Headers, helper.String(headers))
					}
				}
				noticeContent.RecoveryContent = &noticeContentInfo2
			}
			request.NoticeContents = append(request.NoticeContents, &noticeContent)
		}
	}

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseClsV20201016Client().CreateNoticeContentWithContext(ctx, request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s create cls notice content failed, reason:%+v", logId, err)
		return err
	}

	noticeContentId = *response.Response.NoticeContentId

	d.SetId(noticeContentId)

	return resourceTencentCloudClsNoticeContentRead(d, meta)
}

func resourceTencentCloudClsNoticeContentRead(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cls_notice_content.read")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	service := ClsService{client: meta.(tccommon.ProviderMeta).GetAPIV3Conn()}

	noticeContentId := d.Id()

	respData, err := service.DescribeClsNoticeContentById(ctx, noticeContentId)
	if err != nil {
		return err
	}

	if respData == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `cls_notice_content` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}
	if respData.Name != nil {
		_ = d.Set("name", respData.Name)
	}

	if respData.Type != nil {
		_ = d.Set("type", respData.Type)
	}

	noticeContentsList := make([]map[string]interface{}, 0, len(respData.NoticeContents))
	if respData.NoticeContents != nil {
		for _, noticeContents := range respData.NoticeContents {
			noticeContentsMap := map[string]interface{}{}

			if noticeContents.Type != nil {
				noticeContentsMap["type"] = noticeContents.Type
			}

			triggerContentMap := map[string]interface{}{}

			if noticeContents.TriggerContent != nil {
				if noticeContents.TriggerContent.Title != nil {
					triggerContentMap["title"] = noticeContents.TriggerContent.Title
				}

				if noticeContents.TriggerContent.Content != nil {
					triggerContentMap["content"] = noticeContents.TriggerContent.Content
				}

				if noticeContents.TriggerContent.Headers != nil {
					tmpList := make([]string, 0, len(noticeContents.TriggerContent.Headers))
					for _, item := range noticeContents.TriggerContent.Headers {
						tmpList = append(tmpList, *item)
					}

					triggerContentMap["headers"] = tmpList
				}

				noticeContentsMap["trigger_content"] = []interface{}{triggerContentMap}
			}

			recoveryContentMap := map[string]interface{}{}

			if noticeContents.RecoveryContent != nil {
				if noticeContents.RecoveryContent.Title != nil {
					recoveryContentMap["title"] = noticeContents.RecoveryContent.Title
				}

				if noticeContents.RecoveryContent.Content != nil {
					recoveryContentMap["content"] = noticeContents.RecoveryContent.Content
				}

				if noticeContents.RecoveryContent.Headers != nil {
					tmpList := make([]string, 0, len(noticeContents.RecoveryContent.Headers))
					for _, item := range noticeContents.RecoveryContent.Headers {
						tmpList = append(tmpList, *item)
					}

					recoveryContentMap["headers"] = tmpList
				}

				noticeContentsMap["recovery_content"] = []interface{}{recoveryContentMap}
			}

			noticeContentsList = append(noticeContentsList, noticeContentsMap)
		}

		_ = d.Set("notice_contents", noticeContentsList)
	}

	return nil
}

func resourceTencentCloudClsNoticeContentUpdate(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cls_notice_content.update")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)

	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	noticeContentId := d.Id()

	needChange := false
	mutableArgs := []string{"name", "type", "notice_contents"}
	for _, v := range mutableArgs {
		if d.HasChange(v) {
			needChange = true
			break
		}
	}

	if needChange {
		request := clsv20201016.NewModifyNoticeContentRequest()

		request.NoticeContentId = helper.String(noticeContentId)

		if v, ok := d.GetOk("name"); ok {
			request.Name = helper.String(v.(string))
		}

		if v, ok := d.GetOkExists("type"); ok {
			request.Type = helper.IntUint64(v.(int))
		}

		if v, ok := d.GetOk("notice_contents"); ok {
			for _, item := range v.([]interface{}) {
				noticeContentsMap := item.(map[string]interface{})
				noticeContent := clsv20201016.NoticeContent{}
				if v, ok := noticeContentsMap["type"]; ok {
					noticeContent.Type = helper.String(v.(string))
				}
				if triggerContentMap, ok := helper.ConvertInterfacesHeadToMap(noticeContentsMap["trigger_content"]); ok {
					noticeContentInfo := clsv20201016.NoticeContentInfo{}
					if v, ok := triggerContentMap["title"]; ok {
						noticeContentInfo.Title = helper.String(v.(string))
					}
					if v, ok := triggerContentMap["content"]; ok {
						noticeContentInfo.Content = helper.String(v.(string))
					}
					if v, ok := triggerContentMap["headers"]; ok {
						headersSet := v.(*schema.Set).List()
						for i := range headersSet {
							headers := headersSet[i].(string)
							noticeContentInfo.Headers = append(noticeContentInfo.Headers, helper.String(headers))
						}
					}
					noticeContent.TriggerContent = &noticeContentInfo
				}
				if recoveryContentMap, ok := helper.ConvertInterfacesHeadToMap(noticeContentsMap["recovery_content"]); ok {
					noticeContentInfo2 := clsv20201016.NoticeContentInfo{}
					if v, ok := recoveryContentMap["title"]; ok {
						noticeContentInfo2.Title = helper.String(v.(string))
					}
					if v, ok := recoveryContentMap["content"]; ok {
						noticeContentInfo2.Content = helper.String(v.(string))
					}
					if v, ok := recoveryContentMap["headers"]; ok {
						headersSet := v.(*schema.Set).List()
						for i := range headersSet {
							headers := headersSet[i].(string)
							noticeContentInfo2.Headers = append(noticeContentInfo2.Headers, helper.String(headers))
						}
					}
					noticeContent.RecoveryContent = &noticeContentInfo2
				}
				request.NoticeContents = append(request.NoticeContents, &noticeContent)
			}
		}

		err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
			result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseClsV20201016Client().ModifyNoticeContentWithContext(ctx, request)
			if e != nil {
				return tccommon.RetryError(e)
			} else {
				log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
			}
			return nil
		})
		if err != nil {
			log.Printf("[CRITAL]%s update cls notice content failed, reason:%+v", logId, err)
			return err
		}
	}

	return resourceTencentCloudClsNoticeContentRead(d, meta)
}

func resourceTencentCloudClsNoticeContentDelete(d *schema.ResourceData, meta interface{}) error {
	defer tccommon.LogElapsed("resource.tencentcloud_cls_notice_content.delete")()
	defer tccommon.InconsistentCheck(d, meta)()

	logId := tccommon.GetLogId(tccommon.ContextNil)
	ctx := tccommon.NewResourceLifeCycleHandleFuncContext(context.Background(), logId, d, meta)

	noticeContentId := d.Id()

	var (
		request  = clsv20201016.NewDeleteNoticeContentRequest()
		response = clsv20201016.NewDeleteNoticeContentResponse()
	)

	request.NoticeContentId = helper.String(noticeContentId)

	err := resource.Retry(tccommon.WriteRetryTimeout, func() *resource.RetryError {
		result, e := meta.(tccommon.ProviderMeta).GetAPIV3Conn().UseClsV20201016Client().DeleteNoticeContentWithContext(ctx, request)
		if e != nil {
			return tccommon.RetryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s delete cls notice content failed, reason:%+v", logId, err)
		return err
	}

	_ = response
	return nil
}