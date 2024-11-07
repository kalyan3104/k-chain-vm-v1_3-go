package dex

import (
	"errors"
	"fmt"
	"io/ioutil"

	vmi "github.com/kalyan3104/k-chain-vm-common-go"
	worldhook "github.com/kalyan3104/k-chain-vm-v1_3-go/mock/world"
	am "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarioexec"
	fr "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/fileresolver"
	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
	mjparse "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/parse"
	mjwrite "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/write"
)

type fuzzDexExecutorInitArgs struct {
	wrewaTokenId            string
	mexTokenId              string
	busdTokenId             string
	wemeLpTokenId           string
	webuLpTokenId           string
	wemeFarmTokenId         string
	webuFarmTokenId         string
	mexFarmTokenId          string
	numUsers                int
	numEvents               int
	removeLiquidityProb     int
	addLiquidityProb        int
	swapProb                int
	queryPairsProb          int
	enterFarmProb           int
	exitFarmProb            int
	claimRewardsProb        int
	increaseBlockNonceProb  int
	removeLiquidityMaxValue int
	addLiquidityMaxValue    int
	swapMaxValue            int
	enterFarmMaxValue       int
	exitFarmMaxValue        int
	claimRewardsMaxValue    int
	blockNonceIncrease      int
}

type SwapPair struct {
	firstToken  string
	secondToken string
	lpToken     string
	address     string
}

type Farm struct {
	farmingToken string
	farmToken    string
	rewardToken  string
	address      string
}

type FarmerInfo struct {
	user  string
	value int64
	farm  Farm
	rps   string
}

type fuzzDexExecutor struct {
	vmTestExecutor *am.VMTestExecutor
	world          *worldhook.MockWorld
	vm             vmi.VMExecutionHandler
	parser         mjparse.Parser
	txIndex        int

	wrewaTokenId            string
	mexTokenId              string
	busdTokenId             string
	wemeLpTokenId           string
	webuLpTokenId           string
	wemeFarmTokenId         string
	webuFarmTokenId         string
	mexFarmTokenId          string
	ownerAddress            string
	wemeFarmAddress         string
	webuFarmAddress         string
	mexFarmAddress          string
	wemeSwapAddress         string
	webuSwapAddress         string
	numUsers                int
	numTokens               int
	numEvents               int
	removeLiquidityProb     int
	addLiquidityProb        int
	swapProb                int
	queryPairsProb          int
	enterFarmProb           int
	exitFarmProb            int
	claimRewardsProb        int
	increaseBlockNonceProb  int
	removeLiquidityMaxValue int
	addLiquidityMaxValue    int
	swapMaxValue            int
	enterFarmMaxValue       int
	exitFarmMaxValue        int
	claimRewardsMaxValue    int
	blockNonceIncrease      int
	tokensCheckFrequency    int
	currentFarmTokenNonce   map[string]int
	farmers                 map[int]FarmerInfo
	generatedScenario       *mj.Scenario
	farms                   [3]Farm
	swaps                   [2]SwapPair
}

type eventsStatistics struct {
	swapFixedInputHits   int
	swapFixedInputMisses int

	swapFixedOutputHits   int
	swapFixedOutputMisses int

	addLiquidityHits        int
	addLiquidityMisses      int
	addLiquidityPriceChecks int

	removeLiquidityHits        int
	removeLiquidityMisses      int
	removeLiquidityPriceChecks int

	queryPairsHits   int
	queryPairsMisses int

	enterFarmHits   int
	enterFarmMisses int

	exitFarmHits        int
	exitFarmMisses      int
	exitFarmWithRewards int

	claimRewardsHits        int
	claimRewardsMisses      int
	claimRewardsWithRewards int
}

func newFuzzDexExecutor(fileResolver fr.FileResolver) (*fuzzDexExecutor, error) {
	vmTestExecutor, err := am.NewVMTestExecutor()
	if err != nil {
		return nil, err
	}

	parser := mjparse.NewParser(fileResolver)

	return &fuzzDexExecutor{
		vmTestExecutor: vmTestExecutor,
		world:          vmTestExecutor.World,
		vm:             vmTestExecutor.GetVM(),
		parser:         parser,
		txIndex:        0,
		generatedScenario: &mj.Scenario{
			Name: "fuzz generated",
		},
	}, nil
}

