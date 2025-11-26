import request from '../utils/request'

export const getDeliveryFeeSettings = () => {
  return request({
    url: '/admin/delivery-fee/settings',
    method: 'get'
  })
}

export const updateDeliveryFeeSettings = (data) => {
  return request({
    url: '/admin/delivery-fee/settings',
    method: 'put',
    data
  })
}

export const getDeliveryFeeExclusions = () => {
  return request({
    url: '/admin/delivery-fee/exclusions',
    method: 'get'
  })
}

export const createDeliveryFeeExclusion = (data) => {
  return request({
    url: '/admin/delivery-fee/exclusions',
    method: 'post',
    data
  })
}

export const updateDeliveryFeeExclusion = (id, data) => {
  return request({
    url: `/admin/delivery-fee/exclusions/${id}`,
    method: 'put',
    data
  })
}

export const deleteDeliveryFeeExclusion = (id) => {
  return request({
    url: `/admin/delivery-fee/exclusions/${id}`,
    method: 'delete'
  })
}

