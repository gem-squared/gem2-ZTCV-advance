// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

/**
 * @title  ZTCVReceiptAnchor
 * @notice Anchors per-call verification receipts on-chain. On-chain payload
 *         is hashed — NO PII ever leaves the off-chain layer. Every
 *         CallPassport (safe or blocked) produces one receipt; blocked
 *         calls are anchored with isSafe=false ("차단 영수증 기록").
 * @dev    Gas-conscious: custom errors instead of revert strings, single
 *         SSTORE per anchor. Target: EVM testnets (Sepolia) and OmniOne
 *         Chain via developer-portal upload.
 */
contract ZTCVReceiptAnchor {
    struct CallReceipt {
        bytes32 sessionHash;
        bytes32 receiptHash;
        uint256 timestamp;
        bool    isSafe;
        bytes32 policyVersion;
    }

    /// @notice sessionHash → CallReceipt
    mapping(bytes32 => CallReceipt) public registry;

    event ReceiptAnchored(
        bytes32 indexed sessionHash,
        bytes32         receiptHash,
        bool            isSafe,
        bytes32         policyVersion,
        uint256         timestamp
    );

    error ReceiptAlreadyAnchored(bytes32 sessionHash);
    error ReceiptNotFound(bytes32 sessionHash);

    /// @notice Anchor a verification receipt. Reverts on duplicate sessionHash.
    function recordVerification(
        bytes32 _sessionHash,
        bytes32 _receiptHash,
        bool    _isSafe,
        bytes32 _policyVersion
    ) external {
        if (registry[_sessionHash].timestamp != 0) {
            revert ReceiptAlreadyAnchored(_sessionHash);
        }
        registry[_sessionHash] = CallReceipt({
            sessionHash:   _sessionHash,
            receiptHash:   _receiptHash,
            timestamp:     block.timestamp,
            isSafe:        _isSafe,
            policyVersion: _policyVersion
        });
        emit ReceiptAnchored(
            _sessionHash,
            _receiptHash,
            _isSafe,
            _policyVersion,
            block.timestamp
        );
    }

    /// @notice Fetch a stored receipt. Reverts if not found.
    function getReceipt(bytes32 _sessionHash) external view returns (CallReceipt memory) {
        CallReceipt memory r = registry[_sessionHash];
        if (r.timestamp == 0) {
            revert ReceiptNotFound(_sessionHash);
        }
        return r;
    }
}
