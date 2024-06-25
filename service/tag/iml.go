package tag

import (
	"context"
	"time"

	"github.com/eolinker/go-common/auto"
	"github.com/eolinker/go-common/utils"

	"github.com/eolinker/apipark/stores/tag"

	"github.com/eolinker/apipark/service/universally"
)

var (
	_ ITagService = (*imlTagService)(nil)
)

type imlTagService struct {
	store tag.ITagStore `autowired:""`
	universally.IServiceGet[Tag]
	universally.IServiceDelete
	universally.IServiceCreate[CreateTag]
}

func (i *imlTagService) GetLabels(ctx context.Context, ids ...string) map[string]string {
	if len(ids) == 0 {
		return nil
	}
	list, err := i.store.ListQuery(ctx, "`uuid` in (?)", []interface{}{ids}, "id")
	if err != nil {
		return nil
	}
	return utils.SliceToMapO(list, func(i *tag.Tag) (string, string) {
		return i.UUID, i.Name
	})
}

func (i *imlTagService) OnComplete() {
	i.IServiceGet = universally.NewGet[Tag, tag.Tag](i.store, FromEntity)
	i.IServiceCreate = universally.NewCreator[CreateTag, tag.Tag](i.store, "catalogue", createEntityHandler, uniquestHandler, labelHandler)
	i.IServiceDelete = universally.NewDelete[tag.Tag](i.store)
	auto.RegisterService("tag", i)
}

func labelHandler(e *tag.Tag) []string {
	return []string{e.Name, e.UUID}
}
func uniquestHandler(i *CreateTag) []map[string]interface{} {
	return []map[string]interface{}{{"uuid": i.Id}}
}
func createEntityHandler(i *CreateTag) *tag.Tag {
	return &tag.Tag{
		UUID:     i.Id,
		Name:     i.Name,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}
}
