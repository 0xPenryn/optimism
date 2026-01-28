package pipeline

import (
	"context"
	"fmt"
	"math/big"

	altda "github.com/ethereum-optimism/optimism/op-alt-da"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum/go-ethereum/common"
)

func DeployAltDA(env *Env, intent *state.Intent, st *state.State, chainID common.Hash) error {
	lgr := env.Logger.New("stage", "deploy-alt-da")

	chainIntent, err := intent.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain intent: %w", err)
	}

	chainState, err := st.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain state: %w", err)
	}

	if !shouldDeployAltDA(chainIntent, chainState) {
		lgr.Info("alt-da deployment not needed")
		return nil
	}

	lgr.Info("deploying alt-da contracts")
	input := opcm.DeployAltDAInput{
		Salt:                     st.Create2Salt,
		ProxyAdmin:               chainState.OpChainContracts.OpChainProxyAdminImpl,
		ChallengeContractOwner:   chainIntent.Roles.L1ProxyAdminOwner,
		ChallengeWindow:          new(big.Int).SetUint64(chainIntent.DangerousAltDAConfig.DAChallengeWindow),
		ResolveWindow:            new(big.Int).SetUint64(chainIntent.DangerousAltDAConfig.DAResolveWindow),
		BondSize:                 new(big.Int).SetUint64(chainIntent.DangerousAltDAConfig.DABondSize),
		ResolverRefundPercentage: new(big.Int).SetUint64(chainIntent.DangerousAltDAConfig.DAResolverRefundPercentage),
	}

	var output opcm.DeployAltDAOutput

	if env.UseForge {
		if env.ForgeClient == nil {
			return fmt.Errorf("Forge client is nil but UseForge is enabled")
		}
		if env.Context == nil {
			env.Context = context.Background()
		}
		if env.PrivateKey == "" {
			return fmt.Errorf("private key is required for Forge deployments")
		}
		if env.L1RPCUrl == "" {
			return fmt.Errorf("L1 RPC URL is required for Forge deployments")
		}
		lgr.Info("using Forge for DeployAltDA")
		forgeCaller := opcm.NewDeployAltDAForgeCaller(env.ForgeClient)
		forgeOpts := []string{
			"--rpc-url", env.L1RPCUrl,
			"--broadcast",
			"--private-key", env.PrivateKey,
		}
		output, _, err = forgeCaller(env.Context, input, forgeOpts...)
		if err != nil {
			return fmt.Errorf("failed to deploy alt-da contracts with Forge: %w", err)
		}
	} else {
		deployAltDAScript, err := opcm.NewDeployAltDAScript(env.L1ScriptHost)
		if err != nil {
			return fmt.Errorf("failed to load DeployAltDA script: %w", err)
		}
		output, err = deployAltDAScript.Run(input)
		if err != nil {
			return fmt.Errorf("failed to deploy alt-da contracts: %w", err)
		}
	}

	chainState.OpChainContracts.AltDAChallengeProxy = output.DataAvailabilityChallengeProxy
	chainState.OpChainContracts.AltDAChallengeImpl = output.DataAvailabilityChallengeImpl
	return nil
}

func shouldDeployAltDA(chainIntent *state.ChainIntent, chainState *state.ChainState) bool {
	if !(chainIntent.DangerousAltDAConfig.UseAltDA && chainIntent.DangerousAltDAConfig.DACommitmentType == altda.KeccakCommitmentString) {
		return false
	}

	return chainState.OpChainContracts.AltDAChallengeImpl == common.Address{}
}
