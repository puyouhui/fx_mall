-- 添加title字段到carousels表
ALTER TABLE carousels
ADD COLUMN title VARCHAR(255) DEFAULT '' COMMENT '轮播图标题' AFTER image;

-- 更新查询语句相关的函数以包含title字段