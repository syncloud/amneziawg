import { test, expect } from '@playwright/test'

test('dashboard shows router-forward reminder', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText(/Router port forwarding required/i)).toBeVisible()
  await expect(page.getByText(/Forward UDP port/i)).toBeVisible()
})

test('peers page opens add-peer dialog', async ({ page }) => {
  await page.goto('/peers')
  await page.getByRole('button', { name: 'Add peer' }).click()
  await expect(page.getByPlaceholder(/laptop, phone/i)).toBeVisible()
})
