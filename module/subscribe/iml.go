package subscribe

import (
	"context"
	"errors"
	"fmt"
	"github.com/eolinker/apipark/service/partition"
	"github.com/eolinker/eosc/log"
	"gorm.io/gorm"

	"github.com/eolinker/apipark/gateway"

	"github.com/eolinker/apipark/service/cluster"

	"github.com/eolinker/go-common/utils"

	"github.com/eolinker/apipark/service/service"

	"github.com/eolinker/go-common/store"

	"github.com/google/uuid"

	"github.com/eolinker/go-common/auto"

	"github.com/eolinker/apipark/service/subscribe"

	"github.com/eolinker/apipark/service/project"

	subscribe_dto "github.com/eolinker/apipark/module/subscribe/dto"
)

var (
	_ ISubscribeModule = (*imlSubscribeModule)(nil)
)

type imlSubscribeModule struct {
	partitionService        partition.IPartitionService       `autowired:""`
	projectService          project.IProjectService           `autowired:""`
	projectPartitionService project.IProjectPartitionsService `autowired:""`
	subscribeService        subscribe.ISubscribeService       `autowired:""`
	subscribeApplyService   subscribe.ISubscribeApplyService  `autowired:""`
	serviceService          service.IServiceService           `autowired:""`
	clusterService          cluster.IClusterService           `autowired:""`
	transaction             store.ITransaction                `autowired:""`
}

func (i *imlSubscribeModule) PartitionServices(ctx context.Context, app string) ([]*subscribe_dto.PartitionServiceItem, error) {
	pInfo, err := i.projectService.Get(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("get application error: %w", err)
	}
	if !pInfo.AsApp {
		return nil, fmt.Errorf("project %s is not an application", app)
	}
	partitions, err := i.partitionService.List(ctx)
	if err != nil {
		return nil, err
	}
	subscriptions, err := i.subscribeService.SubscriptionsByApplication(ctx, app)
	if err != nil {
		return nil, err
	}
	subscriptionCount := make(map[string]int64)
	for _, s := range subscriptions {
		if s.ApplyStatus != subscribe.ApplyStatusSubscribe && s.ApplyStatus != subscribe.ApplyStatusReview {
			continue
		}
		if _, ok := subscriptionCount[s.Partition]; !ok {
			subscriptionCount[s.Partition] = 0
		}
		subscriptionCount[s.Partition]++
	}
	items := make([]*subscribe_dto.PartitionServiceItem, 0)
	for _, p := range partitions {
		items = append(items, &subscribe_dto.PartitionServiceItem{
			Id:         p.UUID,
			Name:       p.Name,
			ServiceNum: subscriptionCount[p.UUID],
		})
	}
	return items, nil
}

func (i *imlSubscribeModule) getSubscribers(ctx context.Context, partitionId string, projectIds []string) ([]*gateway.SubscribeRelease, error) {
	subscribers, err := i.subscribeService.SubscribersByProject(ctx, partitionId, projectIds...)
	if err != nil {
		return nil, err
	}
	return utils.SliceToSlice(subscribers, func(s *subscribe.Subscribe) *gateway.SubscribeRelease {
		return &gateway.SubscribeRelease{
			Service:     s.Service,
			Application: s.Application,
			Expired:     "0",
		}
	}), nil
}

func (i *imlSubscribeModule) initGateway(ctx context.Context, partitionId string, clientDriver gateway.IClientDriver) error {
	projectPartitions, err := i.projectPartitionService.ListByPartition(ctx, partitionId)
	if err != nil {
		return err
	}
	projectIds := utils.SliceToSlice(projectPartitions, func(p *project.Partition) string {
		return p.Project
	})
	releases, err := i.getSubscribers(ctx, partitionId, projectIds)
	if err != nil {
		return err
	}

	return clientDriver.Subscribe().Online(ctx, releases...)
}

