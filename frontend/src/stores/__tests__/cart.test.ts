import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Stub localStorage before importing the store (store calls localStorage.setItem on init)
const localStorageData: Record<string, string> = {}
vi.stubGlobal('localStorage', {
  getItem: (key: string) => localStorageData[key] ?? null,
  setItem: (key: string, value: string) => { localStorageData[key] = value },
  removeItem: (key: string) => { delete localStorageData[key] },
  clear: () => { Object.keys(localStorageData).forEach(k => delete localStorageData[k]) },
})

import { useCartStore } from '../cart'

describe('cart store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    Object.keys(localStorageData).forEach(k => delete localStorageData[k])
  })

  it('starts empty', () => {
    const cart = useCartStore()
    expect(cart.items).toEqual([])
    expect(cart.itemCount).toBe(0)
    expect(cart.subtotal).toBe(0)
  })

  it('adds item', () => {
    const cart = useCartStore()
    cart.addItem({ id: '1', name: 'Test', price: 10, imageUrl: '' })
    expect(cart.items).toHaveLength(1)
    expect(cart.itemCount).toBe(1)
    expect(cart.subtotal).toBe(10)
  })

  it('increments quantity for same product', () => {
    const cart = useCartStore()
    cart.addItem({ id: '1', name: 'Test', price: 10, imageUrl: '' })
    cart.addItem({ id: '1', name: 'Test', price: 10, imageUrl: '' })
    expect(cart.items).toHaveLength(1)
    expect(cart.items[0].quantity).toBe(2)
    expect(cart.subtotal).toBe(20)
  })

  it('removes item', () => {
    const cart = useCartStore()
    cart.addItem({ id: '1', name: 'Test', price: 10, imageUrl: '' })
    cart.removeItem('1')
    expect(cart.items).toHaveLength(0)
  })

  it('updates quantity', () => {
    const cart = useCartStore()
    cart.addItem({ id: '1', name: 'Test', price: 10, imageUrl: '' })
    cart.updateQuantity('1', 5)
    expect(cart.items[0].quantity).toBe(5)
    expect(cart.subtotal).toBe(50)
  })

  it('minimum quantity is 1', () => {
    const cart = useCartStore()
    cart.addItem({ id: '1', name: 'Test', price: 10, imageUrl: '' })
    cart.updateQuantity('1', 0)
    expect(cart.items[0].quantity).toBe(1)
  })

  it('clears cart', () => {
    const cart = useCartStore()
    cart.addItem({ id: '1', name: 'A', price: 10, imageUrl: '' })
    cart.addItem({ id: '2', name: 'B', price: 20, imageUrl: '' })
    cart.clear()
    expect(cart.items).toHaveLength(0)
    expect(cart.subtotal).toBe(0)
  })

  it('calculates total with discount', () => {
    const cart = useCartStore()
    cart.addItem({ id: '1', name: 'Test', price: 100, imageUrl: '' })
    cart.discount = 20
    expect(cart.total).toBe(80)
  })

  it('total never goes below 0', () => {
    const cart = useCartStore()
    cart.addItem({ id: '1', name: 'Test', price: 10, imageUrl: '' })
    cart.discount = 999
    expect(cart.total).toBe(0)
  })
})
