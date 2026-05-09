package service

import (
	"context"
	"math"
)

type GeoService struct {
}

func NewGeoService() *GeoService {
	return &GeoService{}
}

func (s *GeoService) GetLocationFromIP(ip string) (lat, long float64, err error) {
	lat, long = simpleIPToLocation(ip)
	return lat, long, nil
}

func (s *GeoService) CalculateDistance(loc1Lat, loc1Long, loc2Lat, loc2Long float64) float64 {
	const earthRadius = 6371.0

	lat1Rad := loc1Lat * math.Pi / 180
	lat2Rad := loc2Lat * math.Pi / 180
	deltaLat := (loc2Lat - loc1Lat) * math.Pi / 180
	deltaLong := (loc2Long - loc1Long) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLong/2)*math.Sin(deltaLong/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func (s *GeoService) IsImpossibleTravel(ctx context.Context, userID string, newLat, newLong float64) (bool, float64, error) {
	return false, 0, nil
}

func simpleIPToLocation(ip string) (float64, float64) {
	return 39.9042 + float64(len(ip)%10)*0.01, 116.4074 + float64(len(ip)%10)*0.01
}