func (i *imlSubscribeModule) SearchSubscriptions(ctx context.Context, partitionId string, app string, keyword string) ([]*subscribe_dto.SubscriptionItem, error) {
	pInfo, err := i.projectService.Get(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("get application error: %w", err)
	}
	if !pInfo.AsApp {
		return nil, fmt.Errorf("project %s is not an application", app)
	}

	// 获取当前订阅服务列表
	subscriptions, err := i.subscribeService.MySubscribeServices(ctx, app, nil, nil, partitionId)
	if err != nil {
		return nil, err
	}
	serviceIds := utils.SliceToSlice(subscriptions, func(s *subscribe.Subscribe) string {
		return s.Service
	})
	services, err := i.serviceService.SearchByUuids(ctx, keyword, serviceIds...)
	if err != nil {
		return nil, fmt.Errorf("search service error: %w", err)
	}
	serviceMap := utils.SliceToMapArray(services, func(s *service.Service) string {
		return s.Id
	})

	return utils.SliceToSlice(subscriptions, func(s *subscribe.Subscribe) *subscribe_dto.SubscriptionItem {
		return &subscribe_dto.SubscriptionItem{
			Id:          s.Id,
			Service:     auto.UUID(s.Service),
			Partition:   auto.UUID(s.Partition),
			ApplyStatus: s.ApplyStatus,
			Project:     auto.UUID(s.Project),
			Team:        auto.UUID(pInfo.Team),
			From:        s.From,
			CreateTime:  auto.TimeLabel(s.CreateAt),
		}
	}, func(s *subscribe.Subscribe) bool {
		_, ok := serviceMap[s.Service]
		if !ok {
			return false
		}
		if s.ApplyStatus != subscribe.ApplyStatusSubscribe && s.ApplyStatus != subscribe.ApplyStatusReview {
			return false
		}
		return true
	}), nil

}

func (i *imlSubscribeModule) RevokeSubscription(ctx context.Context, pid string, uuid string) error {
	_, err := i.projectService.Get(ctx, pid)
	if err != nil {
		return fmt.Errorf("get project error: %w", err)
	}
	subscription, err := i.subscribeService.Get(ctx, uuid)
	if err != nil {
		return err
	}
	if subscription.ApplyStatus != subscribe.ApplyStatusSubscribe {
		return fmt.Errorf("subscription can not be revoked")
	}
	applyStatus := subscribe.ApplyStatusUnsubscribe
	return i.transaction.Transaction(ctx, func(ctx context.Context) error {
		err = i.subscribeService.Save(ctx, uuid, &subscribe.UpdateSubscribe{
			ApplyStatus: &applyStatus,
		})
		if err != nil {
			return err
		}

		err = i.offlineForCluster(ctx, subscription.Partition, &gateway.SubscribeRelease{
			Service:     subscription.Service,
			Application: subscription.Application,
		})
		if err != nil {
			log.Warnf("revoke Subscription for partition:%s %s", subscription.Partition, err)
		}

		return nil
	})

}

func (i *imlSubscribeModule) DeleteSubscription(ctx context.Context, pid string, uuid string) error {
	_, err := i.projectService.Get(ctx, pid)
	if err != nil {
		return fmt.Errorf("get project error: %w", err)
	}
	subscription, err := i.subscribeService.Get(ctx, uuid)
	if err != nil {
		return err
	}
	if subscription.ApplyStatus == subscribe.ApplyStatusSubscribe || subscription.ApplyStatus == subscribe.ApplyStatusReview {
		return fmt.Errorf("subscription can not be deleted")
	}
	return i.subscribeService.Delete(ctx, uuid)
}

func (i *imlSubscribeModule) RevokeApply(ctx context.Context, app string, uuid string) error {
	_, err := i.projectService.Get(ctx, app)
	if err != nil {
		return fmt.Errorf("get app error: %w", err)
	}
	subscription, err := i.subscribeService.Get(ctx, uuid)
	if err != nil {
		return err
	}
	if subscription.ApplyStatus != subscribe.ApplyStatusReview {
		return fmt.Errorf("apply can not be revoked")
	}
	applyStatus := subscribe.ApplyStatusCancel
	return i.subscribeService.Save(ctx, uuid, &subscribe.UpdateSubscribe{
		ApplyStatus: &applyStatus,
	})
}

