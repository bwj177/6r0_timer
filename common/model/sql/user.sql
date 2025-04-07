CREATE TABLE IF NOT EXISTS `user`
(
    `id`                bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_name`               varchar(255) NOT NULL COMMENT '用户名',
    `password`              varchar(255) NOT NULL COMMENT '密码',
    `deleted_at`        datetime     DEFAULT NULL COMMENT '删除时间',
    `created_at`        datetime     NOT NULL COMMENT '创建时间',
    `updated_at`        datetime     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_app` (`user_name`) USING BTREE COMMENT 'user_name 索引'
    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;