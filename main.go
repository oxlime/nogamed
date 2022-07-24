package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
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
	leaderboardContract string = "0x0326d5c668224c79be41582a1a63ab7c8213e82c23ee94d28ddd93344c5cb105"
	//	address         string = "0xdeadbeef"
	//	privateKey      string = "0x12345678"
	//	addressDec	string = "1234123412341234" //address in decimal instead of hex
	feeMargin    uint64 = 115
	maxPoll      int    = 5
	pollInterval int    = 150
)

func ogamecall(gw *gateway.GatewayProvider, eps string) ([]string, error) {
	resp, err := gw.Call(context.Background(), types.FunctionCall{
		ContractAddress:    nogameContract,
		EntryPointSelector: eps,
		Calldata: []string{
			addressDec,
		},
	}, "")
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func collectResources(gw *gateway.GatewayProvider, account *caigo.Account) {
	collect := []types.Transaction{
		{
			ContractAddress:    nogameContract,
			EntryPointSelector: "collect_resources",
		},
	}

	// estimate fee for executing transaction
	feeEstimate, err := account.EstimateFee(context.Background(), collect)
	if err != nil {
		panic(err.Error())
	}
	fee := types.Felt{
		Int: new(big.Int).SetUint64((feeEstimate.OverallFee * feeMargin) / 100),
	}
	fmt.Printf("Fee:\n\tEstimate\t\t%v wei\n\tEstimate+Margin\t\t%v wei\n\n", feeEstimate.OverallFee, fee)

	// execute transaction
	execResp, err := account.Execute(context.Background(), &fee, collect)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("execution response: ", execResp.TransactionHash)

	n, receipt, err := gw.PollTx(context.Background(), execResp.TransactionHash, types.ACCEPTED_ON_L2, pollInterval, maxPoll)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Poll %dsec %dx \n\ttransaction(%s) receipt: %s\n\n", n*pollInterval, n, execResp.TransactionHash, receipt.Status)
}

func upgradeMine(gw *gateway.GatewayProvider, eps string, account *caigo.Account) error {
	increment := []types.Transaction{
		{
			ContractAddress:    nogameContract,
			EntryPointSelector: eps + "_upgrade_start",
		},
	}

	// estimate fee for executing transaction
	feeEstimate, err := account.EstimateFee(context.Background(), increment)
	if err != nil {
		if strings.Contains(err.Error(), "Error message: ERC20: burn amount exceeds balance") {
			time.Sleep(10 * time.Second)
			fmt.Println("caught error")
			return err
		}
		panic(err.Error())
	}
	fee := types.Felt{
		Int: new(big.Int).SetUint64((feeEstimate.OverallFee * feeMargin) / 100),
	}
	fmt.Printf("Fee:\n\tEstimate\t\t%v wei\n\tEstimate+Margin\t\t%v wei\n\n", feeEstimate.OverallFee, fee)

	// execute transaction
	execResp, err := account.Execute(context.Background(), &fee, increment)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("execution response: ", execResp.TransactionHash)

	n, receipt, err := gw.PollTx(context.Background(), execResp.TransactionHash, types.ACCEPTED_ON_L2, pollInterval, maxPoll)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Poll %dsec %dx \n\ttransaction(%s) receipt: %s\n\n", n*pollInterval, n, execResp.TransactionHash, receipt.Status)

	buildingId, buildTime, err := getBuildTimeCompletion(gw)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("building  id: ", buildingId)
	fmt.Println("build time completion: ", buildTime)
	completeMineUpgrade(gw, eps, account)
	return nil
}

func completeMineUpgrade(gw *gateway.GatewayProvider, eps string, account *caigo.Account) {
	complete := []types.Transaction{
		{
			ContractAddress:    nogameContract,
			EntryPointSelector: eps + "_upgrade_complete",
		},
	}

	// estimate fee for executing transaction
	feeEstimate, err := account.EstimateFee(context.Background(), complete)
	if err != nil {
		panic(err.Error())
	}
	fee := types.Felt{
		Int: new(big.Int).SetUint64((feeEstimate.OverallFee * feeMargin) / 100),
	}
	fmt.Printf("Fee:\n\tEstimate\t\t%v wei\n\tEstimate+Margin\t\t%v wei\n\n", feeEstimate.OverallFee, fee)

	// execute transaction
	execResp, err := account.Execute(context.Background(), &fee, complete)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("execution response: ", execResp.TransactionHash)

	n, receipt, err := gw.PollTx(context.Background(), execResp.TransactionHash, types.ACCEPTED_ON_L2, pollInterval, maxPoll)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Poll %dsec %dx \n\ttransaction(%s) receipt: %s\n\n", n*pollInterval, n, execResp.TransactionHash, receipt.Status)
}

func getMineLevels(gw *gateway.GatewayProvider) (int64, int64, int64, int64, int64, error) {
	resp, err := ogamecall(gw, "get_structures_levels")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	strmetal := resp[0]
	metal, err := strconv.ParseInt(strmetal[2:], 16, 64)
	if err != nil {
		panic(err)
	}
	strcrystal := resp[1]
	crystal, err := strconv.ParseInt(strcrystal[2:], 16, 64)
	if err != nil {
		panic(err)
	}
	strdeuterium := resp[2]
	deuterium, err := strconv.ParseInt(strdeuterium[2:], 16, 64)
	if err != nil {
		panic(err)
	}
	strsolar := resp[3]
	solar, err := strconv.ParseInt(strsolar[2:], 16, 64)
	if err != nil {
		panic(err)
	}
	strrobot := resp[4]
	robot, err := strconv.ParseInt(strrobot[2:], 16, 64)
	if err != nil {
		panic(err)
	}
	return metal, crystal, deuterium, solar, robot, nil
}

func getBuildTimeCompletion(gw *gateway.GatewayProvider) (*types.Felt, *types.Felt, error) {
	resp, err := ogamecall(gw, "build_time_completion")
	if err != nil {
		return nil, nil, err
	}
	buildingId := types.StrToFelt(resp[0])
	timeEnd := types.StrToFelt(resp[1])
	//test
	respStr := resp[1]
	respStr = respStr[2:]
	timeInt, err := strconv.ParseInt(respStr, 16, 64)
	if err != nil {
		panic(err)
	}
	wTime(timeInt)
	return buildingId, timeEnd, nil
}

//test
func wTime(wait int64) {
	fmt.Println("wait and time in unix", wait, time.Now().Unix())
	sleepTime := wait - time.Now().Unix() + 265 //adding 65 seconds buffer
	time.Sleep(time.Duration(sleepTime * 1000000000))
	fmt.Println("sleep time = ", sleepTime)
}

func main() {
	//leaderboard()
	startMine()
}

func startMine() {
	// init the stark curve with constants
	// 'WithConstants()' will pull the StarkNet 'pedersen_params.json' file if you don't have it locally
	//curve, err := caigo.SC(caigo.WithConstants())
	//if err != nil {
	//		panic(err.Error())
	//	}

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
			"1",
			"0",
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
			"1",
			"0",
		},
	}, "")
	if err != nil {
		panic(err.Error())
	}

	for index, element := range respoints {
		fmt.Println("Wallet: ", resp[index], " Score: ", types.StrToFelt(element))
	}

	//mines := getResources(gw)
	//fmt.Println("Resources on id 2: ", mines[2])

}
