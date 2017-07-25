package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
	Auditrail    map[string][]string `json:"audittrail"`
}

type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "getMydocs":
		return t.getMydocs(stub, args)
	case "getSharedDocs":
		return t.getMydocs(stub, args)

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
	case "revokeAccess":
		return t.revokeAccess(stub, args)
	case "removeDocument":
		return t.removeDocument(stub, args)
	}
	return nil, nil
}

func (t *SimpleChaincode) removeDocument(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) < 2 {
		fmt.Println("Expecting a minimum of three arguments Argument")
		return nil, errors.New("Expected at least one arguments for adding a user")
	}

	var userhash = args[0]
	var dochash = args[1]

	user, err = readFromBlockchain(userhash)
	if err != nil {
		return nil, errors.New("failed to read", err)
	}

}

func (t *SimpleChaincode) revokeAccess(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) < 3 {
		fmt.Println("Expecting a minimum of three arguments Argument")
		return nil, errors.New("Expected at least one arguments for adding a user")
	}

	var userhash = args[0]
	var orghash = args[1]
	var dochash = args[2]

	var user User
	var org User

	//checking if the user exists
	userbytes, err := stub.GetState(userhash)
	if err != nil {
		fmt.Println("could not fetch user", err)
		return nil, err
	}

	err = json.Unmarshal(userbytes, &user)
	if err != nil {
		fmt.Println("Unable to marshal data", err)
		return nil, err
	}

	//checking if the user exists
	orgbytes, orgerr := stub.GetState(orghash)
	if orgerr != nil {
		fmt.Println("could not fetch user", orgerr)
		return nil, err
	}

	err = json.Unmarshal(orgbytes, &org)
	if err != nil {
		fmt.Println("Unable to marshal data", err)
		return nil, err
	}

	userDocsArray := org.SharedwithMe[userhash]

	// removes that particular document from the array
	for i, v := range userDocsArray {
		if v == dochash {
			userDocsArray = append(userDocsArray[:i], userDocsArray[i+1:]...)
			break
		}
	}

	//assign that array to the user map key
	org.SharedwithMe[userhash] = userDocsArray

	bytesvalue, err := json.Marshal(&org)
	if err != nil {
		fmt.Println("Could not marshal personal info object", err)
		return nil, err
	}

	//write back in blockchain
	err = stub.PutState(userhash, bytesvalue)
	if err != nil {
		fmt.Println("Could not save add doc to user", err)
		return nil, err
	}

	fmt.Println("Successfully revoked access to the doc")
	return nil, nil

}
func (t *SimpleChaincode) createUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//func createUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering createUser")

	if len(args) < 1 {
		fmt.Println("Expecting One Argument")
		return nil, errors.New("Expected at least one arguments for adding a user")
	}

	var userid = args[0]
	var userinfo = `{"owns":[],"mymap":{}, "audit":{}}`

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
	fmt.Println("Entering addDocument")
	var user User
	if len(args) < 2 {
		fmt.Println("Expecting two Argument")
		return nil, errors.New("Expected at least two arguments for adding a document")
	}

	var userid = args[0]
	fmt.Println(userid)
	var docid = args[1]
	fmt.Println(docid)
	bytes, err := stub.GetState(userid)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &user)
	if err != nil {
		fmt.Println("unable to unmarshal user data")
		return nil, err
	}

	user.Owns = append(user.Owns, docid)

	_, err = writeIntoBlockchain(userid, user, stub)
	if err != nil {
		fmt.Println("Could not save add doc to user", err)
		return nil, err
	}

	fmt.Println("Successfully added the doc")
	return nil, nil

}

//3. shareDocument()    (#doc,#user, #org)  Invoke
func (t *SimpleChaincode) shareDocument(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering shareDocument")
	var user User
	var org User
	//	var doc DocumentInfo
	//fmt.Println(doc)
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

	if user.Auditrail == nil {
		user.Auditrail = make(map[string][]string)

	}
	//adding the document if it doesnt exists already
	if !contains(org.SharedwithMe[userid], docid) {
		timestamp := makeTimestamp()
		fmt.Println(timestamp)
		//---------------Sharing the doc to Organisation-----------------------
		org.SharedwithMe[userid] = append(org.SharedwithMe[userid], docid)

		//-------------- Adding time stamp to user audit trail array-------------
		user.Auditrail[orgid] = append(user.Auditrail[orgid], timestamp)
		user.Auditrail[orgid] = append(user.Auditrail[orgid], docid)
	}

	_, err = writeIntoBlockchain(orgid, org, stub)
	if err != nil {
		fmt.Println("Could not save org Info", err)
		return nil, err
	}

	_, err = writeIntoBlockchain(userid, user, stub)
	if err != nil {
		fmt.Println("Could not save user Info", err)
		return nil, err
	}

	fmt.Println("Successfully shared the doc")
	return nil, nil

}

//4. getMydocs()    (#user) Query
func (t *SimpleChaincode) getMydocs(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering get my docs")

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
func makeTimestamp() string {
	t := time.Now()

	return t.Format(("2006-01-02T15:04:05.999999-07:00"))
	//return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

//------------- reusable methods -------------------

func (t *SimpleChaincode) writeIntoBlockchain(key string, value User, stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytes, err := json.Marshal(&value)
	if err != nil {
		fmt.Println("Could not marshal info object", err)
		return nil, err
	}

	err = stub.PutState(userid, bytes)
	if err != nil {
		fmt.Println("Could not save sharing info to org", err)
		return nil, err
	}

	return nil, nil
}

func (t *SimpleChaincode) readFromBlockchain(key string) (User, error) {
	userbytes, err := stub.GetState(key)
	if err != nil {
		fmt.Println("could not fetch user", err)
		return nil, err
	}

	var user User
	err = json.Unmarshal(userbytes, &user)
	if err != nil {
		fmt.Println("Unable to marshal data", err)
		return nil, err
	}

	return user, nil
}
