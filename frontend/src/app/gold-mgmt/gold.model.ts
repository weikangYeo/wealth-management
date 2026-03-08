export interface GoldTxn {
  id: number
  bank: string
  txnDate: string
  gram: number
  unitPrice: number
  totalPrice: number
  txnType: 'BUY' | 'SELL'
}
