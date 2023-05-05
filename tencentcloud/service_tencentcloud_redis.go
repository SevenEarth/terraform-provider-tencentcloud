package tencentcloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/connectivity"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/ratelimit"
)

type RedisService struct {
	client  *connectivity.TencentCloudClient
	zoneMap map[int64]string
}

type TencentCloudRedisDetail struct {
	RedisId          string
	Name             string
	Zone             string
	ProjectId        int64
	TypeId           int64
	Type             string
	MemSize          int64
	Status           string
	VpcId            string
	SubnetId         string
	Ip               string
	Port             int64
	RedisShardNum    int64
	RedisReplicasNum int64
	CreateTime       string
	Tags             map[string]string
	BillingMode      string
	NodeInfo         []map[string]interface{}
}

func (me *RedisService) fullZoneId() (errRet error) {
	if me.zoneMap == nil {
		me.zoneMap = make(map[int64]string)
	}
	if len(me.zoneMap) != 0 {
		return
	}
	response, err := me.client.UseCvmClient().DescribeZones(cvm.NewDescribeZonesRequest())
	if err != nil {
		return err
	}
	for _, item := range response.Response.ZoneSet {
		if zoneId, err := strconv.ParseInt(*item.ZoneId, 10, 64); err != nil {
			return fmt.Errorf("[sdk]DescribeZones return ZoneId is not illegal,%s", *item.ZoneId)
		} else {
			me.zoneMap[zoneId] = *item.Zone
		}
	}

	return nil
}

func (me *RedisService) getZoneId(name string) (id int64, errRet error) {
	if errRet = me.fullZoneId(); errRet != nil {
		return
	}
	for key, value := range me.zoneMap {
		if value == name {
			id = key
			return
		}
	}
	errRet = fmt.Errorf("this redis zone %s not support yet", name)
	return
}

func (me *RedisService) getZoneName(id int64) (name string, errRet error) {
	if errRet = me.fullZoneId(); errRet != nil {
		return
	}
	name = me.zoneMap[id]
	if name == "" {
		errRet = fmt.Errorf("this redis zoneid %d not support yet", id)
	}
	return
}

func (me *RedisService) DescribeRedisZoneConfig(ctx context.Context) (sellConfigures []*redis.RegionConf, errRet error) {
	logId := getLogId(ctx)
	request := redis.NewDescribeProductInfoRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeProductInfo(request)
	if err != nil {
		errRet = err
		return
	}
	sellConfigures = response.Response.RegionSet
	return
}

func (me *RedisService) DescribeInstances(ctx context.Context, zoneName, searchKey string,
	projectId, needLimit int64) (instances []TencentCloudRedisDetail, errRet error) {

	logId := getLogId(ctx)

	var zoneId int64 = -1

	if zoneName != "" {
		zoneId, errRet = me.getZoneId(zoneName)
		if errRet != nil {
			return
		}
	}

	listInitSize := map[bool]int64{true: 500, false: needLimit}[needLimit > 500 || needLimit < 1]
	instances = make([]TencentCloudRedisDetail, 0, listInitSize)

	request := redis.NewDescribeInstancesRequest()

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	var (
		limit, offset  uint64 = 2, 0
		leftNumber     int64
		leftNumberInit bool
	)

	request.Limit = &limit
	request.Offset = &offset

needMoreItems:
	if searchKey != "" {
		request.SearchKey = &searchKey
	}
	if projectId >= 0 {
		request.ProjectIds = []*int64{&projectId}
	}
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeInstances(request)
	if err != nil {
		errRet = err
		return
	}
	if !leftNumberInit {
		leftNumber = *response.Response.TotalCount
		leftNumberInit = true
	}
	leftNumber = leftNumber - int64(limit)
	offset = offset + limit
	for _, item := range response.Response.InstanceSet {
		if zoneId != -1 && *item.ZoneId != zoneId {
			continue
		}
		var instance TencentCloudRedisDetail
		instance.Type = REDIS_NAMES[*item.Type]
		if REDIS_STATUS[*item.Status] == "" {
			instance.Status = "unknown"
		} else {
			instance.Status = REDIS_STATUS[*item.Status]
		}

		name, err := me.getZoneName(*item.ZoneId)
		if err != nil {
			errRet = err
			return
		}

		instance.Zone = name
		instance.CreateTime = *item.Createtime
		instance.Ip = *item.WanIp
		instance.MemSize = *item.RedisShardSize
		instance.Name = *item.InstanceName
		instance.Port = *item.Port
		instance.ProjectId = *item.ProjectId
		instance.RedisId = *item.InstanceId
		instance.SubnetId = *item.UniqSubnetId
		instance.VpcId = *item.UniqVpcId
		instance.BillingMode = REDIS_CHARGE_TYPE_NAME[*item.BillingMode]

		instance.TypeId = *item.Type
		if item.RedisReplicasNum != nil {
			instance.RedisReplicasNum = *item.RedisReplicasNum
		}
		if item.RedisShardNum != nil {
			instance.RedisShardNum = *item.RedisShardNum
		}

		if item.NodeSet != nil {
			nodeInfos := make([]map[string]interface{}, 0, len(item.NodeSet))
			for i := range item.NodeSet {
				dMap := make(map[string]interface{})
				nodeInfo := item.NodeSet[i]
				if *nodeInfo.NodeType == 0 {
					dMap["master"] = true
				} else {
					dMap["master"] = false
				}
				dMap["id"] = *nodeInfo.NodeId
				dMap["zone_id"] = *nodeInfo.ZoneId
				nodeInfos = append(nodeInfos, dMap)
			}
			instance.NodeInfo = nodeInfos
		}
		instance.Tags = make(map[string]string, len(item.InstanceTags))
		for _, tag := range item.InstanceTags {
			if tag.TagKey == nil {
				return nil, fmt.Errorf("[CRITAL]%s api[%s] redis instance tag key is nil", logId, request.GetAction())
			}
			if tag.TagValue == nil {
				return nil, fmt.Errorf("[CRITAL]%s api[%s] redis instance tag value is nil", logId, request.GetAction())
			}

			instance.Tags[*tag.TagKey] = *tag.TagValue
		}

		instances = append(instances, instance)

		if needLimit > 0 && int64(len(instances)) >= needLimit {
			return
		}
	}
	if leftNumber < 0 {
		return
	} else {
		goto needMoreItems
	}
}

