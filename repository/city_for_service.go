package repository

import (
	"errors"
	"fmt"
	apitypes "lab4-swag/internal/app/api_types"
	"lab4-swag/internal/app/ds"

	"gorm.io/gorm"
)

func (r *Repository) DeleteCityFromService(serviceId int, cityId int) (ds.ServiceCalc, error) {
	// userId := r.userId
	// if userId == 0 {
	//     return ds.Research{}, fmt.Errorf("%w: пользователь не авторизирован", ErrNotAllowed)
	// }

	// user, err := r.GetUserByID(userId)
	// if err != nil {
	// 	return ds.Research{}, err
	// }

	var service ds.ServiceCalc
	err := r.db.Where("id = ?", serviceId).First(&service).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.ServiceCalc{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, serviceId)
		}
		return ds.ServiceCalc{}, err
	}

	// if research.CreatorID != r.userId && !user.IsModerator{
	// 	return ds.Research{}, fmt.Errorf("%w: Вы не создатель этого исследования", ErrNotAllowed)
	// }

	err = r.db.Where("city_id = ? and service_calc_id = ?", cityId, serviceId).Delete(&ds.CityForService{}).Error
	if err != nil {
		return ds.ServiceCalc{}, err
	}
	return service, nil
}

func (r *Repository) ChangeСityService(serviceId int, cityId int, citesServiceJSON apitypes.CitesServiceJSON) (ds.CityForService, error) {
	var citesService ds.CityForService

	// Если устанавливаем central=true, проверяем что другого центрального города нет
	if citesServiceJSON.Central {
		var existingCentral ds.CityForService
		result := r.db.Where("service_calc_id = ? and central = ? and city_id != ?", serviceId, true, cityId).Find(&existingCentral)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ds.CityForService{}, result.Error
		}
		if result.RowsAffected != 0 {
			return ds.CityForService{}, fmt.Errorf("%w: в заявке уже есть другой центральный город", ErrAlreadyExists)
		}
	}

	err := r.db.Model(&citesService).Where("city_id = ? and service_calc_id = ?", cityId, serviceId).Updates(apitypes.CitesServiceFromJSON(citesServiceJSON)).First(&citesService).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.CityForService{}, fmt.Errorf("%w: планеты в исследовании", ErrNotFound)
		}
		return ds.CityForService{}, err
	}
	return citesService, nil
}
