import type { Page } from '@playwright/test'

export type SyncloudCreds = {
  user: string
  password: string
}

export function credsFromEnv(): SyncloudCreds {
  const user = process.env.DEVICE_USER
  const password = process.env.DEVICE_PASSWORD
  if (!user || !password) {
    throw new Error('DEVICE_USER and DEVICE_PASSWORD must be set')
  }
  return { user, password }
}

export async function openApp(page: Page, path = ''): Promise<void> {
  await page.goto(path, { waitUntil: 'domcontentloaded' })
}

export async function loginOidc(page: Page, creds: SyncloudCreds): Promise<void> {
  await page.locator('#username-textfield').fill(creds.user)
  await page.locator('#password-textfield').fill(creds.password)
  await page.locator('#sign-in-button').click()
}
