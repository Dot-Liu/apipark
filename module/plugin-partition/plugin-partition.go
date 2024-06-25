package plugin_partition

import (
	"context"
	"github.com/eolinker/apipark/gateway"
	"github.com/eolinker/apipark/model/plugin_model"
	"github.com/eolinker/apipark/module/plugin-partition/dto"
	"github.com/eolinker/go-common/autowire"
	"reflect"
)

type IPluginPartitionModule interface {
	List(ctx context.Context, partition string) ([]*dto.Item, error)
	Get(ctx context.Context, partition string, name string) (config *dto.PluginOutput, render plugin_model.Render, er error)
	Set(ctx context.Context, partition string, name string, config *dto.PluginSetting) error
	Options(ctx context.Context) ([]*dto.PluginOption, error)
	GetDefine(ctx context.Context, name string) (*dto.Define, error)
	UpdateDefine(ctx context.Context, defines []*plugin_model.Define) error
}

func init() {
	autowire.Auto[IPluginPartitionModule](func() reflect.Value {
		m := new(imlPluginPartitionModule)
		gateway.RegisterInitHandleFunc(m.initGateway)
		return reflect.ValueOf(m)
	})
}
