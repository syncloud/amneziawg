import { test, expect } from '@playwright/test'

test('open app, see dashboard, navigate to peers, open add-peer dialog', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByTestId('port-forward-reminder')).toBeVisible({ timeout: 15_000 })
  await expect(page.getByTestId('listen-port')).toBeVisible()

  await page.getByTestId('nav-peers').click()
  await page.getByTestId('add-peer-button').click()
  await expect(page.getByTestId('peer-name-input')).toBeVisible()
})
