package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	// "strconv"
	// "strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)
const Version string = "0.1.4"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}


type Transaction struct {
	TransactionID		string `json:"TransactionID"` 
	FromAccount			string `json:"FromAccount"`
	ToAccount			string `json:"color"`
	Amount				string `json:"Amount"`
	FromAccountBalance	string `json:"FromAccountBalance"`
	ToAccountBalance	string `json:"ToAccountBalance"`
	OperateDate       	string `json:"OperateDate"`
	OperateTime       	string `json:"OperateTime"`
	TransactionDate     string `json:"TransactionDate"`
	TransactionTime     string `json:"TransactionTime"`
}

type Enterprice struct{
	EnterpriceID	string `json:"EnterpriceID"`
	Accounts		*map[string]Account `json:"Accounts"` 
}

type Account struct{
	AmountID		string `json:"AmountID"`
	Balance		string `json:"Balance"`
}
// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte(Version))
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function+", args: " + string(len(args)))

	// Handle different functions
	if function == "version" { //create a new marble
		return t.getChaincodeVersion(stub, args)
	}else if function == "createEnterprice"{
		return t.createEnterprice(stub, args)
	}else if function == "getEnterprice"{
		return t.getEnterprice(stub, args)
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) getChaincodeVersion(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success([]byte(Version))
}

func (t *SimpleChaincode) createEnterprice(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 0 {
		return shim.Error("Incorrect number of arguments. Expecting more than 1")
	}
	if len(args)%2 == 0 {
		return shim.Error("usage: enterpriceId, amount1, balance1, amount2, balance2, ...")
	}
	// transaction id is like 20170324_3390569288041060
	// enterprice id is like ENTERPRICE_enterpriceid

	enterpriceId := "ENTERPRICE_"+args[0]
	enterpriceAsBytes, err := stub.GetState(enterpriceId)

	if err != nil {
		return shim.Error("Failed to get enterprice: " + err.Error())
	} else if enterpriceAsBytes != nil {
		fmt.Println("This enterprice already exists: " + args[0])
		return shim.Error("This enterprice already exists: " + args[0])
	}

	accounts := make(map[string]Account)
	accountsCnt := (len(args)-1)/2
	for i:=0;i<accountsCnt;i++ {
		accountId := args[2*i+1]
		balance := args[2*i+2]
		account := Account{accountId, balance}
		accounts[accountId] = account
	}

	enterprice := &Enterprice{args[0], &accounts}
	enterpriceJSONasBytes, err := json.Marshal(enterprice)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(enterpriceId, enterpriceJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) getEnterprice(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if(len(args) == 0){
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	enterpriceId := "ENTERPRICE_"+args[0]
	enterpriceAsBytes, err := stub.GetState(enterpriceId)

	if err != nil {
		return shim.Error("Failed to get enterprice: " + err.Error())
	} else if enterpriceAsBytes == nil {
		fmt.Println("This enterprice don't exists: " + args[0])
		return shim.Error("This enterprice don't exists: " + args[0])
	}

	return shim.Success(enterpriceAsBytes)
}
