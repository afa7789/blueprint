import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface CartItem {
  productId: string
  name: string
  price: number
  quantity: number
  imageUrl: string
}

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])
  const discount = ref(0)
  const couponCode = ref('')

  const itemCount = computed(() => items.value.reduce((sum, i) => sum + i.quantity, 0))
  const subtotal = computed(() => items.value.reduce((sum, i) => sum + i.price * i.quantity, 0))
  const total = computed(() => Math.max(0, subtotal.value - discount.value))

  function addItem(product: { id: string; name: string; price: number; imageUrl: string }) {
    const existing = items.value.find(i => i.productId === product.id)
    if (existing) {
      existing.quantity++
    } else {
      items.value.push({ productId: product.id, name: product.name, price: product.price, quantity: 1, imageUrl: product.imageUrl })
    }
    save()
  }

  function removeItem(productId: string) {
    items.value = items.value.filter(i => i.productId !== productId)
    save()
  }

  function updateQuantity(productId: string, qty: number) {
    const item = items.value.find(i => i.productId === productId)
    if (item) {
      item.quantity = Math.max(1, qty)
      save()
    }
  }

  function clear() {
    items.value = []
    discount.value = 0
    couponCode.value = ''
    save()
  }

  function save() {
    localStorage.setItem('cart', JSON.stringify(items.value))
  }

  function load() {
    const saved = localStorage.getItem('cart')
    if (saved) items.value = JSON.parse(saved)
  }

  load()

  return { items, itemCount, subtotal, discount, couponCode, total, addItem, removeItem, updateQuantity, clear }
})
