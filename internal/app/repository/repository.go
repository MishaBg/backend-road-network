package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type City struct { // вот наша новая структура
	ID          int    // поля структур, которые передаются в шаблон
	Title       string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
	Image       string
	Coordinates string
	Central     bool
	Request     bool
}

func (r *Repository) GetProtocols() ([]City, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	protocols := []City{ // массив элементов из наших структур
		{
			ID:          1,
			Title:       "Москва",
			Image:       "http://127.0.0.1:9000/web/moscow.jpeg",
			Coordinates: "55.75, 37.62",
			Central:     true,
			Request:     false,
		},
		{
			ID:          2,
			Title:       "Санкт-Петербург",
			Image:       "http://127.0.0.1:9000/web/piter.webp",
			Coordinates: "55.75, 37.62",
			Central:     false,
			Request:     false,
		},
		{
			ID:          3,
			Title:       "Сочи",
			Image:       "http://127.0.0.1:9000/web/sochi.webp",
			Coordinates: "55.75, 37.62",
			Central:     false,
			Request:     false,
		},
		{
			ID:          4,
			Title:       "Астрахань",
			Image:       "http://127.0.0.1:9000/web/astrax.jpg",
			Coordinates: "55.75, 37.62",
			Central:     false,
			Request:     false,
		},
		{
			ID:          5,
			Title:       "Краснодар",
			Image:       "http://127.0.0.1:9000/web/krasn.jpeg",
			Coordinates: "55.75, 37.62",
			Central:     false,
			Request:     false,
		},
		{
			ID:          6,
			Title:       "Казань",
			Image:       "http://127.0.0.1:9000/web/kazan.jpeg",
			Coordinates: "55.75, 37.62",
			Central:     false,
			Request:     false,
		},

		{
			ID:          7,
			Title:       "Иваново",
			Image:       "http://127.0.0.1:9000/web/ivanovo.webp",
			Coordinates: "55.75, 37.62",
			Central:     false,
			Request:     false,
		},
	}
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if len(protocols) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return protocols, nil
}

func (r *Repository) GetCity(id int) (City, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
	cites, err := r.GetProtocols()
	if err != nil {
		return City{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, protocol := range cites {
		if protocol.ID == id {
			return protocol, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return City{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
}

func (r *Repository) GetCitesByTitle(title string) ([]City, error) {
	cites, err := r.GetProtocols()
	if err != nil {
		return []City{}, err
	}

	var result []City
	for _, city := range cites {
		if strings.Contains(strings.ToLower(city.Title), strings.ToLower(title)) {
			result = append(result, city)
		}
	}

	return result, nil
}

func (r *Repository) GetRequestCites(id int) []City {
	reqCites := map[int][]int{1: {1, 3, 5}}
	var protocolsInGroup []City
	for _, protocolID := range reqCites[id] {
		protocol, err := r.GetCity(protocolID)
		if err == nil {
			protocolsInGroup = append(protocolsInGroup, protocol)
		}
	}
	return protocolsInGroup
}

func (r *Repository) GetRequestsCount(id int) int {
	return len(r.GetRequestCites(id))
}

func (r *Repository) GetRequestId() int {
	return 1
}
