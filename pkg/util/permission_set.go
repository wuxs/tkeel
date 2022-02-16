package util

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
)

func AddPluginPermissionOnSet(ctx context.Context, kv kv.Operator, pluginID string, ps []*v1.Permission) (RollBackStack, error) {
	rbStack := NewRollbackStack()
	old, err := model.GetPermissionSet().Marshall()
	if err != nil {
		return nil, errors.Wrap(err, "permission set marshal")
	}
	for _, p := range ps {
		_, err = model.GetPermissionSet().Add(pluginID, p)
		if err != nil {
			return nil, errors.Wrapf(err, "permission set add(%s/%s)", pluginID, p)
		}
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("add permission set roll back run")
		model.GetPermissionSet().Delete(pluginID)
		return nil
	})
	b, err := model.GetPermissionSet().Marshall()
	if err != nil {
		rbStack.Run()
		return nil, errors.Wrap(err, "permission set marshal")
	}
	if err = kv.Update(ctx, model.KeyPermissionSet, b, ""); err != nil {
		rbStack.Run()
		return nil, errors.Wrapf(err, "KV operator update(%s/%s)", model.KeyPermissionSet, b)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("kv add permission set roll back run")
		if err := kv.Delete(ctx, model.KeyPermissionSet); err != nil {
			return errors.Wrapf(err, "KV delete %s", model.KeyPermissionSet)
		}
		if err = kv.Create(ctx, model.KeyPermissionSet, old); err != nil {
			return errors.Wrapf(err, "KV create %s/%s", model.KeyPermissionSet, old)
		}
		return nil
	})
	return rbStack, nil
}

func DeletePluginPermissionOnSet(ctx context.Context, kv kv.Operator, pluginID string) (RollBackStack, error) {
	rbStack := NewRollbackStack()
	old, err := model.GetPermissionSet().Marshall()
	if err != nil {
		return nil, errors.Wrap(err, "permission set marshal")
	}
	model.GetPermissionSet().Delete(pluginID)
	b, err := model.GetPermissionSet().Marshall()
	if err != nil {
		return nil, errors.Wrap(err, "permission set marshal")
	}
	rbStack = append(rbStack, func() error {
		model.GetPermissionSet().Unmarshal(old)
		return nil
	})
	if err = kv.Update(ctx, model.KeyPermissionSet, b, ""); err != nil {
		rbStack.Run()
		return nil, errors.Wrapf(err, "KV operator update(%s/%s)", model.KeyPermissionSet, b)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("kv delete permission set roll back run")
		if err := kv.Delete(ctx, model.KeyPermissionSet); err != nil {
			return errors.Wrapf(err, "KV delete %s", model.KeyPermissionSet)
		}
		if err = kv.Create(ctx, model.KeyPermissionSet, old); err != nil {
			return errors.Wrapf(err, "KV create %s/%s", model.KeyPermissionSet, old)
		}
		return nil
	})
	return rbStack, nil
}

func GetPermissionAllDependencePath(p *v1.Permission) ([]string, error) {
	ret := make([]string, 0, len(p.Dependences))
	for _, v := range p.Dependences {
		p, err := model.GetPermissionSet().GetPermission(v.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "get permission by path(%s)", v.Path)
		}
		ps, err := GetPermissionAllDependencePath(p)
		if err != nil {
			return nil, errors.Wrapf(err, "get GetPermissionAllDependencePath(%s)", p.Name)
		}
		ret = append(ret, ps...)
	}
	return ret, nil
}

func GetPermissionPathSet(pathList []string) (map[string]struct{}, error) {
	addPmPathSet := make(map[string]struct{})
	for _, v := range pathList {
		pm, err := model.GetPermissionSet().GetPermission(v)
		if err != nil {
			if errors.Is(err, model.ErrPermissionNotExist) {
				return nil, model.ErrPermissionNotExist
			}
			return nil, errors.Wrapf(err, "check permission %v", pathList)
		}
		addPmPathSet[v] = struct{}{}
		ps, err := GetPermissionAllDependencePath(pm)
		if err != nil {
			return nil, errors.Wrapf(err, "get permission(%s) all dependence path", pm.Name)
		}
		for _, v := range ps {
			addPmPathSet[v] = struct{}{}
		}
	}
	return addPmPathSet, nil
}

func Set2List(set map[string]struct{}) []string {
	ret := make([]string, 0, len(set))
	for k := range set {
		ret = append(ret, k)
	}
	return ret
}
