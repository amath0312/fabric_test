package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	// "strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"errors"
	"net"
	"sync"
	"time"
)
const Version string = "0.0.8"

const ENTERPRISE_STATUS_NORMAL = "NORMAL"
const ENTERPRISE_STATUS_CLOSED = "CLOSED"
const ENTERPRISE_STATUS_FROZEN = "FROZEN"

const PREFIX_ENTERPRISE = "ENTPRISE_"
const PREFIX_TASK = "TASK0000_"

const SUFFIX_MIN_TRANSID = "0000000000000000"
const SUFFIX_MAX_TRANSID = "9999999999999999"

var sf *Sonyflake
var st Settings

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type TransactionTask struct{
	TaskID	string `json:"TaskID"`
	TransactionIDList	[]string `json:"TransactionIDList"`
}
// transaction id is like 20170324_3390569288041060
type Transaction struct {
	TransactionID		string `json:"TransactionID"` 
	FromEnterprise		string `json:"FromEnterprise"`
	FromAccount			string `json:"FromAccount"`
	FromAccountBalance	string `json:"FromAccountBalance"`
	ToEnterprise		string `json:"ToEnterprise"`
	ToAccount			string `json:"ToAccount"`
	ToAccountBalance	string `json:"ToAccountBalance"`
	Amount				string `json:"Amount"`
	OperateDate       	string `json:"OperateDate"`
	OperateTime       	string `json:"OperateTime"`
	TransactionDate     string `json:"TransactionDate"`
	TransactionTime     string `json:"TransactionTime"`
}

// Enterprise id is like ENTPRISE_Enterpriseid
type Enterprise struct{
	EnterpriseID	string `json:"EnterpriseID"`
	Status			string `json:"Status"`
	Accounts		map[string]Account `json:"Accounts"` 
}

