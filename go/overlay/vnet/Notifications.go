package vnet

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/types/go/common"
	"github.com/saichler/types/go/types"
	"sync"
	"time"
)

type NotificationSender struct {
	notificationRequestCount int
	cond                     *sync.Cond
	lastNotificationSentTime int64
	lastNotificationRequest  int64
	vnet                     *VNet
}

func newNotificationSender(vnet *VNet) *NotificationSender {
	ns := &NotificationSender{}
	ns.cond = sync.NewCond(&sync.Mutex{})
	ns.vnet = vnet
	go ns.processHealthServiceNotifications()
	return ns
}

func (this *NotificationSender) requestHealthServiceNotification() {
	this.cond.L.Lock()
	defer this.cond.L.Unlock()
	this.lastNotificationRequest = time.Now().UnixMilli()
	this.notificationRequestCount++
	this.cond.Broadcast()
}

func (this *NotificationSender) processHealthServiceNotifications() {
	for this.vnet.running {
		this.cond.L.Lock()
		sendNotification := false
		//If there are notification requests and either last sent is larger than a second
		//or last one was sent more than a second ago
		//mark to send a notification
		if this.notificationRequestCount > 0 &&
			(time.Now().UnixMilli()-this.lastNotificationRequest >= 1000 ||
				time.Now().UnixMilli()-this.lastNotificationSentTime >= 2000) {
			sendNotification = true
			this.lastNotificationSentTime = time.Now().UnixMilli()
			this.notificationRequestCount = 0
		}
		//if there are no request to send notification and no notification to be sent
		//wait
		if this.notificationRequestCount == 0 && !sendNotification {
			this.cond.Wait()
		}
		//if it is not the time to send a notification, sleep 100 milis
		if !sendNotification {
			time.Sleep(time.Millisecond * 100)
		}
		//Send the notification
		if sendNotification {
			vnetUuid := this.vnet.resources.SysConfig().LocalUuid
			nextId := this.vnet.protocol.NextMessageNumber()
			syncData, _ := this.vnet.protocol.CreateMessageFor("", health.ServiceName, 0, common.P1,
				common.Sync, vnetUuid, vnetUuid, object.New(nil, nil), false, false,
				nextId, nil)
			go this.vnet.HandleData(syncData, nil)
		}
		this.cond.L.Unlock()
	}
}

func (this *VNet) PropertyChangeNotification(set *types.NotificationSet) {
	//only health service will call this callback so check if the notification is from a local source
	//if it is from local source, then just notify local vnics
	protocol.AddPropertyChangeCalled(set, this.resources.SysConfig().LocalAlias)

	vnetUuid := this.resources.SysConfig().LocalUuid
	nextId := this.protocol.NextMessageNumber()
	syncData, _ := this.protocol.CreateMessageFor("", set.ServiceName, uint16(set.ServiceArea), common.P1,
		common.Notify, vnetUuid, vnetUuid, object.New(nil, set), false, false,
		nextId, nil)

	go this.HandleData(syncData, nil)
}
