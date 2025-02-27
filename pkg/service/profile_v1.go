/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"encoding/json"

	pb "github.com/tkeel-io/tkeel/api/profile/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/plgprofile"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"

	"github.com/tkeel-io/kit/log"
)

type ProfileService struct {
	pb.UnimplementedProfileServer
	pluginOp  plugin.Operator
	ProfileOp plgprofile.ProfileOperator
}

func NewProfileService(plgOp plugin.Operator, profileOp plgprofile.ProfileOperator) *ProfileService {
	return &ProfileService{pluginOp: plgOp, ProfileOp: profileOp}
}

func (s *ProfileService) GetTenantProfile(ctx context.Context, req *pb.GetTenantProfileRequest) (*pb.GetTenantProfileResponse, error) {
	profiles, err := s.ProfileOp.GetTenantProfile(ctx, req.GetTenantId())
	if profiles == nil {
		if err != nil {
			log.Error(err)
		}
		plugins, err := s.pluginOp.List(ctx)
		if err != nil {
			log.Error("plugin operator list: ", err)
			return nil, pb.ErrPluginList()
		}
		return &pb.GetTenantProfileResponse{TenantProfiles: plugin2pbProfile(plugins)}, nil
	}
	plugins, err := s.pluginOp.List(ctx)
	profiles = comboProfiles(profiles, plugins)
	return &pb.GetTenantProfileResponse{TenantProfiles: modelProfile2pbProfile(profiles)}, nil
}

func (s *ProfileService) SetTenantPluginProfile(ctx context.Context, req *pb.SetTenantPluginProfileRequest) (*pb.SetTenantPluginProfileResponse, error) {
	if req.GetTenantId() == "" {
		return nil, pb.ErrInvalidArgument()
	}
	modelPluginProfile := pbPlgProfile2model(req.GetBody())
	err := s.ProfileOp.SetTenantPluginProfile(ctx, req.GetTenantId(), modelPluginProfile)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrUnknown()
	}

	if modelPluginProfile.PluginID == plgprofile.PLUGIN_ID_KEEL {
		for i, _ := range modelPluginProfile.Profiles {
			if modelPluginProfile.Profiles[i].Key == plgprofile.MAX_API_REQUEST_LIMIT_KEY {
				plgprofile.SetTenantAPILimit(req.GetTenantId(), int(modelPluginProfile.Profiles[i].Default.(float64)))
				break
			}
		}
	}

	return &pb.SetTenantPluginProfileResponse{}, nil
}

func (s *ProfileService) IsAPIRequestExceededLimit(ctx context.Context, tenantID string) bool {
	plgprofile.OnTenantAPIRequest(tenantID, s.ProfileOp)
	return plgprofile.ISExceededAPILimit(tenantID)
}

func plugin2pbProfile(plugins []*model.Plugin) []*pb.TenantProfiles {
	pbProfiles := make([]*pb.TenantProfiles, 0)
	for i := range plugins {
		if plugins[i].Profiles == nil {
			continue
		}
		profileBytes, _ := json.Marshal(plugins[i].Profiles)
		profile := &pb.TenantProfiles{PluginId: plugins[i].ID, Profiles: profileBytes}
		pbProfiles = append(pbProfiles, profile)
	}
	pbProfiles = append(pbProfiles, plgprofile.KeelProfiles)
	return pbProfiles
}

func modelProfile2pbProfile(profiles []*model.PluginProfile) []*pb.TenantProfiles {
	pbProfiles := make([]*pb.TenantProfiles, 0)
	for i := range profiles {
		if profiles[i].Profiles == nil {
			continue
		}
		profileBytes, _ := json.Marshal(profiles[i].Profiles)
		profile := &pb.TenantProfiles{PluginId: profiles[i].PluginID, Profiles: profileBytes}
		pbProfiles = append(pbProfiles, profile)
	}
	return pbProfiles
}
func pbPlgProfile2model(profiles *pb.TenantProfiles) *model.PluginProfile {
	profilesItems := []*model.ProfileItem{}
	json.Unmarshal(profiles.Profiles, &profilesItems)
	profile := &model.PluginProfile{PluginID: profiles.PluginId, Profiles: profilesItems}
	return profile
}
func comboProfiles(profiles []*model.PluginProfile, plugins []*model.Plugin) []*model.PluginProfile {
	newProfiles := []*model.PluginProfile{}
	for plugini, _ := range plugins {
		for profilesi, _ := range profiles {
			if plugins[plugini].ID == profiles[profilesi].PluginID {
				break
			}
		}
		newProfiles = append(newProfiles, &model.PluginProfile{PluginID: plugins[plugini].ID, Profiles: plugins[plugini].Profiles})
	}
	profiles = append(profiles, newProfiles...)
	return profiles
}
