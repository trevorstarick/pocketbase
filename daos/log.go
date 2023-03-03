package daos

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/models"
)

func (dao *Dao) LogQuery() *dbx.SelectQuery {
	return dao.ModelQuery(&models.Log{})
}

func (dao *Dao) FindLogById(id string) (*models.Log, error) {
	model := &models.Log{}

	err := dao.LogQuery().
		AndWhere(dbx.HashExp{"id": id}).
		Limit(1).
		One(model)

	if err != nil {
		return nil, err
	}

	return model, nil
}