func (me *RedisService) CreateInstances(ctx context.Context,
	zoneName string, typeId int64, password, vpcId, subnetId, redisName string,
	memSize, projectId, port int64,
	securityGroups []string,
	redisShardNum,
	redisReplicasNum int,
	chargeTypeID int64,
	chargePeriod uint64,
	nodeInfo []*redis.RedisNodeInfo,
	noAuth bool,
	autoRenewFlag int,
	replicasReadonly bool,
	paramsTemplateId string,
) (instanceIds []*string, errRet error) {

	logId := getLogId(ctx)
	request := redis.NewCreateInstancesRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	// zone
	var intZoneId int64
	intZoneId, errRet = me.getZoneId(zoneName)
	if errRet != nil {
		return
	}
	request.ZoneId = helper.Int64Uint64(intZoneId)
	request.TypeId = helper.Int64Uint64(typeId)

	// vpc
	if (vpcId == "" && subnetId != "") || (vpcId != "" && subnetId == "") {
		errRet = fmt.Errorf("redis need vpcId and subnetId both set or none")
		return
	}
	if vpcId != "" && subnetId != "" {
		request.VpcId = &vpcId
		request.SubnetId = &subnetId
	} else {
		if len(securityGroups) > 0 {
			errRet = fmt.Errorf("redis need empty security_groups if vpc_id and subnet_id is empty")
			return
		}
	}

	if projectId >= 0 {
		request.ProjectId = &projectId
	}

	var (
		vport           = uint64(port)
		umemSize        = uint64(memSize)
		goodsNum uint64 = 1
	)
	request.VPort = &vport
	request.MemSize = &umemSize
	request.BillingMode = &chargeTypeID
	request.GoodsNum = &goodsNum
	request.Period = &chargePeriod
	if redisShardNum > 0 {
		request.RedisShardNum = helper.IntInt64(redisShardNum)
	}
	if redisReplicasNum > 0 {
		request.RedisReplicasNum = helper.IntInt64(redisReplicasNum)
	}
	if redisName != "" {
		request.InstanceName = &redisName
	}

	request.Password = &password

	if len(securityGroups) > 0 {
		request.SecurityGroupIdList = make([]*string, 0, len(securityGroups))
		for v := range securityGroups {
			request.SecurityGroupIdList = append(request.SecurityGroupIdList, &securityGroups[v])
		}
	}

	if len(nodeInfo) > 0 {
		request.NodeSet = nodeInfo
	}

	if noAuth {
		request.NoAuth = &noAuth
	}
	if chargeTypeID == REDIS_CHARGE_TYPE_ID[REDIS_CHARGE_TYPE_PREPAID] {
		request.AutoRenew = helper.IntUint64(autoRenewFlag)
	}
	if replicasReadonly {
		request.ReplicasReadonly = &replicasReadonly
	}
	if paramsTemplateId != "" {
		request.TemplateId = &paramsTemplateId
	}
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().CreateInstances(request)
	if err != nil {
		errRet = err
		return
	}
	log.Println(response.ToJsonString())
	instanceIds = response.Response.InstanceIds
	return
}

