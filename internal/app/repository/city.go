package repository

import (
	// "database/sql"
	// "errors"
	"fmt"

	"lab1-design-backend/internal/app/ds"
)

func (r *Repository) GetCites() ([]ds.City, error) {
	var cites []ds.City
	err := r.db.Order("id").Where("is_delete = false").Find(&cites).Error
	if err != nil {
		return nil, err
	}
	if len(cites) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return cites, nil
}

func (r *Repository) GetCity(id int) (*ds.City, error) {
	city := ds.City{}
	err := r.db.Order("id").Where("id = ? and is_delete = ?", id, false).First(&city).Error
	if err != nil {
		return &ds.City{}, err
	}
	return &city, nil
}

func (r *Repository) GetCitesByName(name string) ([]ds.City, error) {
	var cites []ds.City
	err := r.db.Order("id").Where("name ILIKE ? and is_delete = ?", "%"+name+"%", false).Find(&cites).Error
	if err != nil {
		return nil, err
	}
	return cites, nil
}

func (r *Repository) AddCityToService(serviceId int, cityId int) error {
	var city ds.City
	if err := r.db.First(&city, cityId).Error; err != nil {
		return err
	}

	var service ds.ServiceCalc
	if err := r.db.First(&service, serviceId).Error; err != nil {
		return err
	}
	citesService := ds.CityForService{}
	result := r.db.Where("city_id = ? and service_calc_id = ?", cityId, serviceId).Find(&citesService)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 0 {
		return nil
	}
	return r.db.Create(&ds.CityForService{
		CityID:        uint(cityId),
		ServiceCalcID: uint(serviceId),
	}).Error
}
