import { chromium, type FullConfig } from '@playwright/test'
import { mkdir, writeFile } from 'node:fs/promises'
import { dirname } from 'node:path'
import { acceptConsent, credsFromEnv, loginOidc, openApp } from './helpers/syncloud'

const storageStatePath = 'e2e/.auth/user.json'

export default async function globalSetup(config: FullConfig): Promise<void> {
  const baseURL = config.projects[0].use.baseURL
  if (!baseURL) throw new Error('global-setup: baseURL not configured')

  await mkdir(dirname(storageStatePath), { recursive: true })

  const browser = await chromium.launch()
  const context = await browser.newContext({ ignoreHTTPSErrors: true, baseURL })
  const page = await context.newPage()

  try {
    const appOrigin = new URL(baseURL).origin
    const backAtApp = (url: URL) =>
      url.origin === appOrigin && !url.pathname.startsWith('/auth/')
    await openApp(page)
    await page.waitForURL(/^https:\/\/auth\./, { timeout: 30_000 })
    console.log('global-setup: at auth page, url=', page.url())
    await loginOidc(page, credsFromEnv())
    await page.waitForURL(
      (url) => url.pathname.includes('/consent/') || backAtApp(url),
      { timeout: 30_000 }
    )
    if (page.url().includes('/consent/')) {
      console.log('global-setup: at consent page, url=', page.url())
      await acceptConsent(page)
      await page.waitForURL(backAtApp, { timeout: 30_000 })
    }
    console.log('global-setup: back on app, url=', page.url())
    await context.storageState({ path: storageStatePath })
  } catch (err) {
    console.log('global-setup: failed, url=', page.url())
    await mkdir('test-results', { recursive: true })
    await page.screenshot({ path: 'test-results/global-setup-fail.png', fullPage: true })
    await writeFile('test-results/global-setup-fail.html', await page.content())
    throw err
  } finally {
    await browser.close()
  }
}