func (pfe *fuzzDexExecutor) saveGeneratedScenario() {
	serialized := mjwrite.ScenarioToJSONString(pfe.generatedScenario)

	err := ioutil.WriteFile("fuzz_gen.scen.json", []byte(serialized), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func (pfe *fuzzDexExecutor) executeStep(stepSnippet string) error {
	step, err := pfe.parser.ParseScenarioStep(stepSnippet)
	if err != nil {
		return err
	}

	pfe.addStep(step)
	return pfe.vmTestExecutor.ExecuteStep(step)
}

func (pfe *fuzzDexExecutor) addStep(step mj.Step) {
	pfe.generatedScenario.Steps = append(pfe.generatedScenario.Steps, step)
}

func (pfe *fuzzDexExecutor) executeTxStep(stepSnippet string) (*vmi.VMOutput, error) {
	step, err := pfe.parser.ParseScenarioStep(stepSnippet)
	if err != nil {
		return nil, err
	}

	txStep, isTx := step.(*mj.TxStep)
	if !isTx {
		return nil, errors.New("tx step expected")
	}

	pfe.addStep(step)

	return pfe.vmTestExecutor.ExecuteTxStep(txStep)
}

func (pfe *fuzzDexExecutor) log(info string, args ...interface{}) {
	fmt.Printf(info+"\n", args...)
}

func (pfe *fuzzDexExecutor) userAddress(userIndex int) string {
	return fmt.Sprintf("address:user%06d", userIndex)
}

func (pfe *fuzzDexExecutor) fullOfDcdtWalletString() string {
	dcdtString := ""

	dcdtString += fmt.Sprintf(`
						"str:%s": "1,000,000,000,000,000,000,000,000,000,000",`, pfe.wrewaTokenId)
	dcdtString += fmt.Sprintf(`
						"str:%s": "1,000,000,000,000,000,000,000,000,000,000",`, pfe.mexTokenId)
	dcdtString += fmt.Sprintf(`
						"str:%s": "1,000,000,000,000,000,000,000,000,000,000",`, pfe.busdTokenId)
	dcdtString += fmt.Sprintf(`
						"str:%s": "1,000,000,000,000,000,000,000,000,000,000",`, pfe.wemeLpTokenId)
	dcdtString += fmt.Sprintf(`
						"str:%s": "1,000,000,000,000,000,000,000,000,000,000"`, pfe.webuLpTokenId)

	return dcdtString
}

func (pfe *fuzzDexExecutor) querySingleResult(from, to, funcName, args string) ([][]byte, error) {
	output, err := pfe.executeTxStep(fmt.Sprintf(`
	{
		"step": "scCall",
		"txId": "%s",
		"tx": {
			"from": "%s",
			"to": "%s",
			"value": "0",
			"function": "%s",
			"arguments": [
				%s
			],
			"gasLimit": "10,000,000",
			"gasPrice": "0"
		},
		"expect": {
			"out": [ "*" ],
			"status": "",
			"logs": [],
			"gas": "*",
			"refund": "*"
		}
	}`,
		funcName,
		from,
		to,
		funcName,
		args,
	))
	if err != nil {
		return [][]byte{}, err
	}

	return output.ReturnData, nil
}

func (pfe *fuzzDexExecutor) querySingleResultStringAddr(from string, to string, funcName string, args string) ([][]byte, error) {
	output, err := pfe.executeTxStep(fmt.Sprintf(`
	{
		"step": "scCall",
		"txId": "%s",
		"tx": {
			"from": "%s",
			"to": "%s",
			"value": "0",
			"function": "%s",
			"arguments": [
				%s
			],
			"gasLimit": "10,000,000",
			"gasPrice": "0"
		},
		"expect": {
			"out": [ "*" ],
			"status": "",
			"logs": [],
			"gas": "*",
			"refund": "*"
		}
	}`,
		funcName,
		from,
		to,
		funcName,
		args,
	))
	if err != nil {
		return [][]byte{}, err
	}

	return output.ReturnData, nil
}

func (pfe *fuzzDexExecutor) increaseBlockNonce(epochDelta int) error {
	currentBlockNonce := uint64(0)
	if pfe.world.CurrentBlockInfo != nil {
		currentBlockNonce = pfe.world.CurrentBlockInfo.BlockNonce
	}

	err := pfe.executeStep(fmt.Sprintf(`
	{
		"step": "setState",
		"comment": "%d - increase block nonce",
		"currentBlockInfo": {
			"blockNonce": "%d"
		}
	}`,
		pfe.nextTxIndex(),
		currentBlockNonce+uint64(epochDelta),
	))
	if err != nil {
		return err
	}

	return nil
}

func (pfe *fuzzDexExecutor) nextTxIndex() int {
	pfe.txIndex++
	return pfe.txIndex
}
