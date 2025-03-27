package coordenadas

import (
	"math/rand"
	"recarga-inteligente/internal/dataJson"
)

func GetLocalizacaoVeiculo(area dataJson.Area) dataJson.Localizacao {
	var localizacaoAtual dataJson.Localizacao
	localizacaoAtual.Latitude = area.Latitude_min + rand.Float64()*(area.Latitude_max-area.Latitude_min)
	localizacaoAtual.Longitude = area.Longitude_min + rand.Float64()*(area.Longitude_max-area.Longitude_min)

	return localizacaoAtual
}