func (i *imlSubscribeModule) AddSubscriber(ctx context.Context, project string, input *subscribe_dto.AddSubscriber) error {
	_, err := i.projectService.Get(ctx, project)
	if err != nil {
		return err
	}
	if len(input.Partition) == 0 {
		return fmt.Errorf("partition is empty")
	}
	if input.Uuid == "" {
		input.Uuid = uuid.New().String()
	}
	sub := &gateway.SubscribeRelease{
		Service:     input.Service,
		Application: input.Project,
		Expired:     "0",
	}
	return i.transaction.Transaction(ctx, func(ctx context.Context) error {
		for _, partitionId := range input.Partition {
			err = i.subscribeService.Create(ctx, &subscribe.CreateSubscribe{
				Uuid:        input.Uuid,
				Service:     input.Service,
				Project:     project,
				Partition:   partitionId,
				Application: input.Project,
				ApplyStatus: subscribe.ApplyStatusSubscribe,
				From:        subscribe.FromUser,
			})
			if err != nil {
				return err
			}
			err := i.onlineSubscriber(ctx, partitionId, sub)
			if err != nil {
				return fmt.Errorf("add subscriber for partition[%s] %v", partitionId, err)
			}

		}

		return nil
	})

}

func (i *imlSubscribeModule) onlineSubscriber(ctx context.Context, partitionId string, subscriber *gateway.SubscribeRelease) error {
	info, err := i.partitionService.Get(ctx, partitionId)
	if err != nil {
		return err
	}

	client, err := i.clusterService.GatewayClient(ctx, info.Cluster)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close(ctx)
	}()
	return client.Subscribe().Online(ctx, subscriber)

}

func (i *imlSubscribeModule) DeleteSubscriber(ctx context.Context, project string, serviceId string, applicationId string) error {
	_, err := i.projectService.Get(ctx, project)
	if err != nil {
		return err
	}

	return i.transaction.Transaction(ctx, func(ctx context.Context) error {
		list, err := i.subscribeService.ListByApplication(ctx, serviceId, applicationId)
		if err != nil {
			return err
		}
		releaseInfo := &gateway.SubscribeRelease{
			Service:     serviceId,
			Application: applicationId,
		}
		for _, s := range list {
			err = i.subscribeService.Delete(ctx, s.Id)
			if err != nil {
				return err
			}
			err := i.offlineForCluster(ctx, s.Partition, releaseInfo)
			if err != nil {
				return fmt.Errorf("offline subscribe for partition[%s] %s", s.Partition, err)
			}
		}
		return nil
	})
}
func (i *imlSubscribeModule) offlineForCluster(ctx context.Context, partitionId string, config *gateway.SubscribeRelease) error {
	info, err := i.partitionService.Get(ctx, partitionId)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return nil
	}
	client, err := i.clusterService.GatewayClient(ctx, info.Cluster)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close(ctx)
	}()
	return client.Subscribe().Offline(ctx, config)
}

