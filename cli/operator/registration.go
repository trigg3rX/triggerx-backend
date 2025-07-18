package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func (o *Operator) RegisterOperatorWithChain() error {
	flag, err := o.avsReader.IsOperator(&bind.CallOpts{}, o.operatorAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec IsOperator", "err", err)
		return err
	}
	if !flag {
		o.logger.Info("Operator is not registered with chain.")
		return fmt.Errorf("operator is not registered with chain: %s. Please register your operator with the chain first", o.operatorAddr.String())
	}

	o.logger.Info("Operator is already registered with chain", "operatorAddr", o.operatorAddr.String())
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
