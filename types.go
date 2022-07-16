package main

type MineLevels struct {
	metal     int64
	crystal   int64
	deuterium int64
	solar     int64
	robot     int64
}

type Mine struct {
	Name      string `json:"name"`
	MineLevel int64  `json:"mineLevel"`
}

type Strat struct {
	Mines []Mine `json:"Mine"`
}

/*
type Cost struct {
	metal 		int64
	crystal 	int64
	deuterium int64
}
*/

/*
type StructureUpgradeCostResponse struct {
	metal_mine_metal					int64	`json:"metal_mine.metal"`
	metal_mine_crystal				int64	`json:"metal_mine.crystal"`
	metal_mine_deuterium			int64	`json:"metal_mine.deuterium"`
	crystal_mine_metal				int64	`json:"crystal_mine.metal"`
	crystal_mine_crystal			int64	`json:"crystal_mine.crystal"`
	crystal_mine_deuterium		int64	`json:"crystal_mine.deuterium"`
	deuterium_mine_metal			int64	`json:"deuterium_mine.metal"`
	deuterium_mine_crystal		int64	`json:"deuterium_mine.crystal"`
	deuterium_mine_deuterium	int64	`json:"deuterium_mine.deuterium"`
}
*/