func (i *imlSubscribeModule) SearchSubscribers(ctx context.Context, projectId string, keyword string) ([]*subscribe_dto.Subscriber, error) {
	pInfo, err := i.projectService.Get(ctx, projectId)
	if err != nil {
		return nil, err
	}

	// 获取当前项目所有订阅方
	list, err := i.subscribeService.ListBySubscribeStatus(ctx, projectId, subscribe.ApplyStatusSubscribe)
	if err != nil {
		return nil, err
	}
	subscriberMap := utils.SliceToMapArrayO(list, func(s *subscribe.Subscribe) (string, string) {
		return fmt.Sprintf("%s-%s", s.Service, s.Application), s.Partition
	})

	if keyword == "" {
		items := make([]*subscribe_dto.Subscriber, 0, len(list))
		for _, subscriber := range list {
			key := fmt.Sprintf("%s-%s", subscriber.Service, subscriber.Application)
			partitionIds, ok := subscriberMap[key]
			if !ok {
				continue
			}

			items = append(items, &subscribe_dto.Subscriber{
				Id:         subscriber.Application,
				Project:    auto.UUID(subscriber.Project),
				Service:    auto.UUID(subscriber.Service),
				Partition:  auto.List(partitionIds),
				Subscriber: auto.UUID(subscriber.Application),
				Team:       auto.UUID(pInfo.Team),
				ApplyTime:  auto.TimeLabel(subscriber.CreateAt),
				From:       subscriber.From,
			})
			delete(subscriberMap, key)
		}
		return items, nil
	}
	serviceList, err := i.serviceService.Search(ctx, keyword, map[string]interface{}{
		"project": projectId,
	})
	if err != nil {
		return nil, err
	}
	serviceMap := utils.SliceToMap(serviceList, func(s *service.Service) string {
		return s.Id
	})
	items := make([]*subscribe_dto.Subscriber, 0, len(list))
	for _, subscriber := range list {
		key := fmt.Sprintf("%s-%s", subscriber.Service, subscriber.Application)
		partitionIds, ok := subscriberMap[key]
		if !ok {
			continue
		}
		if _, ok := serviceMap[subscriber.Service]; ok {
			items = append(items, &subscribe_dto.Subscriber{
				Id:         subscriber.Id,
				Project:    auto.UUID(subscriber.Project),
				Service:    auto.UUID(subscriber.Service),
				Partition:  auto.List(partitionIds),
				Subscriber: auto.UUID(subscriber.Application),
				Team:       auto.UUID(pInfo.Team),
				ApplyTime:  auto.TimeLabel(subscriber.CreateAt),
				From:       subscriber.From,
			})
			delete(subscriberMap, key)
		}
	}
	return items, nil
}

var _ ISubscribeApprovalModule = (*imlSubscribeApprovalModule)(nil)

type imlSubscribeApprovalModule struct {
	subscribeService        subscribe.ISubscribeService       `autowired:""`
	subscribeApplyService   subscribe.ISubscribeApplyService  `autowired:""`
	projectService          project.IProjectService           `autowired:""`
	projectPartitionService project.IProjectPartitionsService `autowired:""`
	clusterService          cluster.IClusterService           `autowired:""`
	transaction             store.ITransaction                `autowired:""`
}

