// request.js - 小程序请求工具类

// 基础API地址（生产环境）
// export const BASE_URL = 'https://mall.sscchh.com/api/mini';
export const BASE_URL = 'http://192.168.1.123:8082/api/mini';


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
            // 检查业务错误码，如果需要清空登录信息
            if (res.data && shouldClearAuthInfo(res.statusCode, res.data.code)) {
              clearLocalAuthInfo();
            }
            // 错误提示
            uni.showToast({
              title: res.data.message || res.data.msg || '请求失败',
              icon: 'none'
            });
            reject(res.data);
          }
        } else {
          // HTTP状态码错误（401, 403, 404等）
          if (shouldClearAuthInfo(res.statusCode, null)) {
            clearLocalAuthInfo();
          }
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

// 清空本地认证信息
function clearLocalAuthInfo() {
  try {
    // 清空token和用户信息
    uni.removeStorageSync('miniUserToken');
    uni.removeStorageSync('miniUserInfo');
    uni.removeStorageSync('miniUserUniqueId');
    
    // 延迟执行，避免在请求回调中直接跳转
    setTimeout(() => {
      // 跳转到首页（如果不在首页）
      const pages = getCurrentPages();
      if (pages.length > 0) {
        const currentPage = pages[pages.length - 1];
        const route = currentPage.route;
        // 如果不在首页，跳转到首页
        if (route !== 'pages/index/index') {
          uni.reLaunch({
            url: '/pages/index/index'
          });
        } else {
          // 如果在首页，刷新页面数据
          if (currentPage.$vm && typeof currentPage.$vm.updateUserInfo === 'function') {
            currentPage.$vm.updateUserInfo();
          }
          // 触发页面刷新
          if (currentPage.$vm && typeof currentPage.$vm.onShow === 'function') {
            currentPage.$vm.onShow();
          }
        }
      }
    }, 100);
  } catch (error) {
    console.error('清空本地信息失败:', error);
  }
}

// 检查状态码是否需要清空登录信息
function shouldClearAuthInfo(statusCode, businessCode) {
  // HTTP状态码：401未授权、403禁止、404未找到
  if (statusCode === 401 || statusCode === 403 || statusCode === 404) {
    return true;
  }
  // 业务错误码：401未授权、404未找到
  if (businessCode === 401 || businessCode === 404) {
    return true;
  }
  return false;
}

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