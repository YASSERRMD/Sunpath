package geo

type Building struct {
	OSMID           int64
	Footprint       Polygon
	Height          float64
	HeightEstimated bool
}
