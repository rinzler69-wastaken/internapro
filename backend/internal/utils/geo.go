package utils

import (
	"math"
)

const earthRadiusKm = 6371.0

// HaversineDistance calculates the distance between two points on Earth
// using the Haversine formula. Returns distance in meters.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Differences
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers
	distanceKm := earthRadiusKm * c

	// Convert to meters
	return distanceKm * 1000
}

// IsWithinRadius checks if a point is within a certain radius of another point
func IsWithinRadius(lat1, lon1, lat2, lon2, radiusMeters float64) bool {
	distance := HaversineDistance(lat1, lon1, lat2, lon2)
	return distance <= radiusMeters
}

// degreesToRadians converts degrees to radians
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// ValidateCoordinates checks if coordinates are valid
func ValidateCoordinates(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}
