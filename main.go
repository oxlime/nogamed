package main

import (
	"fmt"
	"time"
	"context"
	"math/big"
	"strconv"
	"strings"
	"io/ioutil"
	"os"
	"encoding/json"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/dontpanicdao/caigo/types"
)

var (
	name            string = "testnet"
	nogameContract	string = "0x035401b96dc690eda2716068d3b03732d7c18af7c0327787660179108789d84f"
	nogameScore			string = "0x025c1d0a3cfab1f5464b2e6a38c91c89bea77397744a7eb24b3f3645108d4abb"
//	address         string = "0xdeadbeef"
//	privateKey      string = "0x12345678"
//	addressDec			string = "12353251531253215151235123" //your private key in decimal instead of hex
	feeMargin       int64  = 115
	maxPoll         int    = 5
	pollInterval    int    = 150
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
		Int: big.NewInt((feeEstimate.Amount * feeMargin) / 100),
	}
	fmt.Printf("Fee:\n\tEstimate\t\t%v wei\n\tEstimate+Margin\t\t%v wei\n\n", feeEstimate.Amount, fee)

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

func upgradeMine(gw *gateway.GatewayProvider, eps string, account *caigo.Account) (error) {
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
		Int: big.NewInt((feeEstimate.Amount * feeMargin) / 100),
	}
	fmt.Printf("Fee:\n\tEstimate\t\t%v wei\n\tEstimate+Margin\t\t%v wei\n\n", feeEstimate.Amount, fee)

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
		Int: big.NewInt((feeEstimate.Amount * feeMargin) / 100),
	}
	fmt.Printf("Fee:\n\tEstimate\t\t%v wei\n\tEstimate+Margin\t\t%v wei\n\n", feeEstimate.Amount, fee)

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
	startMine()
}

func startMine() {
	// init the stark curve with constants
	// 'WithConstants()' will pull the StarkNet 'pedersen_params.json' file if you don't have it locally
	curve, err := caigo.SC(caigo.WithConstants())
	if err != nil {
		panic(err.Error())
	}

	//uncomment for unused var errors
	//_ = curve
	//_ = big.NewInt

	// init starknet gateway client
	gw := gateway.NewProvider(gateway.WithChain(name))

	//init account handler
	account, err := caigo.NewAccount(&curve, privateKey, address, gw)
	if err != nil {
		panic(err.Error())
	}

	//Scoreboard
	//Need local node or spamming the sequencer
	//points := getAllPoints(gw, getAllOwnerAddr(gw))
	//addr := getAllOwnerAddr(gw)
	//points := getAllPoints(gw, addr)
	//_ = points //so no errors

	//get resources available
	/*
	callResp, err := ogamecall(gw, "resources_available")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Resources available metal: ", types.StrToFelt(callResp[0]), " crystal: ", types.StrToFelt(callResp[1]), " deuterium: ", types.StrToFelt(callResp[2]), " energy: ", types.StrToFelt(callResp[2]))
	*/
	resources := getResources(gw)
	for i, res := range resources {
		fmt.Println(i, res)
	}

	/*
	upgradeCosts, err := getStructureUpgradeCosts(gw)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Structure upgrade costs", upgradeCosts)
	*/

	//collect resources
	//collectResources(gw, account)

	//loop for start upgrade -> complete upgrade in accordance with strategy guide order
	var mineLevels MineLevels
	metalLvl, crystalLvl, deuteriumLvl, solarLvl, robotLvl , err := getMineLevels(gw)
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
					if err == nil{
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
				fmt.Println("Upgrading: ", strat.Mines[i].Name)
				upgradeMine(gw, "metal", account)
			}
		}
		if strat.Mines[i].Name == "Crystal_Mine" {
			if strat.Mines[i].MineLevel > mineLevels.crystal {
				fmt.Println("Upgrading: ", strat.Mines[i].Name)
				upgradeMine(gw, "crystal", account)
			}
		}
		if strat.Mines[i].Name == "Deuterium_Synthesizer" {
			if strat.Mines[i].MineLevel > mineLevels.deuterium {
				fmt.Println("Upgrading: ", strat.Mines[i].Name)
				upgradeMine(gw, "deuterium", account)
			}
		}
		if strat.Mines[i].Name == "Robotics_Factory" {
			if strat.Mines[i].MineLevel > mineLevels.robot {
				fmt.Println("Upgrading: ", strat.Mines[i].Name)
				upgradeMine(gw, "robot_factory", account)
			}
  	}
	}
	
	//if mineLevels < stratlevels 
	//for level in strat.levels 
	//  if level > minelevel
	//    check resources enough to start upgrade by checking balance of token by wallet address
	//		  if not collect resources
	//		if still not then calculate estimated wait time until enough
	//    upgrade mineLevel

	//fmt.Scanln()
	fmt.Println("done")
}
