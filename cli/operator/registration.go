package operator

import (
	"context"
	// "encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func (o *Operator) registerOperatorOnStartup() {
	err := o.RegisterOperatorWithChain()
	if err != nil {
		// This error might only be that the operator was already registered with chain, so we don't want to fatal
		o.logger.Error("Error registering operator with chain", "err", err)
	} else {
		o.logger.Infof("Registered operator with chain")
	}

	err = o.RegisterOperatorWithAvs()
	if err != nil {
		o.logger.Fatal("Error registering operator with avs", "err", err)
	}
}
func (o *Operator) RegisterOperatorWithChain() error {
	flag, err := o.avsReader.IsOperator(&bind.CallOpts{}, o.operatorAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec IsOperator", "err", err)
		return err
	}
	if !flag {
		o.logger.Info("Operator is not registered.")
		panic(fmt.Sprintf("Operator is not registered: %s", o.operatorAddr.String()))

	}
	return nil
}

// RegisterOperatorWithAvs Registration specific functions
func (o *Operator) RegisterOperatorWithAvs() error {
	operators, err := o.avsReader.GetOptInOperators(&bind.CallOpts{}, o.avsAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec IsOperator", "err", err)
		return err
	}
	found := false
	for _, addr := range operators {
		if addr == o.operatorAddr {
			found = true
			break
		}
	}

	if !found {
		o.logger.Info("Operator is not opt-in this avs.")
		_, err = o.avsWriter.RegisterOperatorToAVS(context.Background())
		if err != nil {
			o.logger.Error("Avs failed to RegisterOperatorToAVS", "err", err)
			return err
		}
	}

	o.logger.Info("Operator has opt-in this avs:", "avsAddr", o.avsAddr.String())

	return nil
}

