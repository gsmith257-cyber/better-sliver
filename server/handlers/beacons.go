package handlers

/*
	Sliver Implant Framework
	Copyright (C) 2021  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
	------------------------------------------------------------------------

	WARNING: These functions can be invoked by remote implants without user interaction

*/

import (
	"encoding/json"
	"errors"
	"time"

	consts "github.com/gsmith257-cyber/better-sliver-package/client/constants"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/clientpb"
	sliverpb "github.com/gsmith257-cyber/better-sliver-package/protobuf/sliverpb"
	"github.com/gsmith257-cyber/better-sliver-package/server/core"
	"github.com/gsmith257-cyber/better-sliver-package/server/db"
	"github.com/gsmith257-cyber/better-sliver-package/server/db/models"
	"github.com/gsmith257-cyber/better-sliver-package/server/log"
	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var (
	baconHandlerLog = log.NamedLogger("handlers", "bacons")
)

func baconRegisterHandler(implantConn *core.ImplantConnection, data []byte) *sliverpb.Envelope {
	baconReg := &sliverpb.BaconRegister{}
	err := proto.Unmarshal(data, baconReg)
	if err != nil {
		baconHandlerLog.Errorf("Error decoding bacon registration message: %s", err)
		return nil
	}
	baconHandlerLog.Infof("Bacon registration from %s", baconReg.ID)
	bacon, err := db.BaconByID(baconReg.ID)
	baconHandlerLog.Debugf("Found %v err = %s", bacon, err)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		baconHandlerLog.Errorf("Database query error %s", err)
		return nil
	}
	baconUUID, _ := uuid.FromString(baconReg.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		bacon = &models.Bacon{
			ID: baconUUID,
		}
	}
	baconRegUUID, _ := uuid.FromString(baconReg.Register.Uuid)
	bacon.Name = baconReg.Register.Name
	bacon.Hostname = baconReg.Register.Hostname
	bacon.UUID = baconRegUUID
	bacon.Username = baconReg.Register.Username
	bacon.UID = baconReg.Register.Uid
	bacon.GID = baconReg.Register.Gid
	bacon.OS = baconReg.Register.Os
	bacon.Arch = baconReg.Register.Arch
	bacon.Transport = implantConn.Transport
	bacon.RemoteAddress = implantConn.RemoteAddress
	bacon.PID = baconReg.Register.Pid
	bacon.Filename = baconReg.Register.Filename
	bacon.LastCheckin = implantConn.GetLastMessage()
	bacon.Version = baconReg.Register.Version
	bacon.ReconnectInterval = baconReg.Register.ReconnectInterval
	bacon.ActiveC2 = baconReg.Register.ActiveC2
	bacon.ProxyURL = baconReg.Register.ProxyURL
	// bacon.ConfigID = uuid.FromStringOrNil(baconReg.Register.ConfigID)
	bacon.Locale = baconReg.Register.Locale

	bacon.Interval = baconReg.Interval
	bacon.Jitter = baconReg.Jitter
	bacon.NextCheckin = time.Now().Unix() + baconReg.NextCheckin

	err = db.Session().Save(bacon).Error
	if err != nil {
		baconHandlerLog.Errorf("Database write %s", err)
	}

	eventData, _ := proto.Marshal(bacon.ToProtobuf())
	core.EventBroker.Publish(core.Event{
		EventType: consts.BaconRegisteredEvent,
		Data:      eventData,
		Bacon:    bacon,
	})

	go auditLogBacon(bacon, baconReg.Register)
	return nil
}

type auditLogNewBaconMsg struct {
	Bacon   *clientpb.Bacon
	Register *sliverpb.Register
}

func auditLogBacon(bacon *models.Bacon, register *sliverpb.Register) {
	msg, err := json.Marshal(auditLogNewBaconMsg{
		Bacon:   bacon.ToProtobuf(),
		Register: register,
	})
	if err != nil {
		baconHandlerLog.Errorf("Failed to log new bacon to audit log: %s", err)
	} else {
		log.AuditLogger.Warn(string(msg))
	}
}

