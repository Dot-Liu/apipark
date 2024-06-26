package team

import (
	"time"

	"github.com/eolinker/apipark/stores/team"
)

type Team struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Master       string    `json:"master"`
	Organization string    `json:"organization"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`
	Creator      string    `json:"creator"`
	Updater      string    `json:"updater"`
	//ProjectNum   int64     `json:"project_num"`
}

func FromEntity(e *team.Team) *Team {
	return &Team{
		Id:           e.UUID,
		Name:         e.Name,
		Description:  e.Description,
		Master:       e.Master,
		Organization: e.Organization,
		CreateTime:   e.CreateAt,
		UpdateTime:   e.UpdateAt,
		Creator:      e.Creator,
		Updater:      e.Updater,
	}
}

type CreateTeam struct {
	Id           string `json:"id" `
	Name         string `json:"name" `
	Description  string `json:"description"`
	Master       string `json:"master" `
	Organization string `json:"organization"`
}
type EditTeam struct {
	Name         *string `json:"name" `
	Description  *string `json:"description"`
	Master       *string `json:"master" `
	Organization *string `json:"organization"`
}
