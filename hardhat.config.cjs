// Minimal Hardhat config used ONLY by the chain-sim container.
// The real contracts/hardhat.config.ts lands in WP-03.U4 with the
// ZTCVReceiptAnchor.sol toolchain (Solidity 0.8.27 + tests + deploy
// scripts). This file exists solely so `npx hardhat node` boots inside
// the chain-sim container during local dev.
module.exports = {
  solidity: "0.8.27",
  networks: {
    hardhat: {
      // Hardhat node defaults: chainId 31337, 20 funded accounts.
    },
  },
};