func baconTasksHandler(implantConn *core.ImplantConnection, data []byte) *sliverpb.Envelope {
	BaconTasks := &sliverpb.BaconTasks{}
	err := proto.Unmarshal(data, BaconTasks)
	if err != nil {
		baconHandlerLog.Errorf("Error decoding bacon tasks message: %s", err)
		return nil
	}
	go func() {
		err := db.UpdateBaconCheckinByID(BaconTasks.ID, BaconTasks.NextCheckin)
		if err != nil {
			baconHandlerLog.Errorf("failed to update checkin: %s", err)
		}
	}()

	// If the message contains tasks then process it as results
	// otherwise send the bacon any pending tasks. Currently we
	// don't receive results and send pending tasks at the same
	// time. We only send pending tasks if the request is empty.
	// If we send the Bacon 0 tasks it should not respond at all.
	if 0 < len(BaconTasks.Tasks) {
		baconHandlerLog.Infof("Bacon %s returned %d task result(s)", BaconTasks.ID, len(BaconTasks.Tasks))
		go baconTaskResults(BaconTasks.ID, BaconTasks.Tasks)
		return nil
	}

	baconHandlerLog.Infof("Bacon %s requested pending task(s)", BaconTasks.ID)

	// Pending tasks are ordered by their creation time.
	pendingTasks, err := db.PendingBaconTasksByBaconID(BaconTasks.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		baconHandlerLog.Errorf("Bacon task database error: %s", err)
		return nil
	}
	tasks := []*sliverpb.Envelope{}
	for _, pendingTask := range pendingTasks {
		envelope := &sliverpb.Envelope{}
		err = proto.Unmarshal(pendingTask.Request, envelope)
		if err != nil {
			baconHandlerLog.Errorf("Error decoding pending task: %s", err)
			continue
		}
		envelope.ID = pendingTask.EnvelopeID
		tasks = append(tasks, envelope)
		pendingTask.State = models.SENT
		pendingTask.SentAt = time.Now().Unix()
		err = db.Session().Model(&models.BaconTask{}).Where(&models.BaconTask{
			ID: pendingTask.ID,
		}).Updates(pendingTask).Error
		if err != nil {
			baconHandlerLog.Errorf("Database error: %s", err)
		}
	}
	taskData, err := proto.Marshal(&sliverpb.BaconTasks{Tasks: tasks})
	if err != nil {
		baconHandlerLog.Errorf("Error marshaling bacon tasks message: %s", err)
		return nil
	}
	baconHandlerLog.Infof("Sending %d task(s) to bacon %s", len(pendingTasks), BaconTasks.ID)
	return &sliverpb.Envelope{
		Type: sliverpb.MsgBaconTasks,
		Data: taskData,
	}
}

func baconTaskResults(baconID string, taskEnvelopes []*sliverpb.Envelope) *sliverpb.Envelope {
	for _, envelope := range taskEnvelopes {
		dbTask, err := db.BaconTaskByEnvelopeID(baconID, envelope.ID)
		if err != nil {
			baconHandlerLog.Errorf("Error finding db task: %s", err)
			continue
		}
		if dbTask == nil {
			baconHandlerLog.Errorf("Error: nil db task!")
			continue
		}
		dbTask.State = models.COMPLETED
		dbTask.CompletedAt = time.Now().Unix()
		dbTask.Response = envelope.Data
		id, _ := uuid.FromString(dbTask.ID)
		err = db.Session().Model(&models.BaconTask{}).Where(&models.BaconTask{
			ID: id,
		}).Updates(dbTask).Error
		if err != nil {
			baconHandlerLog.Errorf("Error updating db task: %s", err)
			continue
		}
		eventData, _ := proto.Marshal(dbTask)
		core.EventBroker.Publish(core.Event{
			EventType: consts.BaconTaskResultEvent,
			Data:      eventData,
		})
	}
	return nil
}
