package vnet

import (
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/serializer/go/serialize/object"
	"github.com/saichler/types/go/common"
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
	go ns.sendNotification()
	return ns
}

func (this *NotificationSender) requestNotification() {
	this.cond.L.Lock()
	defer this.cond.L.Unlock()
	this.lastNotificationRequest = time.Now().UnixMilli()
	this.notificationRequestCount++
	this.cond.Broadcast()
}

func (this *NotificationSender) sendNotification() {
	for this.vnet.running {
		this.cond.L.Lock()
		sendNotification := false
		//If there are notification requests and either last sent is larger than a second
		//or last one was sent more than a second ago
		//mark to send a notification
		if this.notificationRequestCount > 0 &&
			(time.Now().UnixMilli()-this.lastNotificationRequest >= 1000 ||
				time.Now().UnixMilli()-this.lastNotificationSentTime >= 1000) {
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
			conns := this.vnet.switchTable.conns.all()

			syncData, _ := this.vnet.protocol.CreateMessageFor("", health.ServiceName, 0, common.P1,
				common.Sync, vnetUuid, vnetUuid, object.New(nil, nil), false, false,
				nextId, nil)
			for _, vnic := range conns {
				go func() {
					vnic.SendMessage(syncData)
				}()
			}
		}
		this.cond.L.Unlock()
	}
}
