// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

/**
 * SDM.sol (stateless, sequencer-gated)
 *
 * Payload ABI encoding:
 *   payload = abi.encode(TxWithGas[] entries)
 *
 * TxWithGas:
 *   - uint32 txIndex
 *   - bytes32 txHash
 *   - uint64 opgas
 *
 * No storage of block contents is used (only an optional event emission).
 */
contract SDM {
    address public owner;
    address public sequencer;

    struct TxWithGas {
        uint32 txIndex;
        bytes32 txHash;
        uint64 opgas;
    }

    event OwnerTransferred(address indexed previousOwner, address indexed newOwner);
    event SequencerUpdated(address indexed previousSequencer, address indexed newSequencer);

    event BlockData(uint256 indexed blockId, bytes32 payloadHash, bytes payload);

    modifier onlyOwner() {
        require(msg.sender == owner, "SDM: not owner");
        _;
    }

    modifier onlySequencer() {
        require(msg.sender == sequencer, "SDM: not sequencer");
        _;
    }

    constructor(address initialSequencer) {
        require(initialSequencer != address(0), "SDM: zero sequencer");
        owner = msg.sender;
        sequencer = initialSequencer;

        emit OwnerTransferred(address(0), msg.sender);
        emit SequencerUpdated(address(0), initialSequencer);
    }

    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "SDM: zero owner");
        emit OwnerTransferred(owner, newOwner);
        owner = newOwner;
    }

    function setSequencer(address newSequencer) external onlyOwner {
        require(newSequencer != address(0), "SDM: zero sequencer");
        emit SequencerUpdated(sequencer, newSequencer);
        sequencer = newSequencer;
    }

    /**
     * Sequencer-only publish.
     * Emits the payload so nodes/indexers can read it from logs (optional).
     */
    function publishBlockData(uint256 blockId, bytes calldata payload) external onlySequencer {
        emit BlockData(blockId, keccak256(payload), payload);
    }

    /**
     * Decode helper returning parallel arrays.
     */
    function decodePayload(bytes calldata payload)
        external
        pure
        returns (uint32[] memory txIndices, bytes32[] memory txHashes, uint64[] memory opgases)
    {
        TxWithGas[] memory entries = abi.decode(payload, (TxWithGas[]));
        uint256 n = entries.length;

        txIndices = new uint32[](n);
        txHashes = new bytes32[](n);
        opgases = new uint64[](n);

        for (uint256 i = 0; i < n; i++) {
            txIndices[i] = entries[i].txIndex;
            txHashes[i] = entries[i].txHash;
            opgases[i] = entries[i].opgas;
        }
    }

    /**
     * Decode helper returning the struct array directly.
     * Off-chain ABI type: tuple(uint32 txIndex,bytes32 txHash,uint64 opgas)[]
     */
    function decodePayloadAsStruct(bytes calldata payload) external pure returns (TxWithGas[] memory entries) {
        entries = abi.decode(payload, (TxWithGas[]));
    }

    function payloadHash(bytes calldata payload) external pure returns (bytes32) {
        return keccak256(payload);
    }
}
