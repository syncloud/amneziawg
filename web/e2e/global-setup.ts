import { chromium, type FullConfig } from '@playwright/test'
import { mkdir } from 'node:fs/promises'
import { dirname } from 'node:path'
import { credsFromEnv, loginOidc, openApp } from './helpers/syncloud'

const storageStatePath = 'e2e/.auth/user.json'

export default async function globalSetup(config: FullConfig): Promise<void> {
  const baseURL = config.projects[0].use.baseURL
  if (!baseURL) throw new Error('global-setup: baseURL not configured')

  await mkdir(dirname(storageStatePath), { recursive: true })

  const browser = await chromium.launch()
  const context = await browser.newContext({ ignoreHTTPSErrors: true, baseURL })
  const page = await context.newPage()

  await openApp(page)
  await loginOidc(page, credsFromEnv())
  await page.waitForURL((url) => url.origin === new URL(baseURL).origin, { timeout: 30_000 })

  await context.storageState({ path: storageStatePath })
  await browser.close()
}
