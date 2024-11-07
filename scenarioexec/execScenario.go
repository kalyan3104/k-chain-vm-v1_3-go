package scenarioexec

import (
	vmi "github.com/kalyan3104/k-chain-vm-common-go"
	mc "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/controller"
	fr "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/fileresolver"
	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
)

// Reset clears state/world.
// Is called in RunAllJSONScenariosInDirectory, but not in RunSingleJSONScenario.
func (ae *VMTestExecutor) Reset() {
	ae.World.Clear()
}

// ExecuteScenario executes an individual test.
func (ae *VMTestExecutor) ExecuteScenario(scenario *mj.Scenario, fileResolver fr.FileResolver) error {
	ae.fileResolver = fileResolver
	ae.checkGas = scenario.CheckGas
	err := ae.SetScenariosGasSchedule(scenario.GasSchedule)
	if err != nil {
		return err
	}

	txIndex := 0
	for _, generalStep := range scenario.Steps {
		err := ae.ExecuteStep(generalStep)
		if err != nil {
			return err
		}

		txIndex++
	}

	return nil
}

// ExecuteStep executes an individual step from a scenario.
func (ae *VMTestExecutor) ExecuteStep(generalStep mj.Step) error {
	err := error(nil)

	switch step := generalStep.(type) {
	case *mj.ExternalStepsStep:
		err = ae.ExecuteExternalStep(step)
	case *mj.SetStateStep:
		err = ae.ExecuteSetStateStep(step)
	case *mj.CheckStateStep:
		err = ae.ExecuteCheckStateStep(step)
	case *mj.TxStep:
		_, err = ae.ExecuteTxStep(step)
	case *mj.DumpStateStep:
		err = ae.DumpWorld()
	}

	return err
}

// ExecuteExternalStep executes an external step referenced by the scenario.
func (ae *VMTestExecutor) ExecuteExternalStep(step *mj.ExternalStepsStep) error {
	log.Trace("ExternalStepsStep", "path", step.Path)
	if len(step.Comment) > 0 {
		log.Trace("ExternalStepsStep", "comment", step.Comment)
	}

	fileResolverBackup := ae.fileResolver
	clonedFileResolver := ae.fileResolver.Clone()
	externalStepsRunner := mc.NewScenarioRunner(ae, clonedFileResolver)

	extAbsPth := ae.fileResolver.ResolveAbsolutePath(step.Path)
	err := externalStepsRunner.RunSingleJSONScenario(extAbsPth)
	if err != nil {
		return err
	}

	ae.fileResolver = fileResolverBackup

	return nil
}

// ExecuteSetStateStep executes a SetStateStep.
func (ae *VMTestExecutor) ExecuteSetStateStep(step *mj.SetStateStep) error {
	if len(step.Comment) > 0 {
		log.Trace("SetStateStep", "comment", step.Comment)
	}

	// append accounts
	for _, scenAccount := range step.Accounts {
		worldAccount, err := convertAccount(scenAccount, ae.World)
		if err != nil {
			return err
		}
		err = validateSetStateAccount(scenAccount, worldAccount)
		if err != nil {
			return err
		}

		ae.World.AcctMap.PutAccount(worldAccount)
	}

	// replace block info
	ae.World.PreviousBlockInfo = convertBlockInfo(step.PreviousBlockInfo)
	ae.World.CurrentBlockInfo = convertBlockInfo(step.CurrentBlockInfo)
	ae.World.Blockhashes = mj.JSONBytesFromStringValues(step.BlockHashes)

	// append NewAddressMocks
	err := validateNewAddressMocks(step.NewAddressMocks)
	if err != nil {
		return err
	}
	addressMocksToAdd := convertNewAddressMocks(step.NewAddressMocks)
	ae.World.NewAddressMocks = append(ae.World.NewAddressMocks, addressMocksToAdd...)

	return nil
}

// ExecuteTxStep executes a TxStep.
func (ae *VMTestExecutor) ExecuteTxStep(step *mj.TxStep) (*vmi.VMOutput, error) {
	log.Trace("ExecuteTxStep", "id", step.TxIdent)
	if len(step.Comment) > 0 {
		log.Trace("ExecuteTxStep", "comment", step.Comment)
	}

	output, err := ae.executeTx(step.TxIdent, step.Tx)
	if err != nil {
		return nil, err
	}

	// check results
	if step.ExpectedResult != nil {
		err = ae.checkTxResults(step.TxIdent, step.ExpectedResult, ae.checkGas, output)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}
