import Dexie from 'dexie'
import type { Table } from 'dexie'

interface CachedProduct {
  id: string
  name: string
  price: number
  imageUrl: string
  updatedAt: number
}

interface PendingOrder {
  id: string
  items: Array<{ productId: string; quantity: number }>
  total: number
  createdAt: number
}

class BlueprintDB extends Dexie {
  products!: Table<CachedProduct>
  pendingOrders!: Table<PendingOrder>

  constructor() {
    super('blueprint')
    this.version(1).stores({
      products: 'id, updatedAt',
      pendingOrders: 'id, createdAt'
    })
  }
}

export const db = new BlueprintDB()
