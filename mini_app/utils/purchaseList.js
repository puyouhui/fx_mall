import {
  getPurchaseList,
  addPurchaseListItem,
  updatePurchaseListItem,
  deletePurchaseListItem,
  clearPurchaseList
} from '../api/index'

export const fetchPurchaseList = async (token) => {
  if (!token) return []
  const res = await getPurchaseList(token)
  if (res && res.code === 200) {
    return Array.isArray(res.data) ? res.data : []
  }
  return []
}

export const addItemToPurchaseList = async ({ token, productId, specName, quantity = 1 }) => {
  if (!token) {
    throw new Error('缺少身份凭证')
  }
  const res = await addPurchaseListItem({
    product_id: productId,
    spec_name: specName,
    quantity
  }, token)
  return res
}

export const updatePurchaseListQuantity = async ({ token, itemId, quantity }) => {
  if (!token) {
    throw new Error('缺少身份凭证')
  }
  const res = await updatePurchaseListItem(itemId, { quantity }, token)
  return res
}

export const deletePurchaseListItemById = async ({ token, itemId }) => {
  if (!token) {
    throw new Error('缺少身份凭证')
  }
  return deletePurchaseListItem(itemId, token)
}

export const clearPurchaseListByToken = async (token) => {
  if (!token) {
    throw new Error('缺少身份凭证')
  }
  return clearPurchaseList(token)
}