func (i *imlSubscribeApprovalModule) Pass(ctx context.Context, pid string, id string, approveInfo *subscribe_dto.Approve) error {
	applyInfo, err := i.subscribeApplyService.Get(ctx, id)
	if err != nil {
		return err
	}
	partitions, err := i.projectPartitionService.GetByProject(ctx, pid)
	if err != nil {
		return err
	}
	partitionMap := utils.SliceToMapO(partitions, func(s string) (string, struct{}) {
		return s, struct{}{}
	})
	for _, pt := range approveInfo.Partition {
		if _, ok := partitionMap[pt]; !ok {
			return fmt.Errorf("partition %s not exists", pt)
		}
	}

	return i.transaction.Transaction(ctx, func(ctx context.Context) error {
		userID := utils.UserId(ctx)
		status := subscribe.ApplyStatusSubscribe
		err = i.subscribeApplyService.Save(ctx, id, &subscribe.EditApply{
			ApplyPartitions: approveInfo.Partition,
			Opinion:         &approveInfo.Opinion,
			Status:          &status,
			Approver:        &userID,
		})
		if err != nil {
			return err
		}
		err = i.subscribeService.UpdateSubscribeStatus(ctx, applyInfo.Application, applyInfo.Service, status)
		if err != nil {
			return err
		}
		cs, err := i.clusterService.List(ctx, approveInfo.Partition...)
		if err != nil {
			return err
		}
		for _, c := range cs {

			err := i.onlineSubscriber(ctx, c.Uuid, &gateway.SubscribeRelease{
				Service:     applyInfo.Service,
				Application: applyInfo.Application,
				Expired:     "0",
			})

			if err != nil {
				log.Warnf("online subscriber for partition[%s] %v", c.Partition, err)

			}
		}
		return nil
	})
}
func (i *imlSubscribeApprovalModule) onlineSubscriber(ctx context.Context, clusterId string, sub *gateway.SubscribeRelease) error {
	client, err := i.clusterService.GatewayClient(ctx, clusterId)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close(ctx)
	}()
	return client.Subscribe().Online(ctx, sub)
}
func (i *imlSubscribeApprovalModule) Reject(ctx context.Context, pid string, id string, approveInfo *subscribe_dto.Approve) error {
	applyInfo, err := i.subscribeApplyService.Get(ctx, id)
	if err != nil {
		return err
	}

	return i.transaction.Transaction(ctx, func(ctx context.Context) error {
		userID := utils.UserId(ctx)
		status := subscribe.ApplyStatusRefuse
		err = i.subscribeApplyService.Save(ctx, id, &subscribe.EditApply{
			ApplyPartitions: approveInfo.Partition,
			Opinion:         &approveInfo.Opinion,
			Status:          &status,
			Approver:        &userID,
		})
		if err != nil {
			return err
		}
		return i.subscribeService.UpdateSubscribeStatus(ctx, applyInfo.Application, applyInfo.Service, status)
	})
}

func (i *imlSubscribeApprovalModule) GetApprovalList(ctx context.Context, pid string, status int) ([]*subscribe_dto.ApprovalItem, error) {
	applyStatus := make([]int, 0, 2)
	if status == 0 {
		// 获取待审批列表
		applyStatus = append(applyStatus, subscribe.ApplyStatusReview)
	} else {
		// 获取已审批列表
		applyStatus = append(applyStatus, subscribe.ApplyStatusRefuse, subscribe.ApplyStatusSubscribe)
	}
	items, err := i.subscribeApplyService.ListByStatus(ctx, pid, applyStatus...)
	if err != nil {
		return nil, err
	}
	return utils.SliceToSlice(items, func(s *subscribe.Apply) *subscribe_dto.ApprovalItem {
		return &subscribe_dto.ApprovalItem{
			Id:           s.Id,
			Service:      auto.UUID(s.Service),
			Project:      auto.UUID(s.Project),
			Team:         auto.UUID(s.Team),
			ApplyProject: auto.UUID(s.Application),
			ApplyTeam:    auto.UUID(s.ApplyTeam),
			ApplyTime:    auto.TimeLabel(s.ApplyAt),
			Applier:      auto.UUID(s.Applier),
			Approver:     auto.UUID(s.Approver),
			ApprovalTime: auto.TimeLabel(s.ApproveAt),
			Status:       s.Status,
		}
	}), nil
}

func (i *imlSubscribeApprovalModule) GetApprovalDetail(ctx context.Context, pid string, id string) (*subscribe_dto.Approval, error) {
	_, err := i.projectService.Get(ctx, pid)
	if err != nil {
		return nil, err
	}
	item, err := i.subscribeApplyService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &subscribe_dto.Approval{
		Id:           item.Id,
		Service:      auto.UUID(item.Service),
		Project:      auto.UUID(item.Project),
		Team:         auto.UUID(item.Team),
		ApplyProject: auto.UUID(item.Application),
		ApplyTeam:    auto.UUID(item.ApplyTeam),
		ApplyTime:    auto.TimeLabel(item.ApplyAt),
		Applier:      auto.UUID(item.Applier),
		Approver:     auto.UUID(item.Approver),
		ApprovalTime: auto.TimeLabel(item.ApproveAt),
		Partition:    auto.List(item.ApplyPartitions),
		Reason:       item.Reason,
		Opinion:      item.Opinion,
		Status:       item.Status,
	}, nil
}
