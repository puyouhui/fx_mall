// request.js - 小程序请求工具类

// 基础API地址
export const BASE_URL = 'http://localhost:8082/api/mini';

// 封装请求方法
export const request = (options = {}) => {
  // 设置默认请求参数
  const defaultOptions = {
    url: '',
    method: 'GET',
    data: {},
    header: {
      'content-type': 'application/json'
    },
    success: () => {},
    fail: () => {},
    complete: () => {}
  };

  // 合并用户参数和默认参数
  const finalOptions = { ...defaultOptions, ...options };
  finalOptions.url = BASE_URL + finalOptions.url;

  // 返回Promise对象
  return new Promise((resolve, reject) => {
    // 调用uni.request发起请求
    uni.request({
      ...finalOptions,
      success: (res) => {
        // 处理成功响应
        if (res.statusCode === 200) {
          if (res.data && res.data.code === 200) {
            resolve(res.data);
          } else {
            // 错误提示
            uni.showToast({
              title: res.data.msg || '请求失败',
              icon: 'none'
            });
            reject(res.data);
          }
        } else {
          // HTTP错误
          uni.showToast({
            title: `HTTP错误: ${res.statusCode}`,
            icon: 'none'
          });
          reject(res);
        }
      },
      fail: (err) => {
        // 网络错误
        uni.showToast({
          title: '网络连接失败',
          icon: 'none'
        });
        reject(err);
      }
    });
  });
};

// 封装GET请求
export const get = (url, data = {}, options = {}) => {
  console.log('发送GET请求，URL:', url, '参数:', data);
  return request({
    url,
    method: 'GET',
    data,
    ...options
  });
};

// 封装POST请求
export const post = (url, data = {}, options = {}) => {
  return request({
    url,
    method: 'POST',
    data,
    ...options
  });
};

// 封装PUT请求
export const put = (url, data = {}, options = {}) => {
  return request({
    url,
    method: 'PUT',
    data,
    ...options
  });
};

// 封装DELETE请求
export const del = (url, data = {}, options = {}) => {
  return request({
    url,
    method: 'DELETE',
    data,
    ...options
  });
};