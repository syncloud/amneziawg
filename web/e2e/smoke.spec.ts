import { test, expect } from '@playwright/test'

test('open app, see peers page, open add-peer dialog', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByTestId('endpoint')).toBeVisible({ timeout: 15_000 })
  await expect(page.getByTestId('dashboard-error')).toHaveCount(0)
  await expect(page.getByTestId('listen-port')).toHaveText(/^\d+$/)

  await page.getByTestId('add-peer-button').click()
  await expect(page.getByTestId('peer-name-input')).toBeVisible()
})
