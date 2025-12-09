package repository

import (
	// "database/sql"
	// "errors"
	"context"
	"errors"
	"fmt"
	apitypes "lab4-swag/internal/app/api_types"

	minio "lab4-swag/internal/app/minioClient"

	"mime/multipart"

	"lab4-swag/internal/app/ds"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w:  планета  с id %d", ErrNotFound, id)
		}
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

func (r *Repository) CreateCity(cityJSON apitypes.CityJSON) (ds.City, error) {
	city := apitypes.CityFromJSON(cityJSON)
	if city.Coordinates_long <= 0 {
		return ds.City{}, errors.New("неправильная долгота")
	}
	if city.Coordinates_width <= 0 {
		return ds.City{}, errors.New("нерпавильная широта")
	}
	err := r.db.Create(&city).First(&city).Error
	if err != nil {
		return ds.City{}, err
	}
	return city, nil
}

func (r *Repository) ChangeCity(id int, cityJSON apitypes.CityJSON) (ds.City, error) {
	city := ds.City{}
	if id < 0 {
		return ds.City{}, errors.New("id должно быть >= 0")
	}
	err := r.db.Where("id = ? and is_delete = ?", id, false).First(&city).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.City{}, fmt.Errorf("%w: планета с id %d", ErrNotFound, id)
		}
		return ds.City{}, err
	}
	if cityJSON.Coordinates_long <= 0 || cityJSON.Coordinates_width <= 0 {
		return ds.City{}, errors.New("нерпавильнo заданы координаты")
	}
	err = r.db.Model(&city).Updates(apitypes.CityFromJSON(cityJSON)).Error
	if err != nil {
		return ds.City{}, err
	}
	return city, nil
}

func (r *Repository) DeleteCity(id int) error {
	city := ds.City{}
	if id < 0 {
		return errors.New("id должно быть >= 0")
	}

	err := r.db.Where("id = ? and is_delete = ?", id, false).First(&city).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: город с id %d", ErrNotFound, id)
		}
		return err
	}
	if city.Image != "" {
		err = minio.DeleteObject(context.Background(), r.mc, minio.GetImgBucket(), city.Image)
		if err != nil {
			return err
		}
	}

	err = r.db.Model(&ds.City{}).Where("id = ?", id).Update("is_delete", true).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddCityToService(serviceId int, cityId int) error {
	var city ds.City
	if err := r.db.First(&city, cityId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: planet with id %d", ErrNotFound, cityId)
		}
		return err
	}

	var service ds.ServiceCalc
	if err := r.db.First(&service, serviceId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: заявка с id %d", ErrNotFound, serviceId)
		}
		return err
	}

	citesService := ds.CityForService{}
	result := r.db.Where("city_id = ? and service_calc_id = ?", cityId, serviceId).Find(&citesService)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 0 {
		return fmt.Errorf("%w: город %d уже в заявке %d", ErrAlreadyExists, cityId, serviceId)
	}
	
	// Если это центральный город, проверяем что его еще нет в заявке
	if city.Central {
		var existingCentral ds.CityForService
		result := r.db.Where("service_calc_id = ? and central = ?", serviceId, true).Find(&existingCentral)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		if result.RowsAffected != 0 {
			return fmt.Errorf("%w: в заявке %d уже есть центральный город", ErrAlreadyExists, serviceId)
		}
	}
	
	return r.db.Create(&ds.CityForService{
		CityID:        uint(cityId),
		ServiceCalcID: uint(serviceId),
		Central:       city.Central,
	}).Error
}

func (r *Repository) GetModeratorAndCreatorLogin(service ds.ServiceCalc) (string, string, error) {
	var creator ds.User
	var moderator ds.User

	err := r.db.Where("id = ?", service.CreatorID).First(&creator).Error
	if err != nil {
		return "", "", err
	}

	var moderatorLogin string
	if service.ModeratorID.Valid {
		err = r.db.Where("id = ?", service.ModeratorID).First(&moderator).Error
		if err != nil {
			return "", "", err
		}
		moderatorLogin = moderator.Login
	}

	return creator.Login, moderatorLogin, nil
}

func (r *Repository) UploadImage(ctx *gin.Context, cityId int, file *multipart.FileHeader) (ds.City, error) {
	city_, err := r.GetCity(cityId)
	if err != nil {
		return ds.City{}, err
	}

	fileName, err := minio.UploadImage(ctx, r.mc, minio.GetImgBucket(), file, *city_)
	if err != nil {
		return ds.City{}, err
	}

	city, err := r.GetCity(cityId)
	if err != nil {
		return ds.City{}, err
	}
	city.Image = fileName
	err = r.db.Save(&city).Error
	if err != nil {
		return ds.City{}, err
	}
	return *city, nil
}
