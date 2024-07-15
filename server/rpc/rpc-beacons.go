package rpc

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
*/

import (
	"context"

	"github.com/gsmith257-cyber/better-sliver-package/protobuf/clientpb"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/commonpb"
	"github.com/gsmith257-cyber/better-sliver-package/server/db"
	"github.com/gsmith257-cyber/better-sliver-package/server/db/models"
	"github.com/gsmith257-cyber/better-sliver-package/server/log"
)

var (
	baconRpcLog = log.NamedLogger("rpc", "bacons")
)

// GetBacons - Get a list of bacons from the database
func (rpc *Server) GetBacons(ctx context.Context, req *commonpb.Empty) (*clientpb.Bacons, error) {
	bacons, err := db.ListBacons()
	if err != nil {
		baconRpcLog.Errorf("Failed to find db bacons: %s", err)
		return nil, ErrDatabaseFailure
	}
	for id, bacon := range bacons {
		all, completed, err := db.CountTasksByBaconID(bacon.ID)
		if err != nil {
			baconRpcLog.Errorf("Task count failed: %s", err)
		}
		bacons[id].TasksCount = all
		bacons[id].TasksCountCompleted = completed
	}
	return &clientpb.Bacons{Bacons: bacons}, nil
}

// GetBacon - Get a list of bacons from the database
func (rpc *Server) GetBacon(ctx context.Context, req *clientpb.Bacon) (*clientpb.Bacon, error) {
	bacon, err := db.BaconByID(req.ID)
	if err != nil {
		baconRpcLog.Error(err)
		return nil, ErrDatabaseFailure
	}
	return bacon.ToProtobuf(), nil
}

// RmBacon - Delete a bacon and any related tasks
func (rpc *Server) RmBacon(ctx context.Context, req *clientpb.Bacon) (*commonpb.Empty, error) {
	bacon, err := db.BaconByID(req.ID)
	if err != nil {
		baconRpcLog.Error(err)
		return nil, ErrInvalidBaconID
	}

	err = db.Session().Where(&models.BaconTask{
		BaconID: bacon.ID},
	).Delete(&models.BaconTask{}).Error
	if err != nil {
		baconRpcLog.Errorf("Database error: %s", err)
		return nil, ErrDatabaseFailure
	}
	err = db.Session().Delete(bacon).Error
	if err != nil {
		baconRpcLog.Errorf("Database error: %s", err)
		return nil, ErrDatabaseFailure
	}
	return &commonpb.Empty{}, nil
}

// GetBaconTasks - Get a list of tasks for a specific bacon
func (rpc *Server) GetBaconTasks(ctx context.Context, req *clientpb.Bacon) (*clientpb.BaconTasks, error) {
	bacon, err := db.BaconByID(req.ID)
	if err != nil {
		return nil, ErrInvalidBaconID
	}
	tasks, err := db.BaconTasksByBaconID(bacon.ID.String())
	return &clientpb.BaconTasks{Tasks: tasks}, err
}

// GetBaconTaskContent - Get the content of a specific task
func (rpc *Server) GetBaconTaskContent(ctx context.Context, req *clientpb.BaconTask) (*clientpb.BaconTask, error) {
	task, err := db.BaconTaskByID(req.ID)
	if err != nil {
		return nil, ErrInvalidBaconTaskID
	}
	return task, nil
}

// CancelBaconTask - Cancel a bacon task
func (rpc *Server) CancelBaconTask(ctx context.Context, req *clientpb.BaconTask) (*clientpb.BaconTask, error) {
	task, err := db.BaconTaskByID(req.ID)
	if err != nil {
		return nil, ErrInvalidBaconTaskID
	}
	if task.State == models.PENDING {
		task.State = models.CANCELED
		err = db.Session().Save(task).Error
		if err != nil {
			baconRpcLog.Errorf("Database error: %s", err)
			return nil, ErrDatabaseFailure
		}
	} else {
		// No real point to cancel the task if it's already been sent
		return task, ErrInvalidBaconTaskCancelState
	}
	task, err = db.BaconTaskByID(req.ID)
	if err != nil {
		return nil, ErrInvalidBaconTaskID
	}
	return task, nil
}

// UpdateBaconIntegrityInformation - Update process integrity information for a bacon
func (rpc *Server) UpdateBaconIntegrityInformation(ctx context.Context, req *clientpb.BaconIntegrity) (*commonpb.Empty, error) {
	resp := &commonpb.Empty{}
	bacon, err := db.BaconByID(req.BaconID)
	if err != nil || bacon == nil {
		return resp, ErrInvalidBaconID
	}
	bacon.Integrity = req.Integrity
	err = db.Session().Save(bacon).Error
	if err != nil {
		return resp, err
	}
	return resp, nil
}
