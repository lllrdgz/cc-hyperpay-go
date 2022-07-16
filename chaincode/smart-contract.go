package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Account
type SmartContract struct {
	contractapi.Contract
}

// Account describes basic details of what makes up a simple account
//Insert struct field in alphabetic order => to achieve determinism accross languages
// golang keeps the order when marshal to json but doesn't order automatically
type Account struct {
	ID      string  `json:"ID"`
	Balance float32 `json:"Balance float32"`
	Bank    string  `json:"Bank string"`
}

// InitLedger adds a base set of accounts to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	accounts := []Account{
		{ID: "account1", Balance: 100, Bank: "JPMorgan Chase & Co."},
		{ID: "account2", Balance: 200, Bank: "Bank of America Corp."},
		{ID: "account3", Balance: 300, Bank: "JPMorgan Chase & Co."},
		{ID: "account4", Balance: 400, Bank: "Bank of America Corp."},
		{ID: "account5", Balance: 500, Bank: "JPMorgan Chase & Co."},
	}

	// For each account encoding and save it
	for _, account := range accounts {
		accountJSON, err := json.Marshal(account)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(account.ID, accountJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// AccountExists returns true when account with given ID exists in world state
func (s *SmartContract) AccountExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {

	// Get client org id and verify it matches peer org id.
	// In this scenario, client is only authorized to read/write private data from its own peer.
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return false, err
	}
	err = verifyClientOrgMatchesPeerOrg(clientOrgID)
	if err != nil {
		return false, err
	}

	accountJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accountJSON == nil {
		return false, fmt.Errorf("the account %s does not exist", id)
	}

	return accountJSON != nil, nil
}

// ReadAccount returns the account stored in the world state with given id.
func (s *SmartContract) ReadAccount(ctx contractapi.TransactionContextInterface, id string) (*Account, error) {

	// Get client org id and verify it matches peer org id.
	// In this scenario, client is only authorized to read/write private data from its own peer.
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return nil, err
	}
	err = verifyClientOrgMatchesPeerOrg(clientOrgID)
	if err != nil {
		return nil, err
	}

	accountJSON, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accountJSON == nil {
		return nil, fmt.Errorf("the account %s does not exist", id)
	}

	var account Account
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// CreateAccount issues a new account to the world state with given details.
func (s *SmartContract) CreateAccount(ctx contractapi.TransactionContextInterface, id string, balance float32, bank string) error {

	// Get client org id and verify it matches peer org id.
	// In this scenario, client is only authorized to read/write private data from its own peer.
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return err
	}
	err = verifyClientOrgMatchesPeerOrg(clientOrgID)
	if err != nil {
		return err
	}

	exists, err := s.AccountExists(ctx, id)

	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the account %s already exists", id)
	}

	account := Account{
		ID:      id,
		Balance: balance,
		Bank:    bank,
	}

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(account.ID, accountJSON)
	if err != nil {
		return fmt.Errorf("failed to put asset in public data: %v", err)
	}

	// Set the endorsement policy such that an owner org peer is required to endorse future updates
	endorsingOrgs := []string{clientOrgID}
	err = setAssetStateBasedEndorsement(ctx, account.ID, endorsingOrgs)
	if err != nil {
		return fmt.Errorf("failed setting state based endorsement for buyer and seller: %v", err)
	}

	return nil
}

// DeleteAccount deletes an given account from the world state.
func (s *SmartContract) DeleteAccount(ctx contractapi.TransactionContextInterface, id string) error {

	// Get client org id and verify it matches peer org id.
	// In this scenario, client is only authorized to read/write private data from its own peer.
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return err
	}
	err = verifyClientOrgMatchesPeerOrg(clientOrgID)
	if err != nil {
		return err
	}

	exists, err := s.AccountExists(ctx, id)

	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the account %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, fromId, toId string, amount float32) error {
	// @todo q solo pueda hacer esto el duenyo d la cuenta fuente

	// Get client org id and verify it matches peer org id.
	// In this scenario, client is only authorized to read/write private data from its own peer.
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return err
	}
	err = verifyClientOrgMatchesPeerOrg(clientOrgID)
	if err != nil {
		return err
	}

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	fromAcc, err := s.ReadAccount(ctx, fromId)
	if err != nil {
		return errors.New("source account doesn't exist")
	}

	toAcc, err := s.ReadAccount(ctx, toId)
	if err != nil {
		return errors.New("destination account doesn't exist")
	}

	fromAcc.Balance -= amount
	toAcc.Balance += amount

	fromAccJson, err := json.Marshal(fromAcc)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState(fromId, fromAccJson); err != nil {
		return err
	}

	toAccJson, err := json.Marshal(toAcc)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(toId, toAccJson)
}