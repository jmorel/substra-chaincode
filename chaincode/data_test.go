package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func TestDataManager(t *testing.T) {
	scc := new(SubstraChaincode)
	mockStub := shim.NewMockStub("substra", scc)

	// Add dataManager with invalid field
	inpDataManager := inputDataManager{
		OpenerHash: "aaa",
	}
	args := inpDataManager.createSample()
	resp := mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 500 {
		t.Errorf("when adding dataManager with invalid opener hash, status %d and message %s", status, resp.Message)
	}
	// Properly add dataManager
	err, resp, tt := registerItem(*mockStub, "dataManager")
	if err != nil {
		t.Errorf(err.Error())
	}
	inpDataManager = tt.(inputDataManager)
	dataManagerKey := string(resp.Payload)
	// check returned dataManager key corresponds to opener hash
	if dataManagerKey != dataManagerOpenerHash {
		t.Errorf("when adding dataManager: dataManager key does not correspond to dataManager opener hash: %s - %s", dataManagerKey, dataManagerOpenerHash)
	}
	// Add dataManager which already exist
	resp = mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 500 {
		t.Errorf("when adding dataManager which already exists, status %d and message %s", status, resp.Message)
	}
	// Query dataManager and check fields match expectations
	args = [][]byte{[]byte("queryDataManager"), []byte(dataManagerKey)}
	resp = mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 200 {
		t.Errorf("when querying the dataManager, status %d and message %s", status, resp.Message)
	}
	dataManager := outputDataManager{}
	err = bytesToStruct(resp.Payload, &dataManager)
	assert.NoError(t, err, "when unmarshalling queried dataManager")
	expectedDataManager := outputDataManager{
		ObjectiveKey: inpDataManager.ObjectiveKey,
		Key:          dataManagerKey,
		Owner:        "bbd157aa8e85eb985aeedb79361cd45739c92494dce44d351fd2dbd6190e27f0",
		Name:         inpDataManager.Name,
		Description: &HashDress{
			StorageAddress: inpDataManager.DescriptionStorageAddress,
			Hash:           inpDataManager.DescriptionHash,
		},
		Permissions: inpDataManager.Permissions,
		Opener: HashDress{
			Hash:           dataManagerKey,
			StorageAddress: inpDataManager.OpenerStorageAddress,
		},
		Type: inpDataManager.Type,
	}
	assert.Exactly(t, expectedDataManager, dataManager)

	// Query all dataManagers and check fields match expectations
	args = [][]byte{[]byte("queryDataManagers")}
	resp = mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 200 {
		t.Errorf("when querying dataManagers - status %d and message %s", status, resp.Message)
	}
	var dataManagers []outputDataManager
	err = json.Unmarshal(resp.Payload, &dataManagers)
	assert.NoError(t, err, "while unmarshalling dataManagers")
	assert.Len(t, dataManagers, 1)
	assert.Exactly(t, expectedDataManager, dataManagers[0], "return objective different from registered one")

	args = [][]byte{[]byte("queryDataset"), []byte(inpDataManager.OpenerHash)}
	resp = mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 200 {
		t.Errorf("when querying Dataset, status %d and message %s", status, resp.Message)
	}
	if !strings.Contains(string(resp.Payload), "\"trainDataSampleKeys\":[]") {
		t.Errorf("when querying Dataset, trainDataSampleKeys should be []")
	}
	if !strings.Contains(string(resp.Payload), "\"testDataSampleKeys\":[]") {
		t.Errorf("when querying Dataset, testDataSampleKeys should be []")
	}
}

