package types

import (
	"github.com/Astenna/Nubes/lib"
	"github.com/jftuga/geodist"
)

type City struct {
	CityName        string `nubes:"Id" dynamodbav:"Id"`
	Region          string
	Description     string
	Hotels          lib.ReferenceNavigationList[Hotel] `nubes:"hasOne-City" dynamodbav:"-"`
	isInitialized   bool
	invocationDepth int
}

func (o City) GetTypeName() string {
	return "City"
}

type CloseToParams struct {
	Longitude float64
	Latitude  float64
	Count     int
}

func (c City) GetHotelsCloseTo(param CloseToParams) ([]Hotel, error) {
	c.invocationDepth++
	if c.isInitialized && c.invocationDepth == 1 {
		_libError := lib.GetStub(c.CityName, &c)
		if _libError != nil {
			c.invocationDepth--
			return *new([]Hotel), _libError
		}
	}
	hotels, err := c.Hotels.GetStubs()
	result := make([]Hotel, 0, param.Count)

	if err != nil {
		c.invocationDepth--
		return nil, err
	}

	if len(hotels) <= param.Count {
		c.invocationDepth--
		return hotels, err
	}
	hotelDists := make([]hotelDist, 0, len(hotels))
	from := geodist.Coord{Lat: param.Latitude, Lon: param.Longitude}
	for i, h := range hotels {

		to := geodist.Coord{
			Lat: h.Coordinates.Lat,
			Lon: h.Coordinates.Lon,
		}

		_, km := geodist.HaversineDistance(from, to)
		hotelDists[i] = hotelDist{hotel: &h, distance: km}
	}

	quickSortHotelDist(hotelDists, 0, len(hotelDists))

	for i := 0; i < param.Count; i++ {
		result[i] = *hotelDists[i].hotel
	}
	_libUpsertError := c.saveChangesIfInitialized()
	c.invocationDepth--
	return result, _libUpsertError
}

func (c City) GetHotelsWithBestRates(count int) ([]Hotel, error) {
	c.invocationDepth++
	if c.isInitialized && c.invocationDepth == 1 {
		_libError := lib.GetStub(c.CityName, &c)
		if _libError != nil {
			c.invocationDepth--
			return *new([]Hotel), _libError
		}
	}
	hotels, err := c.Hotels.GetStubs()
	result := make([]Hotel, 0, count)

	if err != nil {
		c.invocationDepth--
		return nil, err
	}

	if len(result) <= count {
		c.invocationDepth--
		return result, err
	}

	quickSortRate(hotels, 0, len(hotels))

	for i := 0; i < count; i++ {
		result[i] = hotels[len(hotels)-i-1]
	}
	_libUpsertError := c.saveChangesIfInitialized()
	c.invocationDepth--

	return result, _libUpsertError
}

type hotelDist struct {
	hotel    *Hotel
	distance float64
}

func quickSortHotelDist(arr []hotelDist, from int, to int) {

	if from < 0 || from >= to {
		return
	}
	pivot := partitionHotelDist(arr, from, to)

	quickSortHotelDist(arr, from, pivot)
	quickSortHotelDist(arr, pivot+1, to)
}

func partitionHotelDist(arr []hotelDist, from int, to int) int {

	pivot := arr[from]
	pivotPos := from - 1

	for j := from + 1; j < to; j++ {
		if arr[j].distance < pivot.distance {
			pivotPos++

			temp := arr[pivotPos]
			arr[pivotPos] = arr[j]
			arr[j] = temp
		}
	}

	arr[from] = arr[pivotPos]
	arr[pivotPos] = pivot
	return pivotPos
}

func quickSortRate(arr []Hotel, from int, to int) {

	if from < 0 || from >= to {
		return
	}
	pivot := partitionRate(arr, from, to)

	quickSortRate(arr, from, pivot)
	quickSortRate(arr, pivot+1, to)
}

func partitionRate(arr []Hotel, from int, to int) int {

	pivot := arr[from]
	pivotPos := from - 1

	for j := from + 1; j < to; j++ {
		if arr[j].Rate < pivot.Rate {
			pivotPos++

			temp := arr[pivotPos]
			arr[pivotPos] = arr[j]
			arr[j] = temp
		}

	}

	arr[from] = arr[pivotPos]
	arr[pivotPos] = pivot
	return pivotPos
}
func (receiver City) GetId() string {
	return receiver.CityName
}
func (receiver *City) Init() {
	receiver.isInitialized = true
	receiver.Hotels = *lib.NewReferenceNavigationList[Hotel](receiver.CityName, receiver.GetTypeName(), "City", false)
}
func (receiver *City) saveChangesIfInitialized() error {
	if receiver.isInitialized && receiver.invocationDepth == 1 {
		_libError := lib.Upsert(receiver, receiver.CityName)
		if _libError != nil {
			return _libError
		}
	}
	return nil
}