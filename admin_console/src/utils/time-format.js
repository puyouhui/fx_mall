// 时间格式化工具函数

/**
 * 将ISO格式时间或日期对象转换为指定格式
 * @param {string|Date} date - 日期字符串或Date对象
 * @param {string} format - 格式化模板，默认为 'YYYY-MM-DD HH:mm:ss'
 * @returns {string} 格式化后的时间字符串
 */
export function formatDate(date, format = 'YYYY-MM-DD HH:mm:ss') {
  if (!date) return '';
  
  // 确保date是Date对象
  let d;
  if (typeof date === 'string') {
    // 处理ISO格式的时间字符串，如：2025-09-20T13:12:11+08:00
    d = new Date(date);
  } else if (date instanceof Date) {
    d = date;
  } else {
    return '';
  }
  
  // 如果是无效的日期，返回空字符串
  if (isNaN(d.getTime())) return '';
  
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  const hours = String(d.getHours()).padStart(2, '0');
  const minutes = String(d.getMinutes()).padStart(2, '0');
  const seconds = String(d.getSeconds()).padStart(2, '0');
  
  return format
    .replace('YYYY', year)
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds);
}

/**
 * 为Element Plus的表格组件提供时间格式化的过滤器
 * @param {string|Date} date - 日期字符串或Date对象
 * @returns {string} 格式化后的时间字符串
 */
export function tableDateFormat(date) {
  return formatDate(date);
}

/**
 * 递归遍历对象，自动格式化所有时间字段
 * @param {object} data - 要处理的数据对象
 * @param {array} timeFields - 时间字段列表
 * @returns {object} 处理后的数据对象
 */
export function formatTimeFields(data, timeFields = ['created_at', 'updated_at', 'createdAt', 'updatedAt']) {
  if (!data || typeof data !== 'object') {
    return data;
  }
  
  // 如果是数组，递归处理每个元素
  if (Array.isArray(data)) {
    return data.map(item => formatTimeFields(item, timeFields));
  }
  
  // 复制对象，避免直接修改原对象
  const result = { ...data };
  
  // 遍历对象的所有属性
  Object.keys(result).forEach(key => {
    // 如果是时间字段，格式化它
    if (timeFields.includes(key)) {
      result[key] = formatDate(result[key]);
    }
    // 如果是嵌套对象，递归处理
    else if (result[key] && typeof result[key] === 'object') {
      result[key] = formatTimeFields(result[key], timeFields);
    }
  });
  
  return result;
}