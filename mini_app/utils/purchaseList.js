import {
  getPurchaseListSummary,
  addPurchaseListItem,
  updatePurchaseListItem,
  deletePurchaseListItem,
  clearPurchaseList
} from '../api/index'

/** tabBar 中「采购单」项的索引（从 0 开始） */
export const TAB_BAR_CART_INDEX = 2

/**
 * 设置采购单 tabBar 角标（商品种类数）
 * @param {number} count - 采购单商品种类数，0 时移除角标
 */
export const setPurchaseListTabBarBadge = (count) => {
  if (count == null || count <= 0) {
    try {
      uni.removeTabBarBadge({ index: TAB_BAR_CART_INDEX })
    } catch (e) {
      // 忽略未配置 tabBar 等环境
    }
    return
  }
  try {
    const text = count > 99 ? '99+' : String(count)
    uni.setTabBarBadge({ index: TAB_BAR_CART_INDEX, text })
  } catch (e) {
    // 忽略未配置 tabBar 等环境
  }
}

/**
 * 根据当前登录态拉取采购单并更新 tabBar 角标（种类数）
 */
export const updatePurchaseListTabBarBadge = async () => {
  const token = uni.getStorageSync('miniUserToken') || ''
  if (!token) {
    setPurchaseListTabBarBadge(0)
    return
  }
  try {
    const { items } = await fetchPurchaseList(token)
    setPurchaseListTabBarBadge(Array.isArray(items) ? items.length : 0)
  } catch (e) {
    setPurchaseListTabBarBadge(0)
  }
}

export const fetchPurchaseList = async (token, itemIds = null) => {
  if (!token) return { items: [], summary: null, availableCoupons: [], bestCombination: null }
  const params = {}
  if (itemIds && Array.isArray(itemIds) && itemIds.length > 0) {
    params.item_ids = itemIds.join(',')
  }
  const res = await getPurchaseListSummary(token, params)
  if (res && res.code === 200 && res.data) {
    return {
      items: Array.isArray(res.data.items) ? res.data.items : [],
      summary: res.data.summary || null,
      availableCoupons: Array.isArray(res.data.available_coupons) ? res.data.available_coupons : [],
      bestCombination: res.data.best_combination || null
    }
  }
  return { items: [], summary: null, availableCoupons: [], bestCombination: null }
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