func (me *RedisService) CheckRedisOnlineOk(ctx context.Context, redisId string, retryTimeout time.Duration) (has bool,
	online bool,
	info *redis.InstanceSet,
	errRet error) {

	logId := getLogId(ctx)

	request := redis.NewDescribeInstancesRequest()

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	request.InstanceId = &redisId

	// Post https://cdb.tencentcloudapi.com/: always get "Gateway Time-out"
	var response *redis.DescribeInstancesResponse
	err := resource.Retry(retryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		result, e := me.client.UseRedisClient().DescribeInstances(request)
		if e != nil {
			log.Printf("[CRITAL]%s CheckRedisOnlineOk fail, reason:%s\n", logId, e.Error())
			return retryError(e)
		}
		response = result

		if len(response.Response.InstanceSet) == 0 {
			has = false
			return resource.NonRetryableError(fmt.Errorf("instance %s not exist", redisId))
		}

		info = response.Response.InstanceSet[0]
		has = true

		if *info.Status == REDIS_STATUS_ONLINE {
			online = true
			return nil
		}

		if *info.Status == REDIS_STATUS_INIT || *info.Status == REDIS_STATUS_PROCESSING {
			online = false
			return resource.RetryableError(fmt.Errorf("istance %s status is %d, retrying", redisId, *info.Status))
		}

		return nil
	})

	if err != nil {
		errRet = err
		return
	}

	return
}

func (me *RedisService) CheckRedisUpdateOk(ctx context.Context, redisId string) (errRet error) {
	var startUpdate bool
	logId := getLogId(ctx)
	request := redis.NewDescribeInstancesRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	request.InstanceId = &redisId
	errRet = resource.Retry(readRetryTimeout*20, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		result, err := me.client.UseRedisClient().DescribeInstances(request)
		if err != nil {
			return retryError(err)
		}
		if len(result.Response.InstanceSet) == 0 {
			return resource.NonRetryableError(fmt.Errorf("redis %s not exist", redisId))
		}
		info := result.Response.InstanceSet[0]
		if !startUpdate && *info.Status == REDIS_STATUS_ONLINE {
			return resource.RetryableError(fmt.Errorf("waiting for upgrade start"))
		}
		startUpdate = true
		if *info.Status == REDIS_STATUS_PROCESSING || *info.Status == REDIS_STATUS_INIT {
			return resource.RetryableError(fmt.Errorf("instance %s status is %d", redisId, *info.Status))
		}
		return nil
	})

	return
}

func (me *RedisService) CheckRedisDestroyOk(ctx context.Context, redisId string) (has bool,
	isolated bool,
	errRet error) {

	logId := getLogId(ctx)

	request := redis.NewDescribeInstancesRequest()

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	request.InstanceId = &redisId

	// Post https://cdb.tencentcloudapi.com/: always get "Gateway Time-out"
	var response *redis.DescribeInstancesResponse
	err := resource.Retry(10*readRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		result, e := me.client.UseRedisClient().DescribeInstances(request)
		if e != nil {
			log.Printf("[CRITAL]%s CheckRedisDestroyOk fail, reason:%s\n", logId, e.Error())
			return retryError(e)
		}
		response = result
		return nil
	})

	if err != nil {
		errRet = err
		return
	}

	if len(response.Response.InstanceSet) == 0 {
		has = false
		return
	}

	if len(response.Response.InstanceSet) != 1 {
		errRet = fmt.Errorf("redis DescribeInstances one id get %d redis info", len(response.Response.InstanceSet))
		return
	}

	has = true

	info := response.Response.InstanceSet[0]
	if *info.Status <= REDIS_STATUS_ISOLATE {
		isolated = true
		return
	} else {
		isolated = false
		return
	}
}

