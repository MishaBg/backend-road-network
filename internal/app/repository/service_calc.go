package repository

import (
	"errors"
	"lab1-design-backend/internal/app/ds"
	"time"

	"github.com/sirupsen/logrus"
	// "time"
	// "github.com/sirupsen/logrus"
)

var errNoDraft = errors.New("no draft for this user")

func (r *Repository) GetCitesService(id int) ([]ds.CityInfo, ds.ServiceCalc, error) {

	creatorID := r.GetUser()
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	var service ds.ServiceCalc
	err := r.db.Where("id = ?", id).First(&service).Error
	if err != nil {
		return []ds.CityInfo{}, ds.ServiceCalc{}, err
	} else if creatorID != int(service.CreatorID) {
		return []ds.CityInfo{}, ds.ServiceCalc{}, errors.New("you are not allowed")
	} else if service.Status == "deleted" {
		return []ds.CityInfo{}, ds.ServiceCalc{}, errors.New("you can`t watch deleted calculations")
	}

	var cites []ds.City
	var citesServices []ds.CityForService
	sub := r.db.Table("city_for_services").Where("service_calc_id = ?", service.ID).Find(&citesServices)
	err = r.db.Where("id IN (?)", sub.Select("city_id")).Find(&cites).Error
	if err != nil {
		return []ds.CityInfo{}, ds.ServiceCalc{}, err
	}

	var citesResult []ds.CityInfo
	for _, city := range cites {
		for _, citesService := range citesServices {
			if city.ID == int(citesService.CityID) {
				citesResult = append(citesResult, ds.CityInfo{
					ID:      city.ID,
					Name:    city.Name,
					Image:   city.Image,
					Central: city.Central,

					Coordinates_long:  city.Coordinates_long,
					Coordinates_width: city.Coordinates_width,
				})
				break
			}
		}
	}

	return citesResult, service, nil
}

func (r *Repository) CheckCurrentServiceDraft(creatorID int) (ds.ServiceCalc, error) {
	var service ds.ServiceCalc

	res := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").Limit(1).Find(&service)
	if res.Error != nil {
		return ds.ServiceCalc{}, res.Error
	} else if res.RowsAffected == 0 {
		return ds.ServiceCalc{}, errNoDraft
	}
	return service, nil
}

func (r *Repository) GetServiceDraft(creatorID int) (ds.ServiceCalc, error) {
	service, err := r.CheckCurrentServiceDraft(creatorID)
	if err == errNoDraft {
		service = ds.ServiceCalc{
			Status:     "черновик",
			CreatorID:  creatorID,
			DateCreate: time.Now(),
		}
		result := r.db.Create(&service)
		if result.Error != nil {
			return ds.ServiceCalc{}, result.Error
		}
		return service, nil
	} else if err != nil {
		return ds.ServiceCalc{}, err
	}
	return service, nil
}

func (r *Repository) GetServiceCount() int64 {
	var serviceID uint
	var count int64
	creatorID := 1
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	err := r.db.Model(&ds.ServiceCalc{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("id").First(&serviceID).Error
	if err != nil {
		return 0
	}
	err = r.db.Model(&ds.CityForService{}).Where("service_calc_id = ?", serviceID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_planets:", err)
	}

	return count
}

func (r *Repository) DeleteCalculation(serviceId int) error {
	return r.db.Exec("UPDATE service_calcs SET status = 'deleted' WHERE id = ?", serviceId).Error
}
