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

package repository

import (
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/release"

	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	helmAction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const _indexFileName = "index.yaml"

var (
	ErrNotFound       = errors.New("not found")
	ErrNoValidURL     = errors.New("no valid url")
	ErrNoChartInfoSet = errors.New("no chart info set in installer")
)

type Driver string

func (d Driver) String() string {
	return string(d)
}

var _ Repository = &HelmRepo{}

const (
	Secret    Driver = "secret"
	ConfigMap Driver = "configmap"
	Mem       Driver = "memory"
	SQL       Driver = "sql"
)

const (
	_tkeelRepo          = "https://tkeel-io.github.io/helm-charts"
	_componentChartName = "tkeel-plugin-components"
)

type HelmRepo struct {
	info         *Info
	actionConfig *helmAction.Configuration
	httpGetter   getter.Getter
	driver       Driver
	namespace    string
}

func NewHelmRepo(info Info, driver Driver, namespace string) (*HelmRepo, error) {
	httpGetter, err := getter.NewHTTPGetter()
	if err != nil {
		log.Warn("init helm action configuration err", err)
		return nil, err
	}
	repo := &HelmRepo{
		info:       &info,
		namespace:  namespace,
		driver:     driver,
		httpGetter: httpGetter,
	}
	if err = repo.setActionConfig(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *HelmRepo) SetInfo(info Info) {
	r.info = &info
}

func (r *HelmRepo) Namespace() string {
	return r.namespace
}

func (r *HelmRepo) SetNamespace(namespace string) error {
	r.namespace = namespace
	return r.setActionConfig()
}

func (r *HelmRepo) SetDriver(driver Driver) error {
	r.driver = driver
	return r.setActionConfig()
}

func (r HelmRepo) GetDriver() Driver {
	return r.driver
}

func (r *HelmRepo) setActionConfig() error {
	config, err := initActionConfig(r.namespace, r.driver)
	if err != nil {
		log.Warn("init helm action configuration err", err)
		return err
	}
	r.actionConfig = config
	return nil
}

func (r *HelmRepo) Info() *Info {
	return r.info
}

func (r *HelmRepo) Search(word string) ([]*InstallerBrief, error) {
	index, err := r.BuildIndex()
	if err != nil {
		return nil, errors.Wrap(err, "can't build helm index config")
	}

	res := index.Search(word, "")
	briefs := res.ToInstallerBrief()

	// modify briefs installed status
	// 1. get this repo installed
	// 2. range briefs and change status.
	installedList, err := r.getInstalled()
	if err != nil {
		return nil, err
	}

	installedMap := make(map[string]struct{}, len(installedList))
	for i := range installedList {
		installedMap[installedList[i].Brief().Name] = struct{}{}
	}
	for i := 0; i < len(briefs); i++ {
		if _, ok := installedMap[briefs[i].Name]; ok {
			briefs[i].Installed = true
		}
	}

	return briefs, nil
}

func (r *HelmRepo) Get(name, version string) (Installer, error) {
	index, err := r.BuildIndex()
	if err != nil {
		return nil, errors.Wrap(err, "can't build helm index config")
	}

	res := index.Search(name, version)
	if len(res) != 1 {
		return nil, ErrNotFound
	}

	if len(res[0].URLs) == 0 {
		return nil, ErrNoValidURL
	}

	buf, err := r.httpGetter.Get(res[0].URLs[0])
	if err != nil {
		return nil, errors.Wrap(err, "GET target file failed")
	}

	ch, err := loader.LoadArchive(buf)
	if err != nil {
		return nil, errors.Wrap(err, "Load archive to struct Chart failed")
	}

	brief := res[0].ToInstallerBrief()
	i := NewHelmInstaller(brief.Name, ch, *brief, r.namespace, r.actionConfig)
	return &i, nil
}

func (r *HelmRepo) Installed() ([]Installer, error) {
	return r.getInstalled()
}

func (r *HelmRepo) Close() error {
	// TODO implement me
	panic("implement me")
}

func (r *HelmRepo) BuildIndex() (*Index, error) {
	fileContent, err := r.GetIndex()
	if err != nil {
		return nil, err
	}
	return NewIndex(r.info.Name, fileContent)
}

func (r *HelmRepo) GetIndex() ([]byte, error) {
	url := strings.TrimSuffix(r.info.URL, "/")
	url += "/" + _indexFileName

	buf, err := r.httpGetter.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP GET error")
	}

	return buf.Bytes(), nil
}

func (r *HelmRepo) list() ([]*release.Release, error) {
	listAction := helmAction.NewList(r.actionConfig)
	releases, err := listAction.Run()
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func (r *HelmRepo) getInstalled() ([]Installer, error) {
	index, err := r.BuildIndex()
	if err != nil {
		return nil, err
	}

	res := index.Search("*", "")

	rls, err := r.list()
	if err != nil {
		return nil, err
	}

	// TODO: Fix O(n²)
	list := make([]Installer, 0)
	for i := 0; i < len(rls); i++ {
		for j := 0; j < len(res); j++ {
			if rls[i].Chart.Name() == res[j].Name {
				installer := NewHelmInstaller(
					rls[i].Name,                /* Installed Plugin ID. */
					rls[i].Chart,               /* Plugin Chart. */
					*res[j].ToInstallerBrief(), /* Brief. */
					r.namespace,                /* Namespace. */
					r.actionConfig,             /* Action Config. */
				)
				list = append(list, &installer)
			}
		}
	}

	return list, nil
}

func getDebugLogFunc() helmAction.DebugLog {
	return func(format string, v ...interface{}) {
		log.Infof(format, v...)
	}
}

func initActionConfig(namespace string, driver Driver) (*helmAction.Configuration, error) {
	config := new(helmAction.Configuration)
	k8sFlags := &genericclioptions.ConfigFlags{
		Namespace: &namespace,
	}
	err := config.Init(k8sFlags, namespace, driver.String(), getDebugLogFunc())
	if err != nil {
		return nil, fmt.Errorf("helmAction configuration init err:%w", err)
	}
	return config, nil
}
