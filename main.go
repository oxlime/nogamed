package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/dontpanicdao/caigo/types"
)

var (
	name                string = "testnet"
	local               string = "local"
	nogameContract      string = "0x035401b96dc690eda2716068d3b03732d7c18af7c0327787660179108789d84f"
	nogameScore         string = "0x025c1d0a3cfab1f5464b2e6a38c91c89bea77397744a7eb24b3f3645108d4abb"
	leaderboardContract string = "0x04358e376b5c68f17dc1cbdbde19914f1dd6e52a2eddb5b4b0d694716fe5d89b"
	//	address         string = "0xdeadbeef"
	//	privateKey      string = "0x12345678"
	//	addressDec	string = "1234123412341234" //address in decimal instead of hex
	feeMargin    uint64 = 115
	maxPoll      int    = 5
	pollInterval int    = 150
)

func main() {
	leaderboard()
	//startMine()
}

func startMine() {

	// init starknet gateway client
	gw := gateway.NewProvider(gateway.WithChain(name))

	//init account handler
	account, err := caigo.NewAccount(privateKey, address, gw)
	if err != nil {
		panic(err.Error())
	}

	//collect resources
	fmt.Println("Collecting Resources")
	collectResources(gw, account)

	//Check for building upgrades in progress
	building, buildTime, err := getBuildTimeCompletion(gw)
	buildingId := building.String()
	_ = buildTime
	if err != nil {
		panic(err.Error())
	}
	if buildingId == "0x1" {
		completeMineUpgrade(gw, "metal", account)
	}
	if buildingId == "0x2" {
		completeMineUpgrade(gw, "crystal", account)
	}
	if buildingId == "0x3" {
		completeMineUpgrade(gw, "deuterium", account)
	}
	if buildingId == "0x4" {
		completeMineUpgrade(gw, "solar_plant", account)
	}
	if buildingId == "0x5" {
		completeMineUpgrade(gw, "robot_factory", account)
	}

	//loop for start upgrade -> complete upgrade in accordance with strategy guide order
	var mineLevels MineLevels
	metalLvl, crystalLvl, deuteriumLvl, solarLvl, robotLvl, err := getMineLevels(gw)
	if err != nil {
		panic(err.Error())
	}
	mineLevels.metal = metalLvl
	mineLevels.crystal = crystalLvl
	mineLevels.deuterium = deuteriumLvl
	mineLevels.solar = solarLvl
	mineLevels.robot = robotLvl
	fmt.Println("Mine Levels: ", mineLevels)

	//get strat structure from strat.json file
	fileContent, err := os.Open("strat.json")
	if err != nil {
		panic(err.Error())
	}
	defer fileContent.Close()
	byteResult, _ := ioutil.ReadAll(fileContent)
	var strat Strat
	json.Unmarshal(byteResult, &strat)

	for i := 0; i < len(strat.Mines); i++ {
		if strat.Mines[i].Name == "Solar_Plant" {
			if strat.Mines[i].MineLevel > mineLevels.solar {
				for {
					fmt.Println("Upgrading : ", strat.Mines[i].Name)
					err := upgradeMine(gw, "solar_plant", account)
					if err == nil {
						break
					}
					if err != nil {
						if strings.Contains(err.Error(), "Error message: ERC20: burn amount exceeds balance") {
							fmt.Println("Erc20 burn amount exceeds balance, retying after 10 minutes and collecting resources")
							time.Sleep(10 * time.Minute)
							collectResources(gw, account)
						} else {
							panic(err.Error())
						}
					}
				}
			}
		}
		if strat.Mines[i].Name == "Metal_Mine" {
			if strat.Mines[i].MineLevel > mineLevels.metal {
				for {
					fmt.Println("Upgrading : ", strat.Mines[i].Name)
					err := upgradeMine(gw, "metal", account)
					if err == nil {
						break
					}
					if err != nil {
						if strings.Contains(err.Error(), "Error message: ERC20: burn amount exceeds balance") {
							fmt.Println("Erc20 burn amount exceeds balance, retying after 10 minutes and collecting resources")
							time.Sleep(10 * time.Minute)
							collectResources(gw, account)
						} else {
							panic(err.Error())
						}
					}
				}
			}
		}
		if strat.Mines[i].Name == "Crystal_Mine" {
			if strat.Mines[i].MineLevel > mineLevels.crystal {
				for {
					fmt.Println("Upgrading : ", strat.Mines[i].Name)
					err := upgradeMine(gw, "crystal", account)
					if err == nil {
						break
					}
					if err != nil {
						if strings.Contains(err.Error(), "Error message: ERC20: burn amount exceeds balance") {
							fmt.Println("Erc20 burn amount exceeds balance, retying after 10 minutes and collecting resources")
							time.Sleep(10 * time.Minute)
							collectResources(gw, account)
						} else {
							panic(err.Error())
						}
					}
				}
			}
		}
		if strat.Mines[i].Name == "Deuterium_Synthesizer" {
			if strat.Mines[i].MineLevel > mineLevels.deuterium {
				for {
					fmt.Println("Upgrading : ", strat.Mines[i].Name)
					err := upgradeMine(gw, "deuterium", account)
					if err == nil {
						break
					}
					if err != nil {
						if strings.Contains(err.Error(), "Error message: ERC20: burn amount exceeds balance") {
							fmt.Println("Erc20 burn amount exceeds balance, retying after 10 minutes and collecting resources")
							time.Sleep(10 * time.Minute)
							collectResources(gw, account)
						} else {
							panic(err.Error())
						}
					}
				}
			}
		}
		if strat.Mines[i].Name == "Robotics_Factory" {
			if strat.Mines[i].MineLevel > mineLevels.robot {
				for {
					fmt.Println("Upgrading : ", strat.Mines[i].Name)
					err := upgradeMine(gw, "robot_factory", account)
					if err == nil {
						break
					}
					if err != nil {
						if strings.Contains(err.Error(), "Error message: ERC20: burn amount exceeds balance") {
							fmt.Println("Erc20 burn amount exceeds balance, retying after 10 minutes and collecting resources")
							time.Sleep(10 * time.Minute)
							collectResources(gw, account)
						} else {
							panic(err.Error())
						}
					}
				}
			}
		}
	}

	fmt.Println("done")
}

func leaderboard() {

	// init starknet gateway client
	gw := gateway.NewProvider(gateway.WithChain(name))

	resp, err := gw.Call(context.Background(), types.FunctionCall{
		ContractAddress:    leaderboardContract,
		EntryPointSelector: "get_owners_array",
		Calldata: []string{
			"1067376053791535235517906780068452549623571831194888750889088837380668738235",
		},
	}, "")
	if err != nil {
		panic(err.Error())
	}

	respoints, err := gw.Call(context.Background(), types.FunctionCall{
		ContractAddress:    leaderboardContract,
		EntryPointSelector: "get_points_array",
		Calldata: []string{
			"1067376053791535235517906780068452549623571831194888750889088837380668738235",
		},
	}, "")
	if err != nil {
		panic(err.Error())
	}

	for index, element := range respoints {
		fmt.Println("Wallet: ", resp[index], " Score: ", types.StrToFelt(element))
	}
}
