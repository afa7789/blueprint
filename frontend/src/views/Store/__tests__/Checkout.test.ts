import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

// localStorage stub (cart store + auth store touch it on init)
const ls: Record<string, string> = {}
vi.stubGlobal('localStorage', {
  getItem: (k: string) => ls[k] ?? null,
  setItem: (k: string, v: string) => { ls[k] = v },
  removeItem: (k: string) => { delete ls[k] },
  clear: () => { Object.keys(ls).forEach(k => delete ls[k]) },
})

const mocks = vi.hoisted(() => ({
  apiPost: vi.fn(),
  apiGet: vi.fn(),
  stripeConfirmPayment: vi.fn(),
  stripeElementMount: vi.fn(),
  stripeElementsCreate: vi.fn(),
  stripeElements: vi.fn(),
  loadStripeMock: vi.fn(),
  routerPush: vi.fn(),
}))
mocks.stripeElementsCreate.mockImplementation(() => ({ mount: mocks.stripeElementMount }))
mocks.stripeElements.mockImplementation(() => ({ create: mocks.stripeElementsCreate }))

vi.mock('../../../services/api', () => ({
  api: { get: mocks.apiGet, post: mocks.apiPost, delete: vi.fn(), put: vi.fn() },
  ApiError: class ApiError extends Error { status = 0 },
}))

vi.mock('../../../services/featureFlags', () => ({
  fetchFeatureFlags: vi.fn().mockResolvedValue(undefined),
  isFeatureEnabled: (key: string) =>
    key === 'payments_stripe' || key === 'payments_pix' || key === 'pix_manual_enabled',
}))

vi.mock('@stripe/stripe-js', () => ({
  loadStripe: (...args: unknown[]) => mocks.loadStripeMock(...args),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mocks.routerPush }),
  useRoute: () => ({ query: {} }),
}))

const { apiPost, apiGet, stripeConfirmPayment, stripeElementMount, stripeElementsCreate, stripeElements, loadStripeMock, routerPush } = mocks

import Checkout from '../Checkout.vue'
import { useCartStore } from '../../../stores/cart'
import { useAuthStore } from '../../../stores/auth'

function mountCheckout() {
  return mount(Checkout, {
    global: {
      stubs: { 'router-link': { template: '<a><slot/></a>' } },
    },
  })
}

async function seedCart() {
  const cart = useCartStore()
  cart.addItem({ id: 'p1', name: 'Item', price: 50, imageUrl: '' })
}

