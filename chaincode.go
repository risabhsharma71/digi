package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type DocumentInfo struct {
	Owner string   `json:"owner"`
	Docs  []string `json:"docs"`
}
type User struct {
	Owns []string `json:"owns"`
	//SharedwithMe []DocumentInfo `json:"sharedwithme"`
	SharedwithMe map[string][]string `json:"sharedwithme"`
}

type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "getMydocs":
		return t.createUser(stub, args)
	case "getSharedDocs":
		return t.addDocument(stub, args)

	}
	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	switch function {
	case "createUser":
		return t.createUser(stub, args)
	case "addDocument":
		return t.addDocument(stub, args)
	case "shareDocument":
		return t.shareDocument(stub, args)
	}
	return nil, nil
}
func (t *SimpleChaincode) createUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//func createUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering CreateLoanApplication")

	if len(args) < 1 {
		fmt.Println("Expecting One Argument")
		return nil, errors.New("Expected at least one arguments for adding a user")
	}

	var userid = args[0]
	var userinfo = `{[],{}}`

	err := stub.PutState(userid, []byte(userinfo))
	if err != nil {
		fmt.Println("Could not save user to ledger", err)
		return nil, err
	}

	fmt.Println("Successfully saved user/org")
	return nil, nil
}

//2.addDocument()   (#user,#doc)
func (t *SimpleChaincode) addDocument(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering CreateLoanApplication")
	var user User
	if len(args) < 2 {
		fmt.Println("Expecting two Argument")
		return nil, errors.New("Expected at least two arguments for adding a document")
	}

	var userid = args[0]
	var docid = args[1]
	bytes, err := stub.GetState(userid)
	if err != nil {
		//	fmt.Println("Could not fetch loan application with id "+loanApplicationId+" from ledger", err)
		return nil, err
	}

	err = json.Unmarshal(bytes, &user)
	if err != nil {
		fmt.Println("unable to unmarshal user data")
		return nil, err
	}

	user.Owns = append(user.Owns, docid)
	bytesvalue, err := json.Marshal(&user)
	if err != nil {
		fmt.Println("Could not marshal personal info object", err)
		return nil, err
	}

	err = stub.PutState(userid, bytesvalue)
	if err != nil {
		fmt.Println("Could not save add doc to user", err)
		return nil, err
	}

	fmt.Println("Successfully added the doc")
	return nil, nil

}

//3. shareDocument()    (#doc,#user, #org)  Invoke
func (t *SimpleChaincode) shareDocument(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering CreateLoanApplication")
	var user User
	var org User
	var doc DocumentInfo
	fmt.Println(doc)
	if len(args) < 2 {
		fmt.Println("Expecting three Argument")
		return nil, errors.New("Expected at least three arguments for sharing  a document")
	}

	var userid = args[0]
	var docid = args[1]
	var orgid = args[2]
	//fetching the user
	userbytes, err := stub.GetState(userid)
	if err != nil {
		fmt.Println("could not fetch user", err)
		return nil, err
	}
	err = json.Unmarshal(userbytes, &user)
	if err != nil {
		fmt.Println("unable to unmarshal user data")
		return nil, err
	}
	if !contains(user.Owns, docid) {
		fmt.Println("docment doesnt exists")
		return nil, err
	}
	//fetch oraganisation
	orgbytes, err := stub.GetState(orgid)
	if err != nil {
		fmt.Println("could not fetch user", err)
		return nil, err
	}
	err = json.Unmarshal(orgbytes, &org)
	if err != nil {
		fmt.Println("unable to unmarshal org data")
		return nil, err
	}

	if org.SharedwithMe == nil {
		org.SharedwithMe = make(map[string][]string)
	}
	//adding the document if it doesnt exists already
	if !contains(org.SharedwithMe[userid], docid) {
		org.SharedwithMe[userid] = append(org.SharedwithMe[userid], docid)
	}
	bytes, err := json.Marshal(&org)
	if err != nil {
		fmt.Println("Could not marshal personal info object", err)
		return nil, err
	}

	err = stub.PutState(userid, bytes)
	if err != nil {
		fmt.Println("Could not save sharing info to org", err)
		return nil, err
	}

	fmt.Println("Successfully shared the doc")
	return nil, nil

}

//4. getMydocs()    (#user) Query
func (t *SimpleChaincode) getMydocs(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering GetLoanApplication")

	if len(args) < 1 {
		fmt.Println("Invalid number of arguments")
		return nil, errors.New("Missing userid")
	}

	var userid = args[0]
	idasbytes, err := stub.GetState(userid)
	if err != nil {
		fmt.Println("Could not user info", err)
		return nil, err
	}
	return idasbytes, nil
}

//getSharedDocs()
func (t *SimpleChaincode) getSharedDocs(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering GetLoanApplication")

	if len(args) < 1 {
		fmt.Println("Invalid number of arguments")
		return nil, errors.New("Missing userid")
	}

	var userid = args[0]
	bytes, err := stub.GetState(userid)
	if err != nil {
		fmt.Println("Could not user info", err)
		return nil, err
	}
	return bytes, nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))

	if err != nil {
		fmt.Println("Could not start SimpleChaincode")
	} else {
		fmt.Println("SimpleChaincode successfully started")
	}
}
func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
