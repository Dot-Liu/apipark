package project_authorization

import (
	"github.com/gin-gonic/gin"
	"reflect"

	"github.com/eolinker/go-common/autowire"

	project_authorization_dto "github.com/eolinker/apipark/module/project-authorization/dto"
)

type IProjectAuthorizationController interface {
	// AddAuthorization 添加项目鉴权信息
	AddAuthorization(ctx *gin.Context, pid string, info *project_authorization_dto.CreateAuthorization) (*project_authorization_dto.Authorization, error)
	// EditAuthorization 修改项目鉴权信息
	EditAuthorization(ctx *gin.Context, pid string, aid string, info *project_authorization_dto.EditAuthorization) (*project_authorization_dto.Authorization, error)
	// DeleteAuthorization 删除项目鉴权
	DeleteAuthorization(ctx *gin.Context, pid string, aid string) error
	// Authorizations 获取项目鉴权列表
	Authorizations(ctx *gin.Context, pid string) ([]*project_authorization_dto.AuthorizationItem, error)
	// Detail 获取项目鉴权详情（弹窗用）
	Detail(ctx *gin.Context, pid string, aid string) ([]project_authorization_dto.DetailItem, error)
	// Info 获取项目鉴权详情
	Info(ctx *gin.Context, pid string, aid string) (*project_authorization_dto.Authorization, error)
}

func init() {
	autowire.Auto[IProjectAuthorizationController](func() reflect.Value {
		return reflect.ValueOf(new(imlProjectAuthorizationController))
	})
}