type Account struct{
	AccountID		string `json:"AccountID"`
	Balance		string `json:"Balance"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	st.StartTime = time.Date(2017, 3, 1, 0, 0, 0, 0, time.UTC)
	sf = NewSonyflake(st)
	fmt.Println("run snowflake!!!!!")

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
	}else if function == "createEnterprise"{
		return t.createEnterprise(stub, args)
	}else if function == "getEnterprise"{
		return t.getEnterprise(stub, args)
	}else if function == "addAccounts"{
		return t.addAccounts(stub,args)
	}else if function == "makeTransactions"{
		return t.makeTransactions(stub,args)
	}else if function == "getTask"{
		return t.getTask(stub,args)
	}else if function == "getTransaction"{
		return t.getTransaction(stub, args)
	}else if function == "queryTransactionsByDate"{
		return t.queryTransactionsByDate(stub, args)
	}else if function == "queryTransaction"{
		return t.queryTransaction(stub, args)
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) getChaincodeVersion(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success([]byte(Version))
}

func (t *SimpleChaincode) createEnterprise(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 0 {
		return shim.Error("Incorrect number of arguments. Expecting more than 1")
	}
	if len(args)%2 == 0 {
		return shim.Error("usage: EnterpriseId, amount1, balance1, amount2, balance2, ...")
	}
	
	hasEnterprise, err := t.hasEnterprise(stub, args[0])
	if err != nil{
		return shim.Error(err.Error())
	}else if hasEnterprise{
		return shim.Error("This Enterprise already exists: " + args[0])
	}

	enterprise := &Enterprise{args[0], ENTERPRISE_STATUS_NORMAL,nil}
	err = t.addAccountsToEnterprise(enterprise, args[1:])
	if err != nil {
		return shim.Error(err.Error())
	}

	err = t.saveEnterprise(stub, enterprise)
	if err != nil{
		shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) addAccounts(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 0 {
		return shim.Error("Incorrect number of arguments. Expecting more than 1")
	}
	if len(args)%2 == 0 {
		return shim.Error("usage: EnterpriseId, amount1, balance1, amount2, balance2, ...")
	}

	enterpriseAsBytes, err := t.getEnterpriseAsBytes(stub, args[0])
	if err != nil{
		return shim.Error(err.Error())
	}else if enterpriseAsBytes == nil{
		return shim.Error("This Enterprise don't exist: " + args[0])
	}

	enterprise := Enterprise{}
	err = json.Unmarshal(enterpriseAsBytes, &enterprise)
	if err != nil{
		return shim.Error(err.Error())
	}

	err = t.addAccountsToEnterprise(&enterprise, args[1:])
	if err != nil {
		return shim.Error(err.Error())
	}
	
	err = t.saveEnterprise(stub, &enterprise)
	if err != nil{
		shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) makeTransactions(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if(len(args) == 0){
		return shim.Error("Incorrect number of arguments. Expecting more than 1")
	}
	if (len(args)-1) % 7 != 0{
		return shim.Error("Incorrect number of arguments. Expecting: \r\nTaskID, Trans1_From_Enterprise,Trans1_From_Account,Trans1_To_Enterprise,Trans1_to_Account,Trans1_Amount,Trans1_Date,Trans1_Time,trans2_From_Enterprise,trans2_From_Account,trans2_To_Enterprise,trans2_to_Account,trans2_Amount,trans2_Date,trans2_Time,...  ")
	}
	taskId := args[0]
	taskAsBytes, err := t.getTaskAsBytes(stub, taskId)

	if err != nil {
		return shim.Error("Failed to get task: " + err.Error())
	} else if taskAsBytes != nil {
		return shim.Error("This task already exists: " + args[0])
	}

	task := &TransactionTask{taskId,[]string{}}
	transactionCnt := (len(args)-1)/7
	tmpEnterprises := make(map[string]Enterprise)
	for i:=0;i<transactionCnt;i++{
		id,err := sf.NextID()
		if err != nil{
			return shim.Success(nil)
		}

		index := 7*i+1
		now := now()
		transIdSeed := strconv.FormatUint(uint64(id), 10)
		fromEnterpriseId := args[index]
		fromAccountId := args[index+1]
		toEnterpriseId := args[index+2]
		toAccountId := args[index+3]
		amount, err := strconv.Atoi(args[index+4])
		if err != nil{
			return shim.Error("invalid amount value: "+args[index+4])
		}
		transDate := args[index+5]
		transTime := args[index+6]
		operDate := fmt.Sprintf("%04d%02d%02d",now.Year(),now.Month(),now.Day())
		operTime := fmt.Sprintf("%02d%02d%02d.%03d",now.Hour(), now.Minute(), now.Second(),now.Nanosecond()/1000000)
		transId := fmt.Sprintf("%s_%s",transDate,transIdSeed)

		fromBalance, err := t.addBalance(stub,fromEnterpriseId,fromAccountId,-1*amount,tmpEnterprises)
		if err != nil{
			return shim.Error(err.Error())
		}
		toBalance, err := t.addBalance(stub,toEnterpriseId,toAccountId,amount,tmpEnterprises)
		if err != nil{
			return shim.Error(err.Error())
		}

		transaction := &Transaction{
			transId,
			fromEnterpriseId,fromAccountId,strconv.Itoa(fromBalance),
			toEnterpriseId,toAccountId,strconv.Itoa(toBalance),
			args[index+4],
			operDate,operTime,
			transDate,transTime}

		err = t.saveTransaction(stub, transaction)
		if err != nil{
			return shim.Error(err.Error())
		}
		task.TransactionIDList = append(task.TransactionIDList,transId)
	}
	for _, enterprise := range tmpEnterprises {
        err = t.saveEnterprise(stub, &enterprise)
    	if err != nil{
			shim.Error(err.Error())
		}
	}

	err = t.saveTask(stub, task)
	if err != nil{
		shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) getTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if(len(args) == 0){
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	taskAsBytes, err := t.getTaskAsBytes(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	} else if taskAsBytes == nil {
		fmt.Println("This task don't exists: " + args[0])
		return shim.Error("This task don't exists: " + args[0])
	}

	return shim.Success(taskAsBytes)
}

func (t *SimpleChaincode) queryTransactionsByDate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if(len(args) == 0){
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	startDate := args[0]+"_"+SUFFIX_MIN_TRANSID
	endDate := args[1]+"_"+SUFFIX_MAX_TRANSID

	resultsIterator, err := stub.RangeQueryState(startDate, endDate)

	if err != nil {
		return shim.Error("Failed to get Enterprise: " + err.Error())
	} 
	defer resultsIterator.Close()
	
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResultKey, queryResultValue, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TransactionID\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResultKey)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Transaction\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResultValue))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

func (t *SimpleChaincode) getTransaction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if(len(args) == 0){
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	transId := args[0]
	transAsBytes, err := t.getTransactionAsBytes(stub, transId)

	if err != nil {
		return shim.Error(err.Error())
	} else if transAsBytes == nil {
		return shim.Error("This transaction don't exists: " + transId)
	}

	return shim.Success(transAsBytes)
}

func (t *SimpleChaincode) getEnterprise(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if(len(args) == 0){
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	EnterpriseAsBytes, err := t.getEnterpriseAsBytes(stub, args[0])

	if err != nil {
		return shim.Error("Failed to get Enterprise: " + err.Error())
	} else if EnterpriseAsBytes == nil {
		shim.Success(nil)
	}

	return shim.Success(EnterpriseAsBytes)
}

// func (t *SimpleChaincode) queryTransactionByAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	if len(args) < 2 {
// 		return shim.Error("Incorrect number of arguments. Expecting 2")
// 	}

// 	enterpriseId := args[0]
// 	accountId := args[1]

// 	resultsIterator, err := stub.PartialCompositeKeyQuery("enterprise~account~date~transaction", []string{enterpriseId, accountId})
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	defer resultsIterator.Close()
// 	var buffer bytes.Buffer
// 	buffer.WriteString("[")
// 	bArrayMemberAlreadyWritten := false

// 	// Iterate through result set and for each marble found, transfer to newOwner
// 	var i int
// 	for i = 0; resultsIterator.HasNext(); i++ {
// 		// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
// 		nameKey, _, err := resultsIterator.Next()
// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}
// 		if bArrayMemberAlreadyWritten == true {
// 			buffer.WriteString(",")
// 		}

// 		_, compositeKeyParts, err := stub.SplitCompositeKey(nameKey)
// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}
// 		// enterprise := compositeKeyParts[0]
// 		// account := compositeKeyParts[1]
// 		// date := compositeKeyParts[2]
// 		transactionId := compositeKeyParts[3]
// 		transAsBytes, err := t.getTransactionAsBytes(stub, transactionId)

// 		if err != nil {
// 			return shim.Error(err.Error())
// 		} else if transAsBytes == nil {
// 			return shim.Error("This transaction don't exists: " + transactionId)
// 		}
// 		buffer.WriteString("{")
// 		buffer.WriteString("\"TransactionID\":\"")
// 		buffer.WriteString(transactionId)
// 		buffer.WriteString("\",\"Transaction\":")
// 		buffer.WriteString(string(transAsBytes))
// 		buffer.WriteString("}")
// 		bArrayMemberAlreadyWritten = true
// 	}
// 	buffer.WriteString("]")

// 	return shim.Success(buffer.Bytes())
// }


func (t *SimpleChaincode) queryTransaction(stub shim.ChaincodeStubInterface, args []string)pb.Response{
	if len(args) == 0 {
		return shim.Error("Incorrect number of arguments")
	}

	resultsIterator, err := stub.PartialCompositeKeyQuery("enterprise~account~date~transaction", args)
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	// Iterate through result set and for each marble found, transfer to newOwner
	var i int
	for i = 0; resultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
		nameKey, _, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(nameKey)
		if err != nil {
			return shim.Error(err.Error())
		}
		// enterprise := compositeKeyParts[0]
		// account := compositeKeyParts[1]
		// date := compositeKeyParts[2]
		transactionId := compositeKeyParts[3]
		transAsBytes, err := t.getTransactionAsBytes(stub, transactionId)

		if err != nil {
			return shim.Error(err.Error())
		} else if transAsBytes == nil {
			return shim.Error("This transaction don't exists: " + transactionId)
		}
		buffer.WriteString("{")
		buffer.WriteString("\"TransactionID\":\"")
		buffer.WriteString(transactionId)
		buffer.WriteString("\",\"Transaction\":")
		buffer.WriteString(string(transAsBytes))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}



func (t *SimpleChaincode) getEnterpriseAsBytes(stub shim.ChaincodeStubInterface, enterpriseId string) ([]byte, error) {
	innerEnterpriseId := PREFIX_ENTERPRISE+enterpriseId
	EnterpriseAsBytes, err := stub.GetState(innerEnterpriseId)

	if err != nil {
		return nil, errors.New("Failed to get Enterprise: " + err.Error())
	} else if EnterpriseAsBytes == nil {
		return nil, nil
	}

	return EnterpriseAsBytes,nil
}
func (t *SimpleChaincode) hasEnterprise(stub shim.ChaincodeStubInterface, enterpriseId string) (bool, error){
	enterpriseAsBytes, err := t.getEnterpriseAsBytes(stub,enterpriseId)

	if err != nil {
		return false, err
	} else if enterpriseAsBytes == nil {
		return false, nil
	} else{
		return true, nil
	}
}
func (t *SimpleChaincode) addAccountsToEnterprise(enterprise *Enterprise, args []string) (error) {
	if len(args) == 0 {
		return nil
	}
	if len(args)%2 != 0 {
		return errors.New("wrong number of account arguments")
	}

	accountsCnt := (len(args))/2
	accounts := enterprise.Accounts
	if accounts == nil{
		accounts = make(map[string]Account)
		enterprise.Accounts = accounts
	}
	// verify accounts whether is already existed
	for i:=0;i<accountsCnt;i++ {
		accountId := args[2*i]
		if _, ok := accounts[accountId]; ok{
			return errors.New("thie account already exists: "+accountId)
		}
	}

	// add accounts
	for i:=0;i<accountsCnt;i++ {
		accountId := args[2*i]
		balance := args[2*i+1]
		account := Account{accountId, balance}
		accounts[accountId] = account
	}
	
	return nil
}
func (t *SimpleChaincode) saveEnterprise(stub shim.ChaincodeStubInterface, enterprise *Enterprise) (error) {
	enterpriseJSONasBytes, err := json.Marshal(enterprise)
	if err != nil {
		return err
	}

	innerEnterpriseId := PREFIX_ENTERPRISE+enterprise.EnterpriseID

	err = stub.PutState(innerEnterpriseId, enterpriseJSONasBytes)
	if err != nil {
		return err
	}
	return nil
}


func (t *SimpleChaincode) addBalance(stub shim.ChaincodeStubInterface, enterpriseId string, accountId string, amount int, tmpEnterprises map[string]Enterprise) (int,error){
	enterprise, ok:= tmpEnterprises[enterpriseId]
	if !ok{
		enterpriseAsBytes, err:= t.getEnterpriseAsBytes(stub,enterpriseId)
		if err != nil || enterpriseAsBytes == nil{
			return 0,errors.New("Enterprise not exists: "+enterpriseId)
		}
		enterprise = Enterprise{}
		err = json.Unmarshal(enterpriseAsBytes, &enterprise)
		if err != nil{
			return 0,errors.New("Enterprise occur some error: "+enterpriseId)
		}
		tmpEnterprises[enterpriseId] = enterprise
	}

	account,ok:=enterprise.Accounts[accountId]

	if !ok{
		return 0,errors.New("Account of "+enterpriseId+" not exists: "+accountId)
	}
	accountBalance ,err := strconv.Atoi(account.Balance)
	if err != nil {
		accountBalance = 0
	}
	accountBalance += amount
	account.Balance = strconv.Itoa(accountBalance)
	enterprise.Accounts[accountId] = account
	// enterpriseJSONasBytes, err := json.Marshal(fromEnterprise)	
	// err = stub.PutState(PREFIX_ENTERPRISE+fromEnterpriseId, fromEnterpriseJSONasBytes) 
	return accountBalance,nil
}

func (t *SimpleChaincode) getTaskAsBytes(stub shim.ChaincodeStubInterface, taskId string) ([]byte, error) {
	innerTaskId := PREFIX_TASK + taskId
	taskAsBytes, err := stub.GetState(innerTaskId)

	if err != nil {
		return nil, errors.New("Failed to get Transaction: " + err.Error())
	} else if taskAsBytes == nil {
		return nil, nil
	}

	return taskAsBytes,nil
}

func (t *SimpleChaincode) saveTask(stub shim.ChaincodeStubInterface, task *TransactionTask) (error) {
	taskJSONasBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}
	
	taskId := PREFIX_TASK + task.TaskID
	err = stub.PutState(taskId, taskJSONasBytes)
	if err != nil {
		return err
	}
	return nil
}



func (t *SimpleChaincode) getTransactionAsBytes(stub shim.ChaincodeStubInterface, transactionId string) ([]byte, error) {
	transactionAsBytes, err := stub.GetState(transactionId)

	if err != nil {
		return nil, errors.New("Failed to get Transaction: " + err.Error())
	} else if transactionAsBytes == nil {
		return nil, errors.New("This Transaction don't exists: " + transactionId)
	}

	return transactionAsBytes,nil
}

func (t *SimpleChaincode) saveTransaction(stub shim.ChaincodeStubInterface, transaction *Transaction)(error){
	transactionJSONasBytes, err := json.Marshal(transaction)
	transId := transaction.TransactionID

	err = stub.PutState(transId, transactionJSONasBytes) 
	if err != nil {
		return err
	}

	indexName := "enterprise~account~date~transaction"
	indexValue := []byte{0x00}
	
	indexKey, err := stub.CreateCompositeKey(indexName, []string{transaction.FromEnterprise, transaction.FromAccount, transaction.TransactionDate, transaction.TransactionID})
	if err != nil {
		return err
	}
	err = stub.PutState(indexKey, indexValue)
	if err != nil {
		return err
	}

	if transaction.FromEnterprise != transaction.ToEnterprise || transaction.FromAccount != transaction.ToAccount{
		indexKey, err := stub.CreateCompositeKey(indexName, []string{transaction.ToEnterprise, transaction.ToAccount, transaction.TransactionDate, transaction.TransactionID})
		if err != nil {
			return err
		}
		
		err = stub.PutState(indexKey, indexValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func now() time.Time{
	loc, _:= time.LoadLocation("Asia/Chongqing")
	return time.Now().In(loc)
}
// snowflake
// These constants are the bit lengths of Sonyflake ID parts.
const (
	BitLenTime      = 39                               // bit length of time
	BitLenSequence  = 8                                // bit length of sequence number
	BitLenMachineID = 63 - BitLenTime - BitLenSequence // bit length of machine id
)

// Settings configures Sonyflake:
//
// StartTime is the time since which the Sonyflake time is defined as the elapsed time.
// If StartTime is 0, the start time of the Sonyflake is set to "2014-09-01 00:00:00 +0000 UTC".
// If StartTime is ahead of the current time, Sonyflake is not created.
//
// MachineID returns the unique ID of the Sonyflake instance.
// If MachineID returns an error, Sonyflake is not created.
// If MachineID is nil, default MachineID is used.
// Default MachineID returns the lower 16 bits of the private IP address.
//
// CheckMachineID validates the uniqueness of the machine ID.
// If CheckMachineID returns false, Sonyflake is not created.
// If CheckMachineID is nil, no validation is done.
type Settings struct {
	StartTime      time.Time
	MachineID      func() (uint16, error)
	CheckMachineID func(uint16) bool
}

// Sonyflake is a distributed unique ID generator.
type Sonyflake struct {
	mutex       *sync.Mutex
	startTime   int64
	elapsedTime int64
	sequence    uint16
	machineID   uint16
}

// NewSonyflake returns a new Sonyflake configured with the given Settings.
// NewSonyflake returns nil in the following cases:
// - Settings.StartTime is ahead of the current time.
// - Settings.MachineID returns an error.
// - Settings.CheckMachineID returns false.
func NewSonyflake(st Settings) *Sonyflake {
	sf := new(Sonyflake)
	sf.mutex = new(sync.Mutex)
	sf.sequence = uint16(1<<BitLenSequence - 1)

	if st.StartTime.After(now()) {
		return nil
	}
	if st.StartTime.IsZero() {
		sf.startTime = toSonyflakeTime(time.Date(2014, 9, 1, 0, 0, 0, 0, time.UTC))
	} else {
		sf.startTime = toSonyflakeTime(st.StartTime)
	}

	var err error
	if st.MachineID == nil {
		sf.machineID, err = lower16BitPrivateIP()
	} else {
		sf.machineID, err = st.MachineID()
	}
	if err != nil || (st.CheckMachineID != nil && !st.CheckMachineID(sf.machineID)) {
		return nil
	}
	return sf
}

// NextID generates a next unique ID.
// After the Sonyflake time overflows, NextID returns an error.
func (sf *Sonyflake) NextID() (uint64, error) {
	const maskSequence = uint16(1<<BitLenSequence - 1)

	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	current := currentElapsedTime(sf.startTime)
	if sf.elapsedTime < current {
		sf.elapsedTime = current
		sf.sequence = 0
	} else { // sf.elapsedTime >= current
		sf.sequence = (sf.sequence + 1) & maskSequence
		if sf.sequence == 0 {
			sf.elapsedTime++
			overtime := sf.elapsedTime - current
			time.Sleep(sleepTime((overtime)))
		}
	}

	return sf.toID()
}

const sonyflakeTimeUnit = 1e7 // nsec, i.e. 10 msec

func toSonyflakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / sonyflakeTimeUnit
}

func currentElapsedTime(startTime int64) int64 {
	return toSonyflakeTime(now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(now().UTC().UnixNano()%sonyflakeTimeUnit)*time.Nanosecond
}

func (sf *Sonyflake) toID() (uint64, error) {
	if sf.elapsedTime >= 1<<BitLenTime {
		return 0, errors.New("over the time limit")
	}

	return uint64(sf.elapsedTime)<<(BitLenSequence+BitLenMachineID) |
		uint64(sf.sequence)<<BitLenMachineID |
		uint64(sf.machineID), nil
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

// Decompose returns a set of Sonyflake ID parts.
func Decompose(id uint64) map[string]uint64 {
	const maskSequence = uint64((1<<BitLenSequence - 1) << BitLenMachineID)
	const maskMachineID = uint64(1<<BitLenMachineID - 1)

	msb := id >> 63
	time := id >> (BitLenSequence + BitLenMachineID)
	sequence := id & maskSequence >> BitLenMachineID
	machineID := id & maskMachineID
	return map[string]uint64{
		"id":         id,
		"msb":        msb,
		"time":       time,
		"sequence":   sequence,
		"machine-id": machineID,
	}
}