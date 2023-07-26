// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"fmt"
	"memphis/db"
	"memphis/memphis_cache"
	"strconv"
	"time"
)

func (s *Server) CompleteRelevantStuckAsyncTasks() {
	exist, asyncTasks, err := db.GetAsyncTaskByNameAndBrokerName("resend_all_dls_msgs", s.opts.ServerName)
	if err != nil {
		serv.Errorf("CompleteRelevantStuckAsyncTasks: failed to get async tasks resend_all_dls_msgs: %v", err.Error())
		return
	}
	if !exist {
		return
	}

	for _, asyncTask := range asyncTasks {
		exist, station, err := db.GetStationById(asyncTask.StationId, asyncTask.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v]CompleteRelevantStuckAsyncTasks at GetStationById: %v", asyncTask.TenantName, err.Error())
			return
		}
		if !exist {
			stationIdStr := strconv.Itoa(asyncTask.StationId)
			errMsg := fmt.Sprintf("Station %v does not exist", stationIdStr)
			serv.Warnf("[tenant: %v][user: %v]CompleteRelevantStuckAsyncTasks at GetStationById: %s", asyncTask.TenantName, errMsg)
			continue
		}

		exist, user, err := memphis_cache.GetUser(station.CreatedByUsername, asyncTask.TenantName)
		if err != nil {
			serv.Errorf("[tenant:%v][user: %v] CompleteRelevantStuckAsyncTasks could not retrive user model from cache or db error: %v", asyncTask.TenantName, user.Username, err)
			continue
		}
		if !exist {
			serv.Warnf("[tenant:%v][user: %v] CompleteRelevantStuckAsyncTasks user does not exist", asyncTask.TenantName, user.Username)
			continue
		}

		s.ResendAllDlsMsgs(station.Name, station.ID, station.TenantName, user)
	}
}

func (s *Server) RemoveInactiveAsyncTasks() {
	duration := 20 * time.Minute
	stationIds, err := db.RemoveAllAsyncTasks(duration)
	if err != nil {
		serv.Errorf("RemoveInactiveAsyncTasks: failed to get async tasks resend_all_dls_msgs: %v", err.Error())
		return
	}

	if len(stationIds) == 0 {
		return
	}

	err = db.UpdateResendDisabledInStations(false, stationIds)
	if err != nil {
		serv.Errorf("RemoveInactiveAsyncTasks: failed to update resend disabled in station: %v", err.Error())
		return
	}
}