describe('Checkout — Stripe inline branch', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiPost.mockReset()
    apiGet.mockReset()
    loadStripeMock.mockReset()
    stripeConfirmPayment.mockReset()
    stripeElementMount.mockReset()
    routerPush.mockReset()
    Object.keys(ls).forEach(k => delete ls[k])
  })

  it('after placing order with stripe method, calls /payments/stripe and mounts PaymentElement', async () => {
    seedCart()
    const auth = useAuthStore()
    auth.user = { id: 'u1', email: 'u@example.com', role: 'user' } as never

    apiPost.mockResolvedValueOnce({ id: 'order-1' }) // POST /orders
    apiPost.mockResolvedValueOnce({
      client_secret: 'cs_123',
      publishable_key: 'pk_test',
      payment_intent_id: 'pi_1',
      status: 'requires_payment_method',
    })
    apiGet.mockResolvedValueOnce([]) // saved-cards

    loadStripeMock.mockResolvedValueOnce({
      elements: stripeElements,
      confirmPayment: stripeConfirmPayment,
    })

    const wrapper = mountCheckout()
    await flushPromises()

    // Fill shipping
    const inputs = wrapper.findAll('input.input')
    expect(inputs.length).toBeGreaterThan(0)
    for (const i of inputs) {
      const el = i.element as HTMLInputElement
      if (el.type === 'text') await i.setValue('x')
    }

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(apiPost).toHaveBeenNthCalledWith(1, '/api/v1/orders', expect.any(Object))
    expect(apiPost).toHaveBeenNthCalledWith(2, '/api/v1/payments/stripe', { order_id: 'order-1' })
    expect(loadStripeMock).toHaveBeenCalledWith('pk_test')
    expect(stripeElementsCreate).toHaveBeenCalledWith('payment', expect.objectContaining({ layout: 'tabs' }))
    expect(stripeElementMount).toHaveBeenCalled()
    expect(wrapper.find('[data-testid="confirm-stripe-btn"]').exists()).toBe(true)
  })

  it('shows error when Stripe.js fails to load', async () => {
    seedCart()
    apiPost.mockResolvedValueOnce({ id: 'order-2' })
    apiPost.mockResolvedValueOnce({
      client_secret: 'cs', publishable_key: 'pk', payment_intent_id: 'pi', status: 'x',
    })
    apiGet.mockResolvedValueOnce([])
    loadStripeMock.mockResolvedValueOnce(null)

    const wrapper = mountCheckout()
    await flushPromises()
    const inputs = wrapper.findAll('input.input')
    for (const i of inputs) {
      const el = i.element as HTMLInputElement
      if (el.type === 'text') await i.setValue('x')
    }

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Stripe.js failed to load')
  })

  it('lists saved cards for an authenticated user', async () => {
    seedCart()
    const auth = useAuthStore()
    auth.user = { id: 'u1', email: 'u@example.com', role: 'user' } as never

    apiPost.mockResolvedValueOnce({ id: 'order-3' })
    apiPost.mockResolvedValueOnce({
      client_secret: 'cs', publishable_key: 'pk', payment_intent_id: 'pi', status: 'x',
    })
    apiGet.mockResolvedValueOnce([
      { id: 'pm_1', brand: 'visa', last4: '4242', exp_month: 12, exp_year: 2030 },
    ])
    loadStripeMock.mockResolvedValueOnce({
      elements: stripeElements,
      confirmPayment: stripeConfirmPayment,
    })

    const wrapper = mountCheckout()
    await flushPromises()
    const inputs = wrapper.findAll('input.input')
    for (const i of inputs) {
      const el = i.element as HTMLInputElement
      if (el.type === 'text') await i.setValue('x')
    }
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(apiGet).toHaveBeenCalledWith('/api/v1/user/saved-cards')
    expect(wrapper.text()).toContain('4242')
  })

  it('selecting a saved card and confirming sends payment_method_id and redirects', async () => {
    seedCart()
    const auth = useAuthStore()
    auth.user = { id: 'u1', email: 'u@example.com', role: 'user' } as never

    apiPost.mockResolvedValueOnce({ id: 'order-4' })
    apiPost.mockResolvedValueOnce({
      client_secret: 'cs', publishable_key: 'pk', payment_intent_id: 'pi', status: 'x',
    })
    apiGet.mockResolvedValueOnce([
      { id: 'pm_saved', brand: 'visa', last4: '0001', exp_month: 1, exp_year: 2099 },
    ])
    loadStripeMock.mockResolvedValueOnce({
      elements: stripeElements,
      confirmPayment: stripeConfirmPayment,
    })
    apiPost.mockResolvedValueOnce({ client_secret: 'cs_saved', status: 'succeeded' }) // third post: confirm with payment_method_id

    const wrapper = mountCheckout()
    await flushPromises()
    const inputs = wrapper.findAll('input.input')
    for (const i of inputs) {
      const el = i.element as HTMLInputElement
      if (el.type === 'text') await i.setValue('x')
    }
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    // Pick the saved card
    const savedRadios = wrapper.findAll('input[name="saved-card"]')
    expect(savedRadios.length).toBeGreaterThanOrEqual(2)
    await savedRadios[0].setValue('pm_saved')

    await wrapper.find('[data-testid="confirm-stripe-btn"]').trigger('click')
    await flushPromises()

    expect(apiPost).toHaveBeenLastCalledWith('/api/v1/payments/stripe', {
      order_id: 'order-4',
      payment_method_id: 'pm_saved',
    })
    expect(routerPush).toHaveBeenCalledWith(expect.stringContaining('/store/orders?payment=success&order=order-4'))
  })
})
