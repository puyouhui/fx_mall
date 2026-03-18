// request.js - 小程序请求工具类

// 基础API地址（生产环境）
export const BASE_URL = 'https://api.sscchh.com/api/mini';
// export const BASE_URL = 'http://192.168.1.3:8082/api/mini';


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
    complete: () => {},
    ignoreAuthClear: false
  };

  // 合并用户参数和默认参数
  const finalOptions = { ...defaultOptions, ...options };
  const silent = finalOptions.silent === true;
  const ignoreAuthClear = finalOptions.ignoreAuthClear === true;
  delete finalOptions.silent;
  delete finalOptions.ignoreAuthClear;
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
            // 检查业务错误码，如果需要清空登录信息（如 401 token 过期）
            const needClearAuth = res.data && shouldClearAuthInfo(res.statusCode, res.data.code, silent, ignoreAuthClear);
            if (needClearAuth) {
              clearLocalAuthInfo();
            }
            // 401/403 时静默处理，不显示错误 toast，由自动登录接管
            if (!silent && !needClearAuth) {
              uni.showToast({
                title: res.data.message || res.data.msg || '请求失败',
                icon: 'none'
              });
            }
            reject(res.data);
          }
        } else {
          // HTTP状态码错误（401, 403, 404等）
          const needClearAuth = shouldClearAuthInfo(res.statusCode, null, silent, ignoreAuthClear);
          if (needClearAuth) {
            clearLocalAuthInfo();
          }
          // 401/403 时静默处理，不显示错误 toast，由自动登录接管
          if (!silent && !needClearAuth) {
            uni.showToast({
              title: `HTTP错误: ${res.statusCode}`,
              icon: 'none'
            });
          }
          reject(res);
        }
      },
      fail: (err) => {
        if (!silent) {
          uni.showToast({
            title: '网络连接失败',
            icon: 'none'
          });
        }
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
      const pages = getCurrentPages();
      const isIndexPage = pages.length > 0 && (() => {
        const r = (pages[pages.length - 1].route || '').replace(/^\//, '');
        return r === 'pages/index/index';
      })();
      if (isIndexPage) {
        // 在首页：优先通过 $vm 调用 onShow，否则用全局事件兜底（$vm 在小程序端可能不可用）
        const cur = pages[pages.length - 1];
        const vm = cur.$vm || cur;
        if (vm && typeof vm.checkAndAutoLogin === 'function') {
          vm.checkAndAutoLogin();
        } else if (vm && typeof vm.onShow === 'function') {
          vm.onShow();
        } else {
          uni.$emit('auth:401'); // 兜底：首页需 uni.$on('auth:401') 监听
        }
      } else {
        uni.reLaunch({ url: '/pages/index/index' });
      }
    }, 100);
  } catch (error) {
    console.error('清空本地信息失败:', error);
  }
}

// 检查状态码是否需要清空登录信息
// silent 为 true 时（如轮询订单详情）404 不应清空登录，避免误跳首页
// ignoreAuthClear 为 true 时，不触发清空登录（用于个别非关键接口的临时兼容）
function shouldClearAuthInfo(statusCode, businessCode, silent, ignoreAuthClear) {
  if (ignoreAuthClear) return false;
  if (silent && statusCode === 404) return false;
  // HTTP状态码：401未授权、403禁止
  if (statusCode === 401 || statusCode === 403) return true;
  // HTTP 404 且非静默：如会话失效导致的资源不存在
  if (statusCode === 404) return true;
  if (businessCode === 401 || businessCode === 404) return true;
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