package daos

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/models"
)

func (dao *Dao) ErrorQuery() *dbx.SelectQuery {
	return dao.ModelQuery(&models.Error{})
}

func (dao *Dao) FindErrorById(id string) (*models.Error, error) {
	model := &models.Error{}

	err := dao.ErrorQuery().
		AndWhere(dbx.HashExp{"id": id}).
		Limit(1).
		One(model)

	if err != nil {
		return nil, err
	}

	return model, nil
}
