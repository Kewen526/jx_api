/*
 Navicat Premium Data Transfer

 Source Server         : jiangxin
 Source Server Type    : MySQL
 Source Server Version : 80044
 Source Host           : 8.146.210.145:3306
 Source Schema         : jx_data_info

 Target Server Type    : MySQL
 Target Server Version : 80044
 File Encoding         : 65001

 Date: 27/02/2026 02:47:25
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for review_detail_dianping
-- ----------------------------
DROP TABLE IF EXISTS `review_detail_dianping`;
CREATE TABLE `review_detail_dianping`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `review_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '评价ID',
  `shop_id` bigint NOT NULL COMMENT '门店ID',
  `shop_name` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '门店名称',
  `city_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '城市',
  `city_id` int NULL DEFAULT NULL COMMENT '城市ID',
  `user_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户ID',
  `user_nickname` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户昵称',
  `user_face` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户头像URL',
  `user_power` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户等级',
  `vip_level` int NULL DEFAULT NULL COMMENT 'VIP等级',
  `add_time` datetime NULL DEFAULT NULL COMMENT '评价时间',
  `update_time` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `edit_time` datetime NULL DEFAULT NULL COMMENT '编辑时间',
  `star` int NULL DEFAULT NULL COMMENT '星级(原始值,50=5星)',
  `star_display` decimal(2, 1) NULL DEFAULT NULL COMMENT '星级(1-5)',
  `accurate_star` int NULL DEFAULT NULL COMMENT '精确星级',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '评价内容',
  `score_technician` decimal(2, 1) NULL DEFAULT NULL COMMENT '技师评分',
  `score_service` decimal(2, 1) NULL DEFAULT NULL COMMENT '服务评分',
  `score_environment` decimal(2, 1) NULL DEFAULT NULL COMMENT '环境评分',
  `score_map` json NULL COMMENT '评分详情JSON',
  `pic_count` int NULL DEFAULT 0 COMMENT '图片数量',
  `video_count` int NULL DEFAULT 0 COMMENT '视频数量',
  `pic_info` json NULL COMMENT '图片信息[{pic_id,pic_url,origin_pic_url}]',
  `video_info` json NULL COMMENT '视频信息',
  `shop_reply` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '商家回复内容(最新一条)',
  `shop_reply_time` datetime NULL DEFAULT NULL COMMENT '商家回复时间',
  `is_reply_with_photo` tinyint NULL DEFAULT 0 COMMENT '回复是否带图',
  `reply_list` json NULL COMMENT '所有回复列表',
  `order_id` bigint NULL DEFAULT NULL COMMENT '订单ID',
  `deal_group_id` bigint NULL DEFAULT NULL COMMENT '团购ID',
  `refer_type` int NULL DEFAULT NULL COMMENT '来源类型',
  `avg_price` int NULL DEFAULT NULL COMMENT '人均价格',
  `serial_numbers` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '券号',
  `total_cost` decimal(12, 2) NULL DEFAULT NULL COMMENT '总价',
  `consume_date` date NULL DEFAULT NULL COMMENT '消费日期',
  `status` int NULL DEFAULT NULL COMMENT '评价状态',
  `quality_score` int NULL DEFAULT NULL COMMENT '质量分',
  `case_status` int NULL DEFAULT NULL COMMENT '投诉状态',
  `case_status_desc` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '投诉状态描述',
  `report_status` int NULL DEFAULT NULL COMMENT '举报状态',
  `report_status_desc` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '举报状态描述',
  `case_id` bigint NULL DEFAULT NULL COMMENT '投诉ID',
  `show_deal` tinyint NULL DEFAULT NULL COMMENT '是否显示交易',
  `raw_data` json NULL COMMENT '原始JSON数据',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `ai_gen` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT 'AI生成内容',
  `manual_confirm` int NOT NULL DEFAULT 0 COMMENT '人工确认 0-未确认 1-已确认',
  `task_reply` int NOT NULL DEFAULT 0 COMMENT '任务回复 0-未回复 1-已回复 2-回复失败',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_review_id`(`review_id` ASC) USING BTREE,
  INDEX `idx_shop_id`(`shop_id` ASC) USING BTREE,
  INDEX `idx_add_time`(`add_time` ASC) USING BTREE,
  INDEX `idx_star`(`star` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 9672 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '点评评价详情表' ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
