/*
 SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke marbles ====
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarble","marble2","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarblesBasedOnColor","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["delete","marble1"]}'

// ==== Query marbles ====
// peer chaincode query -C myc1 -n marbles -c '{"Args":["readMarble","marble1"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getMarblesByRange","marble1","marble3"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getHistoryForMarble","marble1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
// peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarblesByOwner","tom"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"owner\":\"tom\"}}"]}'

// Rich Query with Pagination (Only supported if CouchDB is used as state database):
// peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarblesWithPagination","{\"selector\":{\"owner\":\"tom\"}}","3",""]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Each index must be defined in its own
// text file with extension *.json with the index definition formatted in JSON following the
// CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This marbles02 example chaincode demonstrates a packaged
// index which you can find in META-INF/statedb/couchdb/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory, for packaging and deployment
// to managed environments.
//
// In the examples below you can find index definitions that support marbles02
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface or a curl command line utility.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, owner.
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"docType\",\"owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1_marbles/_index
//

// Index for docType, owner, size (descending order).
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"size\":\"desc\"},{\"docType\":\"desc\"},{\"owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1_marbles/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":{\"$eq\":\"marble\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"sort"
	//"strconv"
	"strings"
	//"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type marble struct {
	uniqueID     string  `json:"unq_id"`
	UserIDs      map[string]int  `json:"u_ids"`
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
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	//fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "accessConsent" { //create a new marble
		return t.accessConsent(stub, args)
	} else if function == "queryConsent" { //find consent based on an ad hoc rich query
		return t.queryConsent(stub, args)
	} else if function == "updateConsent" {
		return t.updateConsent(stub, args)
	} else if function == "updateRole" {
		return t.updateRole(stub, args)
	} else if function == "initialize" {
		return t.initialize(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) updateRole(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	// watchdog id, role_id, data consumer id action
	// ==== Input sanitation ====
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	w_id := strings.ToLower(args[0])
	r_id := strings.ToLower(args[1])
	dc_id := strings.ToLower(args[2])
	action := strings.ToLower(args[3])
	var unq_id string
	unq_id = w_id + r_id + dc_id
	marbleAsBytes, err := stub.GetState(unq_id)
	if action == "g" {
		if err != nil {
			return shim.Error("Failed to get marble: " + err.Error())
		} else if marbleAsBytes == nil {
			user_ids := make(map[string]int)
			user_ids[unq_id] = 1
			//fmt.Println("inside1")
			marble := &marble{unq_id, user_ids}
			marbleJSONasBytes, err := json.Marshal(marble)
			if err != nil {
				return shim.Error(err.Error())
			}
			//fmt.Println("inside2")
			// === Save marble to state ===
			err = stub.PutState(unq_id, marbleJSONasBytes)
			if err != nil {
				return shim.Error(err.Error())
			}
		}
		
	} else if action == "r" {
		if err != nil {
			return shim.Error("Failed to get marble: " + err.Error())
		} else if marbleAsBytes != nil {
			err = stub.DelState(unq_id)
			if err != nil {
				return shim.Error("Failed to delete state:" + err.Error())
			}
		}
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) accessConsent(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// role id, start date, end date, column ids, watchdog id, data consumer id
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}
	// ==== Input sanitation ====
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	s_date := strings.ToLower(args[1])
	e_date := strings.ToLower(args[2])
	r_id := strings.ToLower(args[0])
	w_id := strings.ToLower(args[4])
	dc_id := strings.ToLower(args[5])
	user_ids := make(map[string]int)
	count := 0
	ids := strings.Split(args[3], ",")
	var unq_id string
	unq_id = w_id + r_id + dc_id
	consent, err := stub.GetState(unq_id)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if consent != nil {
		for _, c_id := range ids {
			unq_id = c_id + r_id + s_date + e_date + w_id
			marbleAsBytes, err := stub.GetState(unq_id)
			if err != nil {
				return shim.Error("Failed to get marble: " + err.Error())
			} else if marbleAsBytes != nil {
				marbleToTransfer := marble{}
				err = json.Unmarshal(marbleAsBytes, &marbleToTransfer) //unmarshal it aka JSON.parse()
				if err != nil {
					return shim.Error(err.Error())
				}
				user_ids = marbleToTransfer.UserIDs
				// if there are user ids in the value map only then create and store the string list made from them
				if len(user_ids) > 0 {
					count = count + 1
				}
			}
		}
	} else {
		return shim.Error("Watchdog has not approved role given for the data consumer")
	}
	if count  == 0 {
		// display error message even no consent certificate can be given
		return shim.Error("Consent not found")
	}
	//fmt.Println("- end init marble")
	return shim.Success(nil)
}

func contains(s []string, e string) int {
    for i, a := range s {
        if a == e {
            return i
        }
    }
    return -1
}

func remove(s []string, i int) []string {
    s[i] = s[len(s)-1]
    // We do not need to put s[i] at the end, as it will be discarded anyway
    return s[:len(s)-1]
}

func (t *SimpleChaincode) updateConsent(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}
	//patient_id, action, role_id, start date, end date, arr[column ids], watchdog id
	// ==== Input sanitation ====
	//fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	s_date := strings.ToLower(args[3])
	e_date := strings.ToLower(args[4])
	p_id := strings.ToLower(args[0])
	action := strings.ToLower(args[1])
	w_id := strings.ToLower(args[6])
	r_id := strings.ToLower(args[2])
	ids := strings.Split(args[5], ",")
	var unq_id string
	for _, c_id := range ids {
		// TODO: we might not need to store all this extra information, can it make a diffence in performance?
		unq_id = c_id + r_id + s_date + e_date + w_id
		//fmt.Println(unq_id)
		marbleAsBytes, err := stub.GetState(unq_id)
		if err != nil {
			return shim.Error("Failed to get marble: " + err.Error())
		} else if marbleAsBytes != nil {
			// if the marble already exists then fetch the user id array and add the new user id
			marbleToTransfer := marble{}
			err = json.Unmarshal(marbleAsBytes, &marbleToTransfer) //unmarshal it aka JSON.parse()
			if err != nil {
				return shim.Error(err.Error())
			}
			user_ids := marbleToTransfer.UserIDs
			// check if given patientid already exists in the key-value pair
			index := user_ids[p_id]
			changedone := false
			if action == "g" && index == 0 {
				// if action is grant and the patient id is not present then add
				user_ids[p_id] = 1
				changedone = true
			} else if action == "r" && index != 0 {
				// if action is revoke and the patient id is present then delete
				delete(user_ids, p_id)
				changedone = true
				if len(user_ids) == 0 {
					// if the last user id is deleted, then delete that setting
					err = stub.DelState(unq_id)
					if err != nil {
						return shim.Error("Failed to delete state:" + err.Error())
					}
					//fmt.Println("- end init marble")
					return shim.Success(nil)
				}
			}
			// ideally the state should updated in the database only if user_ids are modified as shown above.
			// if changedone is not there we would still be doing a put state even if no real change was made to the resource and this could help reduce collisions
			// but ideally we should also inform the user that no update was made
			if changedone == true {
				marbleToTransfer.UserIDs = user_ids
				marbleJSONasBytes, _ := json.Marshal(marbleToTransfer)
				err = stub.PutState(unq_id, marbleJSONasBytes) //rewrite the marble
				if err != nil {
					return shim.Error(err.Error())
				}
			}
		} else if action == "g" {
			// if a configuration does not exist create one
			//fmt.Println("inside")
			user_ids := make(map[string]int)
			user_ids[p_id] = 1
			//fmt.Println("inside1")
			marble := &marble{unq_id, user_ids}
			marbleJSONasBytes, err := json.Marshal(marble)
			if err != nil {
				return shim.Error(err.Error())
			}
			//fmt.Println("inside2")
			// === Save marble to state ===
			err = stub.PutState(unq_id, marbleJSONasBytes)
			if err != nil {
				return shim.Error(err.Error())
			}
			//fmt.Println("inside3")
		}
	}
	//fmt.Println("- end init marble")
	return shim.Success(nil)
}

func (t *SimpleChaincode) initialize(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}
	//column id, action, role_id, start date, end date, arr[patient ids], watchdog id
	// ==== Input sanitation ====
	//fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	s_date := strings.ToLower(args[3])
	e_date := strings.ToLower(args[4])
	c_id := strings.ToLower(args[0])
	//action := strings.ToLower(args[1])
	w_id := strings.ToLower(args[6])
	r_id := strings.ToLower(args[2])
	ids := strings.Split(args[5], ",")
	var unq_id string
	user_ids := make(map[string]int)
	for _, p_id := range ids {
		user_ids[p_id] = 1
	}
	unq_id = c_id + r_id + s_date + e_date + w_id
	marble := &marble{unq_id, user_ids}
	marbleJSONasBytes, err := json.Marshal(marble)
	if err != nil {
		return shim.Error(err.Error())
	}
	//fmt.Println("inside2")
	// === Save marble to state ===
	err = stub.PutState(unq_id, marbleJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	//fmt.Println("- end init marble")
	return shim.Success(nil)
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

// ===== Example: Ad hoc rich query ========================================================
// queryMarbles uses a query string to perform a query for marbles.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryConsent(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}