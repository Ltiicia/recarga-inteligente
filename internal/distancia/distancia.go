package distancia

import (
	"math"
)

// Converte decimais em radianos
func decToRad(dec float64) float64 {
	//formula de conversao
	rad := dec * (math.Pi / 180)
	return rad
}

// Calcula a variacao da latitude e da longitude
func getDelta(x1 float64, x2 float64) float64 {
	return (x2 - x1)
}

// Calcula a distancia entre os dois pontos usando a formula de Haversine
// a = [sen²(Δlatitude/2) + cos(latitude1)] x cos(latitude2) x sen²(Δlongitude/2)
// c = 2 x atan²(√a,√(1−a))
// d = r x c
func getDistancia(latitude1, longitude1, latitude2, longitude2 float64) float64 {
	//Raio da Terra em metros
	const raioTerra_m = 6371000

	latitude1, latitude2 = decToRad(latitude1), decToRad(latitude2)
	longitude1, longitude2 = decToRad(longitude1), decToRad(longitude2)

	deltaLatitude := getDelta(latitude1, latitude2)
	deltaLongitude := getDelta(longitude1, longitude2)

	a := math.Pow(math.Sin(deltaLatitude/2), 2) + math.Cos(latitude1)*math.Cos(latitude2)*math.Pow(math.Sin(deltaLongitude/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distancia := raioTerra_m * c
	return distancia
}
