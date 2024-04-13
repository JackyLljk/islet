package logic

import (
	"islet/dao/mysql"
	"islet/models"
)

func GetCommunityList() (list []*models.Community, err error) {
	return mysql.GetCommunityList()
}

func GetCommunityDetail(id string) (community *models.CommunityDetail, err error) {
	return mysql.GetCommunityDetail(id)
}
