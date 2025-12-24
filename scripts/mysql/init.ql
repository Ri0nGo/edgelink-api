CREATE DATABASE edgelink;

-- 物模型表
CREATE TABLE thing_model (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  model_key VARCHAR(64) NOT NULL COMMENT '模型唯一标识',
  model_name VARCHAR(128) NOT NULL COMMENT '模型名称',
  description VARCHAR(255),
  created_time DATETIME,
  updated_time DATETIME,
  UNIQUE KEY uk_model_key (model_key)
) COMMENT='物模型定义';


-- 产品表
CREATE TABLE product (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    product_key VARCHAR(64) NOT NULL,
    product_name VARCHAR(128) NOT NULL,
    model_id BIGINT NOT NULL COMMENT '绑定物模型',
    created_time DATETIME,
    updated_time DATETIME,
    UNIQUE KEY uk_product_key (product_key)
) COMMENT='产品';

-- 设备表
CREATE TABLE device (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    device_key VARCHAR(64) NOT NULL,
    device_name VARCHAR(128),
    product_id BIGINT NOT NULL,
    protocol SMALLINT NOT NULL COMMENT '协议类型，1:mqtt',
    address      json         not null comment '设备上行/下行地址',
    description VARCHAR(255) DEFAULT NULL COMMENT '设备描述',
    created_time DATETIME,
    updated_time DATETIME,
    UNIQUE KEY uk_device_key (device_key)
) COMMENT='设备表';

-- 设备属性表
CREATE TABLE device_property_ref (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    device_id BIGINT NOT NULL,
    property_id BIGINT NOT NULL,
    persistent TINYINT(1) DEFAULT 1 COMMENT '是否存历史数据',
    store_mode ENUM('full','change','aggregate') DEFAULT 'full',
    created_time DATETIME,
    updated_time DATETIME,
    UNIQUE KEY uk_device_property (device_id, property_id)
) COMMENT='设备属性实例配置';

-- 历史数据表
CREATE TABLE history_data (
    device_id BIGINT NOT NULL,
    property_id BIGINT NOT NULL,
    ts DATETIME NOT NULL,
    value DOUBLE,
    PRIMARY KEY (device_id, property_id, ts)
) COMMENT='设备历史数据';
