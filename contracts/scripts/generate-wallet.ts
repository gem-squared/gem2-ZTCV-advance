import { ethers } from 'ethers'
import fs from 'fs'
import path from 'path'

function main() {
  console.log('Generating fresh Demo wallet for ZTCV contract deployment')
  console.log('  (per W1 invariant: NEVER use a private wallet — this is a Demo wallet only)')
  console.log('')

  const wallet = ethers.Wallet.createRandom()

  console.log(`Address      : ${wallet.address}`)
  console.log(`Private Key  : ${wallet.privateKey}`)
  console.log(`Mnemonic     : ${wallet.mnemonic?.phrase}`)
  console.log('')
  console.log('NEXT STEPS:')
  console.log('  1. Copy private key into contracts/.env:')
  console.log(`       DEPLOYER_PRIVATE_KEY=${wallet.privateKey}`)
  console.log('     (the .env file is gitignored — never reaches remote)')
  console.log('')
  console.log(`  2. Fund the address ${wallet.address} via a Sepolia faucet:`)
  console.log('       https://sepolia-faucet.pk910.de/   (PoW — no mainnet ETH required)')
  console.log('       https://www.alchemy.com/faucets/ethereum-sepolia')
  console.log('       https://cloud.google.com/application/web3/faucet/ethereum/sepolia')
  console.log('       https://faucets.chain.link/sepolia')
  console.log('')
  console.log(`  3. Verify funding on Etherscan:`)
  console.log(`       https://sepolia.etherscan.io/address/${wallet.address}`)
  console.log('')
  console.log('  4. Once funded with ≥ 0.01 SepoliaETH, run:')
  console.log('       npm run compile')
  console.log('       npm run test')
  console.log('       npm run deploy:sepolia')
  console.log('       npm run extract-abi')
  console.log('')
  console.log('SECURITY')
  console.log('  - This is a Demo wallet — testnet ETH only, never mainnet')
  console.log('  - If compromised, generate a new wallet (cost = nothing)')
  console.log('  - Private key NEVER committed to git (.env is gitignored)')

  const outDir = path.join(__dirname, '..', 'deployments')
  fs.mkdirSync(outDir, { recursive: true })
  fs.writeFileSync(
    path.join(outDir, 'demo-wallet-address.txt'),
    `${wallet.address}\n`
  )
  console.log('')
  console.log(`✓ Address marker → contracts/deployments/demo-wallet-address.txt`)
}

main()