func (me *RedisService) DescribeInstanceDealDetail(ctx context.Context, dealId string) (done bool, redisId string, errRet error) {
	logId := getLogId(ctx)
	request := redis.NewDescribeInstanceDealDetailRequest()

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	request.DealIds = []*string{&dealId}

	// Post https://cdb.tencentcloudapi.com/: always get "Gateway Time-out"
	var response *redis.DescribeInstanceDealDetailResponse
	err := resource.Retry(10*readRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		result, e := me.client.UseRedisClient().DescribeInstanceDealDetail(request)
		if e != nil {
			log.Printf("[CRITAL]%s DescribeInstanceDealDetail fail, reason:%s\n", logId, e.Error())
			return retryError(e)
		}
		response = result
		return nil
	})

	if err != nil {
		errRet = err
		return
	}

	if len(response.Response.DealDetails) != 1 {
		errRet = fmt.Errorf("Redis api DescribeInstanceDealDetail one dealId[%s] return %d deal infos.",
			dealId, len(response.Response.DealDetails))
		return
	}

	dealDetail := response.Response.DealDetails[0]
	status := *dealDetail.Status

	if status == REDIS_ORDER_SUCCESS_DELIVERY {

		if len(dealDetail.InstanceIds) != 1 {
			errRet = fmt.Errorf("redis one dealid give %d redis id", len(dealDetail.InstanceIds))
			return
		}
		redisId = *dealDetail.InstanceIds[0]
		done = true
		return
	}
	if status < REDIS_ORDER_SUCCESS_DELIVERY || status == REDIS_ORDER_PAYMENT {
		return
	}
	errRet = fmt.Errorf("redis instance delivery failure, deal status is %d", status)
	return
}

