package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	//"encoding/json"
	"context"
	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/dontpanicdao/caigo/types"
	"math/big"
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

/* cairo
(mine_factor mine_level)
    let fact1 = mine_factor * mine_level
    let (fact2) = pow(11, mine_level)
    local fact3 = fact1 * fact2
    let (fact4) = pow(10, mine_level)
    let (fact5, _) = unsigned_div_rem(fact3, fact4)
    return (production_hour=fact5)
*/
/*
func resourceProductionFormula(mineFactor types.Felt, mineLevel types.Felt) {
	fact1 := mineFactor * mineLevel
	var temp = big.NewInt(11)
	fact2 := temp.Exp(temp, mineLevel)
	fact3 := fact1.Mul(fact1, fact2)
	var temp2 = bigNewInt(10)
	fact4 := temp2.Exp(temp2, mineLevel)
	fact5

}
*/

func getResources(gw *gateway.GatewayProvider) [3]int64 {
	callResp, err := ogamecall(gw, "resources_available")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Resources available metal: ", types.StrToFelt(callResp[0]), " crystal: ", types.StrToFelt(callResp[1]), " deuterium: ", types.StrToFelt(callResp[2]), " energy: ", types.StrToFelt(callResp[2]))

	var temp string
	var resAvailable [3]int64
	for i := 0; i < 3; i++ {
		temp = callResp[i]
		resAvailable[i], err = strconv.ParseInt(temp[2:], 16, 64)
		if err != nil {
			panic(err)
		}
	}
	return resAvailable
}

//rough estimate of production per minute
//Get Resources available
//Wait 10 mins
//Check Resource avaible again
//Calculate production per minute
func estimateResourceProduction(gw *gateway.GatewayProvider, id int) int64 {
	callResp, err := ogamecall(gw, "resources_available")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Resources available metal: ", types.StrToFelt(callResp[0]), " crystal: ", types.StrToFelt(callResp[1]), " deuterium: ", types.StrToFelt(callResp[2]), " energy: ", types.StrToFelt(callResp[2]))

	strAmount := callResp[id]
	amount, err := strconv.ParseInt(strAmount[2:], 16, 64)
	if err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Minute)

	callResp2, err := ogamecall(gw, "resources_available")
	if err != nil {
		panic(err.Error())
	}
	strAmount2 := callResp2[id]
	amount2, err := strconv.ParseInt(strAmount2[2:], 16, 64)
	if err != nil {
		panic(err)
	}

	ppm := amount2 - amount
	return ppm
}

//Functions for leaderboard
//Scoreboard will need local node as it will spam sequencer with requests
func getAllOwnerAddr(gw *gateway.GatewayProvider) [200]string {
	var res [200]string
	for i := 1; i < 201; i++ {
		resp, err := gw.Call(context.Background(), types.FunctionCall{
			ContractAddress:    nogameScore,
			EntryPointSelector: "ownerOf",
			Calldata: []string{
				strconv.Itoa(i), "0",
			},
		}, "0x74c8899f93435848b9adb756e6efa3774e36c49c18bb2fa75f429e374a23506")
		if err != nil {
			//return nil, err
			panic(err.Error())
		}
		res[i-1] = resp[0]
	}
	return res
}

func getAllPoints(gw *gateway.GatewayProvider, address [200]string) [200]string {
	var res [200]string
	for i := 0; i < 200; i++ {
		resp, err := gw.Call(context.Background(), types.FunctionCall{
			ContractAddress:    nogameContract,
			EntryPointSelector: "player_points",
			Calldata: []string{
				address[i],
			},
		}, "0x74c8899f93435848b9adb756e6efa3774e36c49c18bb2fa75f429e374a23506")
		if err != nil {
			//return nil, err
			panic(err.Error())
		}
		res[i] = resp[0]
		fmt.Println("Player: ", address[i], " Score: ", res[i])
	}
	return res
}
