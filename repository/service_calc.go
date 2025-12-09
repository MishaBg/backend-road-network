package repository

import (
	"database/sql"
	"errors"
	"fmt"
	apitypes "lab4-swag/internal/app/api_types"
	"lab4-swag/internal/app/ds"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var errNoDraft = errors.New("no draft for this user")

func (r *Repository) GetServices(from, to time.Time, status string) ([]ds.ServiceCalc, error) {
	var services []ds.ServiceCalc
	sub := r.db.Where("status != 'deleted'")
	if !from.IsZero() {
		sub = sub.Where("date_create > ?", from)
	}
	if !to.IsZero() {
		sub = sub.Where("date_create < ?", to.Add(time.Hour*24))
	}
	if status != "" {
		sub = sub.Where("status = ?", status)
	}
	err := sub.Order("id").Find(&services).Error
	if err != nil {
		return nil, err
	}
	return services, nil
}

func (r *Repository) GetCitesServices(serviceId int) ([]ds.CityForService, error) {
	var citesServices []ds.CityForService
	err := r.db.Where("service_calc_id = ?", serviceId).Find(&citesServices).Error
	if err != nil {
		return nil, err
	}
	return citesServices, nil
}

func (r *Repository) GetCitesService(cityId int, serviceId int) (ds.CityForService, error) {
	var citesService ds.CityForService
	err := r.db.Where("city_id = ? and service_calc_id = ?", cityId, serviceId).First(&citesService).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.CityForService{}, fmt.Errorf("%w: cites service not found", ErrNotFound)
		}
		return ds.CityForService{}, err
	}
	return citesService, nil
}

func (r *Repository) GetServiceCites(id int) ([]ds.City, ds.ServiceCalc, error) {
	service, err := r.GetSingleService(id)
	if err != nil {
		return []ds.City{}, ds.ServiceCalc{}, err
	}

	var cites []ds.City
	sub := r.db.Table("city_for_services").Where("service_calc_id = ?", service.ID)
	err = r.db.Order("id DESC").Where("id IN (?)", sub.Select("city_id")).Find(&cites).Error

	if err != nil {
		return []ds.City{}, ds.ServiceCalc{}, err
	}

	return cites, service, nil
}

func (r *Repository) CheckCurrentServiceDraft(creatorID uuid.UUID) (ds.ServiceCalc, error) {
	// if creatorID == 0 {
	//     return ds.Research{}, fmt.Errorf("%w: user not authenticated", ErrNotAllowed)
	// }

	var service ds.ServiceCalc
	res := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").Limit(1).Find(&service)
	if res.Error != nil {
		return ds.ServiceCalc{}, res.Error
	} else if res.RowsAffected == 0 {
		return ds.ServiceCalc{}, ErrNoDraft
	}
	return service, nil
}

func (r *Repository) GetServiceDraft(creatorID uuid.UUID) (ds.ServiceCalc, bool, error) {
	// if creatorID == 0 {
	//     return ds.Research{}, false, fmt.Errorf("%w: user not authenticated", ErrNotAllowed)
	// }

	service, err := r.CheckCurrentServiceDraft(creatorID)
	if errors.Is(err, ErrNoDraft) {
		service = ds.ServiceCalc{
			Status:     "черновик",
			CreatorID:  creatorID,
			DateCreate: time.Now(),
		}
		result := r.db.Create(&service)
		if result.Error != nil {
			return ds.ServiceCalc{}, false, result.Error
		}
		return service, true, nil
	} else if err != nil {
		return ds.ServiceCalc{}, false, err
	}
	return service, true, nil
}

func (r *Repository) GetServiceCount(creatorID uuid.UUID) int64 {
	var count int64
	service, err := r.CheckCurrentServiceDraft(creatorID)
	if err != nil {
		return 0
	}
	err = r.db.Model(&ds.CityForService{}).Where("service_calc_id = ?", service.ID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_cites:", err)
	}

	return count
}

func (r *Repository) DeleteCalculation(serviceId int) error {
	return r.db.Exec("UPDATE service_calcs SET status = 'deleted' WHERE id = ?", serviceId).Error
}

func (r *Repository) GetSingleService(id int) (ds.ServiceCalc, error) {
	if id < 0 {
		return ds.ServiceCalc{}, errors.New("неверное id, должно быть >= 0")
	}

	// userId := r.GetUserID()
	// if userId == 0 {
	//     return ds.Research{}, fmt.Errorf("%w: пользователь не авторизирован", ErrNotAllowed)
	// }

	// user, err := r.GetUserByID(userId)
	// if err != nil {
	//  return ds.Research{}, err
	// }

	var service ds.ServiceCalc
	err := r.db.Where("id = ?", id).First(&service).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.ServiceCalc{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, id)
		}
		return ds.ServiceCalc{}, err
	} else if service.Status == "deleted" {
		return ds.ServiceCalc{}, fmt.Errorf("%w: заявка удалена", ErrNotAllowed)
	}
	return service, nil
}

