package privatetxpool

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

func TestRestrictedTxPool(gt *testing.T) {
	// You need to add 2 IP addresses locally for this test:
	// sudo ifconfig lo0 alias 127.0.0.2 up
	// sudo ifconfig lo0 alias 127.0.0.3 up
	t := devtest.SerialT(gt)
	sys := presets.NewSingleChainThreeNodes(t)
	l := t.Logger()

	l.Info("Confirm that the CL nodes are progressing the unsafe chain")
	delta := uint64(3)
	dsl.CheckAll(t,
		sys.L2CL.AdvancedFn(types.LocalUnsafe, delta, 30),
		sys.L2CLB.AdvancedFn(types.LocalUnsafe, delta, 30),
		sys.L2CLC.AdvancedFn(types.LocalUnsafe, delta, 30),
	)

	l.Info("Stop the L2 batcher")
	sys.L2Batcher.Stop()
}