func (me *RedisService) ModifyInstanceName(ctx context.Context, redisId string, name string) (errRet error) {
	logId := getLogId(ctx)
	request := redis.NewModifyInstanceRequest()

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	op := "rename"
	request.InstanceName = &name
	request.Operation = &op
	request.InstanceId = &redisId
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().ModifyInstance(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	errRet = err
	return
}

func (me *RedisService) ModifyInstanceProjectId(ctx context.Context, redisId string, projectId int64) (errRet error) {
	logId := getLogId(ctx)
	request := redis.NewModifyInstanceRequest()

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	op := "modifyProject"
	request.ProjectId = &projectId
	request.Operation = &op
	request.InstanceId = &redisId
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().ModifyInstance(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	errRet = err
	return
}

func (me *RedisService) DescribeInstanceSecurityGroup(ctx context.Context, redisId string) (sg []string, errRet error) {
	logId := getLogId(ctx)
	request := redis.NewDescribeInstanceSecurityGroupRequest()
	request.InstanceIds = []*string{&redisId}
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeInstanceSecurityGroup(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	if err != nil {
		errRet = err
		return
	}

	if len(response.Response.InstanceSecurityGroupsDetail) > 0 {
		for _, item := range response.Response.InstanceSecurityGroupsDetail {
			if *item.InstanceId == redisId {
				sg = make([]string, 0, len(item.SecurityGroupDetails))
				for _, v := range item.SecurityGroupDetails {
					sg = append(sg, *v.SecurityGroupId)
				}
				break
			}
		}
	}
	return
}

// DescribeDBSecurityGroups support query different type of DB by passing product name
func (me *RedisService) DescribeDBSecurityGroups(ctx context.Context, product string, instanceId string) (sg []string, errRet error) {
	logId := getLogId(ctx)
	request := redis.NewDescribeDBSecurityGroupsRequest()
	request.Product = &product
	request.InstanceId = &instanceId
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeDBSecurityGroups(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	if err != nil {
		errRet = err
		return
	}

	groups := response.Response.Groups
	if len(groups) > 0 {
		for i := range groups {
			sg = append(sg, *groups[i].SecurityGroupId)
		}
	}
	return
}

func (me *RedisService) ModifyDBInstanceSecurityGroups(ctx context.Context, product string, instanceId string, securityGroupIds []*string) (errRet error) {
	logId := getLogId(ctx)
	request := redis.NewModifyDBInstanceSecurityGroupsRequest()
	request.Product = &product
	request.InstanceId = &instanceId
	request.SecurityGroupIds = securityGroupIds

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().ModifyDBInstanceSecurityGroups(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	errRet = err
	return
}

func (me *RedisService) DestroyPostpaidInstance(ctx context.Context, redisId string) (taskId int64, errRet error) {
	logId := getLogId(ctx)
	request := redis.NewDestroyPostpaidInstanceRequest()
	request.InstanceId = &redisId
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DestroyPostpaidInstance(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	} else {
		errRet = err
		return
	}

	taskId = *response.Response.TaskId
	return
}

func (me *RedisService) DestroyPrepaidInstance(ctx context.Context, redisId string) (dealId string, errRet error) {
	logId := getLogId(ctx)
	request := redis.NewDestroyPrepaidInstanceRequest()
	request.InstanceId = &redisId
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	// For prepaid instance, deal status synchronization will take some time so need to retry.
	var response *redis.DestroyPrepaidInstanceResponse
	err := resource.Retry(5*writeRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		result, e := me.client.UseRedisClient().DestroyPrepaidInstance(request)
		if e != nil {
			log.Printf("[CRITAL]%s DestroyPrepaidInstance fail, reason:%s\n", logId, e.Error())
			return retryError(e)
		}
		response = result
		return nil
	})
	if err == nil {
		dealId = *response.Response.DealId
	} else {
		errRet = err
		return
	}

	return
}

func (me *RedisService) CleanUpInstance(ctx context.Context, redisId string) (taskId int64, errRet error) {
	logId := getLogId(ctx)
	request := redis.NewCleanUpInstanceRequest()
	request.InstanceId = &redisId
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	// Cleaning up action for prepaid instances needs to retry.
	var response *redis.CleanUpInstanceResponse
	err := resource.Retry(6*writeRetryTimeout, func() *resource.RetryError {
		ratelimit.Check(request.GetAction())
		result, e := me.client.UseRedisClient().CleanUpInstance(request)
		if e != nil {
			log.Printf("[CRITAL]%s CleanUpInstance fail, reason:%s\n", logId, e.Error())
			return retryError(e)
		}
		response = result
		return nil
	})
	if err != nil {
		errRet = err
		return
	}
	taskId = *response.Response.TaskId
	return
}

func (me *RedisService) UpgradeInstance(ctx context.Context, redisId string, newMemSize, redisShardNum, redisReplicasNum int, nodeSet []*redis.RedisNodeInfo) (dealId string, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewUpgradeInstanceRequest()
	request.InstanceId = &redisId
	if newMemSize > 0 {
		request.MemSize = helper.IntUint64(newMemSize)
	}
	if redisShardNum > 0 {
		request.RedisShardNum = helper.IntUint64(redisShardNum)
	}
	if redisReplicasNum != 0 {
		request.RedisReplicasNum = helper.IntUint64(redisReplicasNum)
	}
	if len(nodeSet) > 0 {
		request.NodeSet = nodeSet
	}

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().UpgradeInstance(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	} else {
		errRet = err
		return
	}

	dealId = *response.Response.DealId
	return
}

func (me *RedisService) DescribeTaskInfo(ctx context.Context, redisId string, taskId int64) (ok bool, errRet error) {
	logId := getLogId(ctx)
	var uintTaskId = uint64(taskId)
	request := redis.NewDescribeTaskInfoRequest()
	request.TaskId = &uintTaskId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeTaskInfo(request)

	if err != nil {
		errRet = err
		return
	}
	if *response.Response.Status == REDIS_TASK_RUNNING || *response.Response.Status == REDIS_TASK_PREPARING {
		return
	}
	if *response.Response.Status == REDIS_TASK_SUCCEED {
		ok = true
		return
	}
	errRet = fmt.Errorf("redis task exe fail, task status is %s", *response.Response.Status)
	return
}

func (me *RedisService) ResetPassword(ctx context.Context, redisId string, newPassword string, noAuth bool) (taskId int64, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewResetPasswordRequest()
	request.InstanceId = &redisId
	request.Password = &newPassword
	request.NoAuth = &noAuth
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())

	response, err := me.client.UseRedisClient().ResetPassword(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	} else {
		errRet = err
		return
	}

	taskId = *response.Response.TaskId
	return

}

func (me *RedisService) ModifyAutoBackupConfig(ctx context.Context, redisId string, weekDays []string, timePeriod string) (errRet error) {
	logId := getLogId(ctx)

	request := redis.NewModifyAutoBackupConfigRequest()
	request.InstanceId = &redisId
	request.WeekDays = make([]*string, 0, len(weekDays))
	for index := range weekDays {
		request.WeekDays = append(request.WeekDays, &weekDays[index])
	}
	request.TimePeriod = &timePeriod
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().ModifyAutoBackupConfig(request)
	errRet = err
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	}
	return
}

func (me *RedisService) DescribeAutoBackupConfig(ctx context.Context, redisId string) (weekDays []string, timePeriod string, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDescribeAutoBackupConfigRequest()
	request.InstanceId = &redisId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeAutoBackupConfig(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	}
	if err != nil {
		errRet = err
		return
	}

	timePeriod = *response.Response.TimePeriod

	if len(response.Response.WeekDays) > 0 {
		weekDays = make([]string, 0, len(response.Response.WeekDays))
		for _, v := range response.Response.WeekDays {
			weekDays = append(weekDays, *v)
		}
	}
	return
}

func (me *RedisService) DescribeParamTemplates(ctx context.Context, request *redis.DescribeParamTemplatesRequest) (params []*redis.ParamTemplateInfo, errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeParamTemplates(request)

	if err != nil {
		errRet = err
		return
	}

	params = response.Response.Items

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *RedisService) DescribeParamTemplateById(ctx context.Context, id string) (params *redis.ParamTemplateInfo, errRet error) {
	request := redis.NewDescribeParamTemplatesRequest()

	request.TemplateIds = []*string{&id}

	result, err := me.DescribeParamTemplates(ctx, request)

	if err != nil {
		errRet = err
		return
	}

	if len(result) == 0 {
		return
	}

	params = result[0]

	return
}

func (me *RedisService) ApplyParamsTemplate(ctx context.Context, request *redis.ApplyParamsTemplateRequest) (taskIds []*int64, errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().ApplyParamsTemplate(request)

	if err != nil {
		errRet = err
		return
	}

	taskIds = response.Response.TaskIds

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *RedisService) DescribeParamTemplateInfo(ctx context.Context, templateId string) (info *redis.DescribeParamTemplateInfoResponseParams, errRet error) {
	logId := getLogId(ctx)
	request := redis.NewDescribeParamTemplateInfoRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	request.TemplateId = &templateId

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeParamTemplateInfo(request)

	if err != nil {
		errRet = err
		return
	}

	info = response.Response

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *RedisService) CreateParamTemplate(ctx context.Context, request *redis.CreateParamTemplateRequest) (id string, errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().CreateParamTemplate(request)

	if err != nil {
		errRet = err
		return
	}

	if response.Response == nil {
		errRet = fmt.Errorf("[%s] returns nil response", request.GetAction())
		return
	}

	id = *response.Response.TemplateId

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *RedisService) ModifyParamTemplate(ctx context.Context, request *redis.ModifyParamTemplateRequest) (errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().ModifyParamTemplate(request)

	if err != nil {
		errRet = err
		return
	}

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *RedisService) DeleteParamTemplate(ctx context.Context, request *redis.DeleteParamTemplateRequest) (errRet error) {
	logId := getLogId(ctx)
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DeleteParamTemplate(request)

	if err != nil {
		errRet = err
		return
	}

	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return
}

func (me *RedisService) DescribeRedisAccountById(ctx context.Context, instanceId, accountName string) (account *redis.Account, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDescribeInstanceAccountRequest()
	request.InstanceId = &instanceId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	var offset int64 = 0
	var limit int64 = 50
	for {
		request.Offset = &offset
		request.Limit = &limit
		ratelimit.Check(request.GetAction())
		response, err := me.client.UseRedisClient().DescribeInstanceAccount(request)
		if err != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), err.Error())
			errRet = err
			return
		}
		log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

		if response == nil || len(response.Response.Accounts) < 1 {
			break
		}
		for _, v := range response.Response.Accounts {
			if *v.AccountName == accountName {
				account = v
				return
			}
		}
		if len(response.Response.Accounts) < int(limit) {
			break
		}
		offset += limit
	}

	return
}

