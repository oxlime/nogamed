package main

import (
	"fmt"
	"time"
	"strconv"
	//"encoding/json"

	//"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/dontpanicdao/caigo/types"
)
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

func getResources(gw *gateway.GatewayProvider) ([3]int64) {
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
func estimateResourceProduction(gw *gateway.GatewayProvider, id int) (int64) {
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
