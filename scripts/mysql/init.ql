CREATE DATABASE edgelink;

CREATE TABLE `device` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `device_key` VARCHAR(64) NOT NULL COMMENT '设备唯一标识',
    `device_name` VARCHAR(128) NOT NULL COMMENT '设备名称',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '设备描述',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '所属用户',
    `created_time` DATETIME,
    `updated_time` DATETIME,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_device_key` (`device_key`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备表';


CREATE TABLE `device_point` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `device_id` BIGINT UNSIGNED NOT NULL COMMENT '所属设备 ID',
    `origin_field` VARCHAR(64) NOT NULL COMMENT '设备原始字段',
    `name` VARCHAR(128) NOT NULL COMMENT '点位名称',
    `point_id` VARCHAR(128) COMMENT '全局点位 ID（用于数据存储）',
    `unit` VARCHAR(32) DEFAULT NULL COMMENT '单位',
    `persistent` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否持久化历史数据1',
    `created_time` DATETIME,
    `updated_time` DATETIME,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_point_id` (`point_id`),
    UNIQUE KEY `uk_device_field` (`device_id`, `origin_field`),
    KEY `idx_device_id` (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备点位定义表';


CREATE TABLE `history_data` (
    `point_id` BIGINT UNSIGNED NOT NULL COMMENT '点位 ID',
    `ts` BIGINT UNSIGNED NOT NULL COMMENT '时间戳（秒或毫秒，需统一）',
    `value` DOUBLE NOT NULL COMMENT '点位数值',
    PRIMARY KEY (`point_id`, `ts`),
    KEY `idx_ts` (`ts`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备点位历史数据';


CREATE TABLE device_group (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    parent_id BIGINT DEFAULT 0,
    group_name VARCHAR(64) NOT NULL,
    description VARCHAR(256),
    user_id BIGINT,
    created_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_parent_id (parent_id)
);