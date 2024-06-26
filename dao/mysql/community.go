package mysql

import (
	"database/sql"
	"errors"
	"islet/models"

	"go.uber.org/zap"
)

func GetCommunityList() (list []*models.Community, err error) {
	sqlStr := "select community_id, community_name from community"
	err = db.Select(&list, sqlStr)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return
	}

	return
}

func GetCommunityDetail(id string) (community *models.CommunityDetail, err error) {
	community = new(models.CommunityDetail)
	sqlStr := `select community_id, community_name, introduction, create_time
				from community
				where community_id = ?`
	err = db.Get(community, sqlStr, id)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrorInvalidID
		return
	}
	if err != nil {
		zap.L().Error("query community failed", zap.String("sql", sqlStr), zap.Error(err))
		err = ErrorQueryFailed
		return
	}

	return
}
