CREATE DATABASE edgelink;

-- 物模型表
CREATE TABLE `thing_model` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `identifier` varchar(64) NOT NULL COMMENT '模型唯一标识',
  `name` varchar(128) NOT NULL COMMENT '模型名称',
  `description` varchar(255) DEFAULT NULL,
  `created_time` datetime DEFAULT NULL,
  `updated_time` datetime DEFAULT NULL,
  `icon` text COMMENT '图标',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_model_identifier` (`identifier`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='物模型定义';

-- 模型属性表
CREATE TABLE `thing_model_property` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `model_id` bigint NOT NULL,
  `key` varchar(64) NOT NULL COMMENT '数据key',
  `name` varchar(128) NOT NULL COMMENT '属性名称',
  `data_type` smallint NOT NULL COMMENT '数据类型，1:bool, 2:int, 3:float',
  `unit` varchar(32) DEFAULT NULL COMMENT '单位',
  `source_type` smallint DEFAULT '1' COMMENT '来源类型，1:raw, 2:formula',
  `expr` text COMMENT '公式/聚合表达式',
  `created_time` datetime DEFAULT NULL,
  `updated_time` datetime DEFAULT NULL,
  `type` smallint NOT NULL COMMENT '属性类型，1property,2:function,3event',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_model_id_key` (`model_id`,`key`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='物模型属性定义';

-- 产品表
CREATE TABLE `product` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `identifier` varchar(64) NOT NULL,
  `name` varchar(128) NOT NULL,
  `thing_model_id` bigint NOT NULL COMMENT '绑定物模型id',
  `created_time` datetime DEFAULT NULL,
  `updated_time` datetime DEFAULT NULL,
  `protocol` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_identifier` (`identifier`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='产品';

-- 设备表
CREATE TABLE `device` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `key` varchar(64) NOT NULL,
  `name` varchar(128) DEFAULT NULL,
  `product_id` bigint NOT NULL,
  `address` json NOT NULL COMMENT '设备上行/下行地址',
  `description` varchar(255) DEFAULT NULL COMMENT '设备描述',
  `created_time` datetime DEFAULT NULL,
  `updated_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_device_key` (`key`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='设备表';

-- 设备属性表
CREATE TABLE `device_property_ref` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `device_id` bigint NOT NULL,
  `property_id` bigint NOT NULL COMMENT '物模型属性id',
  `persistent` tinyint(1) DEFAULT '1' COMMENT '是否存历史数据',
  `store_mode` varchar(16) DEFAULT 'minute' COMMENT '数据存储方式，full全量存储,change值变化才存储, minute整点数据',
  `created_time` datetime DEFAULT NULL,
  `updated_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_device_property` (`device_id`,`property_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='设备属性关系表';

-- 历史数据表
CREATE TABLE `history_data` (
  `device_id` bigint NOT NULL,
  `property_id` bigint NOT NULL,
  `ts` datetime NOT NULL,
  `value` double DEFAULT NULL,
  PRIMARY KEY (`device_id`,`property_id`,`ts`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='设备历史数据';