func (me *RedisService) DeleteRedisAccountById(ctx context.Context, instanceId, accountName string) (taskId int64, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDeleteInstanceAccountRequest()
	request.InstanceId = &instanceId
	request.AccountName = &accountName

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseRedisClient().DeleteInstanceAccount(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	taskId = *response.Response.TaskId

	return
}

func (me *RedisService) RedisAccountStateRefreshFunc(instanceId, accountName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ctx := contextNil

		object, err := me.DescribeRedisAccountById(ctx, instanceId, accountName)
		if err != nil {
			return nil, "", err
		}

		if object == nil {
			return nil, "", nil
		}

		return object, helper.PString(helper.String(strconv.FormatInt(*object.Status, 10))), nil
	}
}

func (me *RedisService) DescribeRedisInstanceById(ctx context.Context, instanceId string) (param *redis.InstanceSet, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDescribeInstancesRequest()
	request.InstanceId = &instanceId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseRedisClient().DescribeInstances(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if len(response.Response.InstanceSet) < 1 {
		return
	}

	param = response.Response.InstanceSet[0]
	return
}

func (me *RedisService) DescribeRedisParamById(ctx context.Context, instanceId string) (params map[string]interface{}, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDescribeInstanceParamsRequest()
	request.InstanceId = &instanceId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseRedisClient().DescribeInstanceParams(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	param := response.Response
	instanceParams := make(map[string]interface{})
	if param.InstanceEnumParam != nil {
		for _, v := range param.InstanceEnumParam {
			key := *v.ParamName
			value := *v.CurrentValue
			instanceParams[key] = value
		}
	}
	if param.InstanceIntegerParam != nil {
		for _, v := range param.InstanceIntegerParam {
			key := *v.ParamName
			value := *v.CurrentValue
			instanceParams[key] = value
		}
	}
	if param.InstanceMultiParam != nil {
		for _, v := range param.InstanceMultiParam {
			key := *v.ParamName
			value := *v.CurrentValue
			instanceParams[key] = value
		}
	}
	if param.InstanceTextParam != nil {
		for _, v := range param.InstanceTextParam {
			key := *v.ParamName
			value := *v.CurrentValue
			instanceParams[key] = value
		}
	}
	params = instanceParams
	return
}

func (me *RedisService) DescribeRedisSslById(ctx context.Context, instanceId string) (ssl *redis.DescribeSSLStatusResponseParams, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDescribeSSLStatusRequest()
	request.InstanceId = &instanceId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseRedisClient().DescribeSSLStatus(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	ssl = response.Response
	return
}

func (me *RedisService) DescribeRedisMaintenanceWindowById(ctx context.Context, instanceId string) (maintenanceWindow *redis.DescribeMaintenanceWindowResponseParams, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDescribeMaintenanceWindowRequest()
	request.InstanceId = &instanceId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseRedisClient().DescribeMaintenanceWindow(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	maintenanceWindow = response.Response
	return
}

func (me *RedisService) DescribeRedisBackupDownloadInfoByFilter(ctx context.Context, param map[string]interface{}) (backup []*redis.BackupDownloadInfo, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = redis.NewDescribeBackupUrlRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "instance_id" {
			request.InstanceId = v.(*string)
		}
		if k == "backup_id" {
			request.BackupId = v.(*string)
		}
		if k == "limit_type" {
			request.LimitType = v.(*string)
		}
		if k == "vpc_comparison_symbol" {
			request.VpcComparisonSymbol = v.(*string)
		}
		if k == "ip_comparison_symbol" {
			request.IpComparisonSymbol = v.(*string)
		}
		if k == "limit_vpc" {
			request.LimitVpc = v.([]*redis.BackupLimitVpcItem)
		}
		if k == "limit_ip" {
			request.LimitIp = v.([]*string)
		}
	}

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseRedisClient().DescribeBackupUrl(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	backup = response.Response.BackupInfos

	return
}

func (me *RedisService) DescribeRedisBackupByFilter(ctx context.Context, param map[string]interface{}) (backup []*redis.RedisBackupSet, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = redis.NewDescribeInstanceBackupsRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "instance_id" {
			request.InstanceId = v.(*string)
		}
		if k == "begin_time" {
			request.BeginTime = v.(*string)
		}
		if k == "end_time" {
			request.EndTime = v.(*string)
		}
		if k == "status" {
			request.Status = v.([]*int64)
		}
		if k == "instance_name" {
			request.InstanceName = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())

	var (
		offset int64 = 0
		limit  int64 = 20
	)
	for {
		request.Offset = &offset
		request.Limit = &limit
		response, err := me.client.UseRedisClient().DescribeInstanceBackups(request)
		if err != nil {
			errRet = err
			return
		}
		log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

		if response == nil || len(response.Response.BackupSet) < 1 {
			break
		}
		backup = append(backup, response.Response.BackupSet...)
		if len(response.Response.BackupSet) < int(limit) {
			break
		}

		offset += limit
	}

	return
}

func (me *RedisService) DescribeRedisParamRecordsByFilter(ctx context.Context, param map[string]interface{}) (params []*redis.InstanceParamHistory, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = redis.NewDescribeInstanceParamRecordsRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "InstanceId" {
			request.InstanceId = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())

	var (
		offset    uint64 = 0
		limit     uint64 = 20
		instances        = make([]*redis.InstanceParamHistory, 0)
	)
	for {
		request.Offset = &offset
		request.Limit = &limit
		response, err := me.client.UseRedisClient().DescribeInstanceParamRecords(request)
		if err != nil {
			errRet = err
			return
		}
		log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

		if response == nil || len(response.Response.InstanceParamHistory) < 1 {
			break
		}
		instances = append(instances, response.Response.InstanceParamHistory...)
		if len(response.Response.InstanceParamHistory) < int(limit) {
			break
		}

		offset += limit
	}
	if len(instances) < 1 {
		return
	}
	params = instances

	return
}

func (me *RedisService) DescribeRedisInstanceShardsByFilter(ctx context.Context, param map[string]interface{}) (instanceShards []*redis.InstanceClusterShard, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = redis.NewDescribeInstanceShardsRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "InstanceId" {
			request.InstanceId = v.(*string)
		}
		if k == "FilterSlave" {
			request.FilterSlave = v.(*bool)
		}
	}

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeInstanceShards(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || len(response.Response.InstanceShards) < 1 {
		return
	}
	instanceShards = response.Response.InstanceShards

	return
}

func (me *RedisService) DescribeRedisInstanceZoneInfoByFilter(ctx context.Context, param map[string]interface{}) (instanceZoneInfo []*redis.ReplicaGroup, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = redis.NewDescribeInstanceZoneInfoRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "InstanceId" {
			request.InstanceId = v.(*string)
		}
	}

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseRedisClient().DescribeInstanceZoneInfo(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if response == nil || len(response.Response.ReplicaGroups) < 1 {
		return
	}
	instanceZoneInfo = response.Response.ReplicaGroups

	return
}

func (me *RedisService) DescribeRedisInstanceTaskListByFilter(ctx context.Context, param map[string]interface{}) (instanceTaskList []*redis.TaskInfoDetail, errRet error) {
	var (
		logId   = getLogId(ctx)
		request = redis.NewDescribeTaskListRequest()
	)

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	for k, v := range param {
		if k == "InstanceId" {
			request.InstanceId = v.(*string)
		}
		if k == "InstanceName" {
			request.InstanceName = v.(*string)
		}
		if k == "ProjectIds" {
			request.ProjectIds = v.([]*int64)
		}
		if k == "TaskTypes" {
			request.TaskTypes = v.([]*string)
		}
		if k == "BeginTime" {
			request.BeginTime = v.(*string)
		}
		if k == "EndTime" {
			request.EndTime = v.(*string)
		}
		if k == "TaskStatus" {
			request.TaskStatus = v.([]*int64)
		}
		if k == "Result" {
			request.Result = v.([]*int64)
		}
		if k == "OperateUin" {
			request.OperateUin = v.([]*string)
		}
	}

	ratelimit.Check(request.GetAction())

	var (
		offset int64 = 0
		limit  int64 = 20
	)
	for {
		request.Offset = &offset
		request.Limit = &limit
		response, err := me.client.UseRedisClient().DescribeTaskList(request)
		if err != nil {
			errRet = err
			return
		}
		log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

		if response == nil || len(response.Response.Tasks) < 1 {
			break
		}
		instanceTaskList = append(instanceTaskList, response.Response.Tasks...)
		if len(response.Response.Tasks) < int(limit) {
			break
		}

		offset += limit
	}

	return
}

func (me *RedisService) DescribeRedisReplicateInstanceById(ctx context.Context, instanceId string, groupId string) (replicateInstance *redis.Instances, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDescribeReplicationGroupRequest()
	request.GroupId = &groupId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	var (
		offset int64 = 0
		limit  int64 = 20
	)
	for {
		request.Offset = &offset
		request.Limit = &limit
		response, err := me.client.UseRedisClient().DescribeReplicationGroup(request)
		if err != nil {
			errRet = err
			return
		}
		log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

		if response == nil || len(response.Response.Groups) < 1 {
			break
		}
		for _, v := range response.Response.Groups {
			for _, instance := range v.Instances {
				if *instance.InstanceId == instanceId {
					replicateInstance = instance
					return
				}
			}
		}
		if len(response.Response.Groups) < int(limit) {
			break
		}

		offset += limit
	}

	return
}

func (me *RedisService) DeleteRedisReplicateInstanceById(ctx context.Context, instanceId string, groupId string) (taskId int64, errRet error) {
	logId := getLogId(ctx)

	request := redis.NewDeleteReplicationInstanceRequest()
	request.InstanceId = &instanceId
	request.GroupId = &groupId

	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n", logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	ratelimit.Check(request.GetAction())

	response, err := me.client.UseRedisClient().DeleteReplicationInstance(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	taskId = int64(*response.Response.TaskId)
	return
}
