package domain

type RuleResult struct {
	RuleName string
	Tai      int64
}

type TaiRule func(tiles []*Tile, player *PlayerState) int64

var TaiRules = map[string]TaiRule{
	"PingHu":     CheckPengHu,
	"PengPengHu": CheckPengPengHu,
	"HonorPungs": CheckHonorPungs,
}