func (r *Repository) FormService(serviceId int, status string) (ds.ServiceCalc, error) {
	service, err := r.GetSingleService(serviceId)
	if err != nil {
		return ds.ServiceCalc{}, err
	}

	// user, err := r.GetUserByID(r.GetUserID())
	// if err != nil{
	//  return ds.Research{}, fmt.Errorf("%w: пользователь на авторизирован", ErrNotAllowed)
	// }

	// if research.CreatorID != r.userId && !user.IsModerator{
	//  return ds.Research{}, fmt.Errorf("%w: у вас нет прав чтобы эта заявка имела статус %s", ErrNotAllowed, status)
	// }

	if service.Status != "черновик" {
		return ds.ServiceCalc{}, fmt.Errorf("эта заявка не может быть %s", status)
	}

	if status != "deleted" {
		if service.DateResearch.IsZero() {
			return ds.ServiceCalc{}, errors.New("вы не написали дату исследования")
		}
		citesService, _ := r.GetCitesServices(int(service.ID))
		for _, cityResearch := range citesService {
			if cityResearch.CityID == 0 {
				return ds.ServiceCalc{}, errors.New("вы не написали id города")
			}
		}
	}

	err = r.db.Model(&service).Updates(ds.ServiceCalc{
		Status: status,
		DateForm: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}).Error
	if err != nil {
		return ds.ServiceCalc{}, err
	}

	return service, nil
}

func (r *Repository) ChangeService(id int, serviceJSON apitypes.ServiceJSON) (ds.ServiceCalc, error) {
	service := ds.ServiceCalc{}
	if id < 0 {
		return ds.ServiceCalc{}, errors.New("неправильное id, должно быть >= 0")
	}
	if serviceJSON.DateResearch == "" {
		return ds.ServiceCalc{}, errors.New("неправильная дата исследования")
	}
	err := r.db.Where("id = ? and status != 'deleted'", id).First(&service).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.ServiceCalc{}, fmt.Errorf("%w: исследование с id %d", ErrNotFound, id)
		}
		return ds.ServiceCalc{}, err
	}
	err = r.db.Model(&service).Updates(apitypes.ServiceFromJSON(serviceJSON)).Error
	if err != nil {
		return ds.ServiceCalc{}, err
	}
	return service, nil
}

func CalculateCityCost(coordinates_long float64, dateResearch string, coordinates_width float64) (float64, error) {
	if dateResearch == "" {
		return 0, errors.New("неправильная дата исследования")
	}
	return float64(coordinates_long) * math.Sqrt(float64(coordinates_width/100)), nil
}

func (r *Repository) ModerateService(id int, status string, currUserId uuid.UUID) (ds.ServiceCalc, error) {
	if status != "completed" && status != "rejected" {
		return ds.ServiceCalc{}, errors.New("неверный статус")
	}

	var result_cost float64

	service, err := r.GetSingleService(id)
	if err != nil {
		return ds.ServiceCalc{}, err
	} else if service.Status != "formed" {
		return ds.ServiceCalc{}, fmt.Errorf("это исследование не может быть %s", status)
	}

	err = r.db.Model(&service).Updates(ds.ServiceCalc{
		Status: status,
		DateFinish: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		ModeratorID: uuid.NullUUID{
			UUID:  currUserId,
			Valid: true,
		},
	}).Error
	if err != nil {
		return ds.ServiceCalc{}, err
	}

	if status == "completed" {
		citesService, err := r.GetCitesServices(int(service.ID))
		if err != nil {
			return ds.ServiceCalc{}, err
		}
		for _, cityService := range citesService {
			city, err := r.GetCity(int(cityService.CityID))
			if err != nil {
				return ds.ServiceCalc{}, err
			}
			cityCost, err := CalculateCityCost(city.Coordinates_long, service.DateResearch.Format("2006-01-02"), city.Coordinates_width)
			result_cost += cityCost
			if err != nil {
				return ds.ServiceCalc{}, err
			}
		}
		result := r.db.Model(&ds.ServiceCalc{}).Where("id = ?", service.ID).Update("cost", result_cost)
		err = result.Error
		if err != nil {
			return ds.ServiceCalc{}, err
		}
	}

	return service, nil
}

// SetServiceCost обновляет поле cost в таблице service_calcs
func (r *Repository) SetServiceCost(serviceId int, cost float64) error {
	result := r.db.Model(&ds.ServiceCalc{}).Where("id = ?", serviceId).Update("cost", cost)
	return result.Error
}

func DistanceBetweenCities(lat1, lon1, lat2, lon2 float64) float64 {
	// Конвертируем градусы в радианы
	radLat1 := lat1 * math.Pi / 180
	radLat2 := lat2 * math.Pi / 180
	radLon1 := lon1 * math.Pi / 180
	radLon2 := lon2 * math.Pi / 180

	// Разница координат
	deltaLat := radLat2 - radLat1
	deltaLon := radLon2 - radLon1

	// Формула гаверсинусов
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Радиус Земли в километрах
	radius := 6371.0

	return radius * c
}
