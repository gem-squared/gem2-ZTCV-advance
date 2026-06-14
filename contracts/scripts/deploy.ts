import hre from 'hardhat'
import fs from 'fs'
import path from 'path'

async function main() {
  const network = hre.network.name
  console.log(`Deploying ZTCVReceiptAnchor to: ${network}`)

  const F = await hre.ethers.getContractFactory('ZTCVReceiptAnchor')
  const c = await F.deploy()
  await c.waitForDeployment()
  const address = await c.getAddress()
  const deployTx = c.deploymentTransaction()
  const txHash = deployTx?.hash || ''

  const chainId = Number((await hre.ethers.provider.getNetwork()).chainId)

  console.log(`✓ Address  : ${address}`)
  console.log(`✓ Tx Hash  : ${txHash}`)
  console.log(`✓ Chain ID : ${chainId}`)

  const out = {
    network,
    address,
    deployTxHash: txHash,
    chainId,
    deployedAt: new Date().toISOString(),
    contractName: 'ZTCVReceiptAnchor'
  }

  const outDir = path.join(__dirname, '..', 'deployments')
  fs.mkdirSync(outDir, { recursive: true })
  fs.writeFileSync(
    path.join(outDir, `${network}.json`),
    JSON.stringify(out, null, 2) + '\n'
  )
  console.log(`✓ Deployment info → contracts/deployments/${network}.json`)

  if (network === 'sepolia') {
    console.log(`✓ Etherscan      : https://sepolia.etherscan.io/address/${address}`)
    console.log(`✓ Deploy tx scan : https://sepolia.etherscan.io/tx/${txHash}`)
  }
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
