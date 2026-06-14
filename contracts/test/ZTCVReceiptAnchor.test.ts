import { expect } from 'chai'
import { ethers } from 'hardhat'

describe('ZTCVReceiptAnchor', () => {
  const sessionHash   = '0x' + 'aa'.repeat(32)
  const receiptHash   = '0x' + 'bb'.repeat(32)
  const policyVersion = '0x' + 'cc'.repeat(32)

  async function deploy() {
    const F = await ethers.getContractFactory('ZTCVReceiptAnchor')
    const c = await F.deploy()
    await c.waitForDeployment()
    return c
  }

  it('emits ReceiptAnchored on recordVerification', async () => {
    const c = await deploy()
    await expect(c.recordVerification(sessionHash, receiptHash, true, policyVersion))
      .to.emit(c, 'ReceiptAnchored')
  })

  it('stores receipt correctly', async () => {
    const c = await deploy()
    await c.recordVerification(sessionHash, receiptHash, true, policyVersion)
    const r = await c.getReceipt(sessionHash)
    expect(r.sessionHash).to.equal(sessionHash)
    expect(r.receiptHash).to.equal(receiptHash)
    expect(r.isSafe).to.equal(true)
    expect(r.policyVersion).to.equal(policyVersion)
    expect(r.timestamp).to.be.gt(0n)
  })

  it('reverts on duplicate sessionHash', async () => {
    const c = await deploy()
    await c.recordVerification(sessionHash, receiptHash, true, policyVersion)
    await expect(
      c.recordVerification(sessionHash, receiptHash, true, policyVersion)
    )
      .to.be.revertedWithCustomError(c, 'ReceiptAlreadyAnchored')
      .withArgs(sessionHash)
  })

  it('getReceipt returns stored data for BLOCK case (isSafe=false)', async () => {
    const c = await deploy()
    await c.recordVerification(sessionHash, receiptHash, false, policyVersion)
    const r = await c.getReceipt(sessionHash)
    expect(r.isSafe).to.equal(false)
  })

  it('reverts on getReceipt for missing sessionHash', async () => {
    const c = await deploy()
    await expect(c.getReceipt(sessionHash))
      .to.be.revertedWithCustomError(c, 'ReceiptNotFound')
      .withArgs(sessionHash)
  })
})