func TestGetTestDatasetKeys(t *testing.T) {
	scc := new(SubstraChaincode)
	mockStub := shim.NewMockStub("substra", scc)

	// Input DataManager
	inpDataManager := inputDataManager{}
	args := inpDataManager.createSample()
	mockStub.MockInvoke("42", args)

	// Add both train and test dataSample
	inpDataSample := inputDataSample{Hashes: testDataSampleHash1}
	args = inpDataSample.createSample()
	mockStub.MockInvoke("42", args)
	inpDataSample.Hashes = testDataSampleHash2
	inpDataSample.TestOnly = "true"
	args = inpDataSample.createSample()
	mockStub.MockInvoke("42", args)

	// Querry the DataManager
	args = [][]byte{[]byte("queryDataset"), []byte(inpDataManager.OpenerHash)}
	resp := mockStub.MockInvoke("42", args)
	assert.EqualValues(t, 200, resp.Status, "querrying the dataManager should return an ok status")
	payload := map[string]interface{}{}
	err := json.Unmarshal(resp.Payload, &payload)
	assert.NoError(t, err)

	v, ok := payload["testDataSampleKeys"]
	assert.True(t, ok, "payload should contains the test dataSample keys")
	assert.Contains(t, v, testDataSampleHash2, "testDataSampleKeys should contain the test dataSampleHash")
	assert.NotContains(t, v, testDataSampleHash1, "testDataSampleKeys should not contains the train dataSampleHash")
}
func TestDataset(t *testing.T) {
	scc := new(SubstraChaincode)
	mockStub := shim.NewMockStub("substra", scc)

	// Add dataSample with invalid field
	inpDataSample := inputDataSample{
		Hashes: "aaa",
	}
	args := inpDataSample.createSample()
	resp := mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 500 {
		t.Errorf("when adding dataSample with invalid hash, status %d and message %s", status, resp.Message)
	}

	// Add dataSample with unexiting dataManager
	inpDataSample = inputDataSample{}
	args = inpDataSample.createSample()
	resp = mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 500 {
		t.Errorf("when adding dataSample with unexisting dataManager, status %d and message %s", status, resp.Message)
	}
	// TODO Would be nice to check failure when adding dataSample to a dataManager owned by a different people

	// Properly add dataSample
	// 1. add associated dataManager
	inpDataManager := inputDataManager{}
	args = inpDataManager.createSample()
	mockStub.MockInvoke("42", args)
	// 2. add dataSample
	inpDataSample = inputDataSample{}
	args = inpDataSample.createSample()
	resp = mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 200 {
		t.Errorf("when adding dataSample, status %d and message %s", status, resp.Message)
	}
	// check payload correspond to input dataSample keys
	dataSampleKeys := string(resp.Payload)
	if expectedResp := "{\"keys\": [\"" + strings.Replace(inpDataSample.Hashes, ", ", "\", \"", -1) + "\"]}"; dataSampleKeys != expectedResp {
		t.Errorf("when adding dataSample: dataSample keys does not correspond to dataSample hashes: %s - %s", dataSampleKeys, expectedResp)
	}

	// Add dataSample which already exist
	resp = mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 500 {
		t.Errorf("when adding dataSample which already exist, status %d and message %s", status, resp.Message)
	}

	/**
	// Query dataSample and check it corresponds to what was input
	args = [][]byte{[]byte("queryDataset"), []byte(inpDataManager.OpenerHash)}
	resp = mockStub.MockInvoke("42", args)
	if status := resp.Status; status != 200 {
		t.Errorf("when querying dataManager dataSample with status %d and message %s", status, resp.Message)
	}
	payload := make(map[string]interface{})
	json.Unmarshal(resp.Payload, &payload)
	if _, ok := payload["key"]; !ok {
		t.Errorf("when querying dataManager dataSample, payload should contain the dataManager key")
	}
	v, ok := payload["trainDatasetKeys"]
	if !ok {
		t.Errorf("when querying dataManager dataSample, payload should contain the train dataSample keys")
	}
	if reflect.DeepEqual(v, strings.Split(strings.Replace(inpDataSample.Hashes, " ", "", -1), ",")) {
		t.Errorf("when querying dataManager dataSample, unexpected train keys")
	}
	**/
}
