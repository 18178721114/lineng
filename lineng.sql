/*
 Navicat Premium Data Transfer

 Source Server         : 127.0.0.1
 Source Server Type    : MySQL
 Source Server Version : 80046 (8.0.46)
 Source Host           : 127.0.0.1:3306
 Source Schema         : lineng

 Target Server Type    : MySQL
 Target Server Version : 80046 (8.0.46)
 File Encoding         : 65001

 Date: 31/10/2026 21:49:23
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for daily_reports
-- ----------------------------
DROP TABLE IF EXISTS `daily_reports`;
CREATE TABLE `daily_reports` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `project_id` bigint unsigned NOT NULL COMMENT '所属项目 ID',
  `report_date` date NOT NULL COMMENT '出勤日期',
  `is_outsourced` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否外包：1 是，0 否',
  `notes` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_date` (`project_id`,`report_date`),
  UNIQUE KEY `uk_proj_date` (`project_id`,`report_date`),
  KEY `idx_date` (`report_date`),
  CONSTRAINT `fk_report_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日报主表';

-- ----------------------------
-- Records of daily_reports
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for project_prices
-- ----------------------------
DROP TABLE IF EXISTS `project_prices`;
CREATE TABLE `project_prices` (
  `project_id` bigint unsigned NOT NULL,
  `labor_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '人工单价，单位 元/人',
  `meal_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '餐费单价，单位 元/份',
  `water_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '水单价，单位 元/份',
  `fork3_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '叉车 3 吨单价，单位 元/台班',
  `fork5_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '叉车 5 吨单价，单位 元/台班',
  `fork7_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '叉车 7 吨单价，单位 元/台班',
  `crane_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '吊车单价，单位 元/台班',
  `rescue_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '救援车单价，单位 元/台班',
  `tax_rate` decimal(5,2) NOT NULL DEFAULT '6.00' COMMENT '税率百分比，6 表示 6%',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`project_id`),
  CONSTRAINT `fk_prices_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目默认单价';

-- ----------------------------
-- Records of project_prices
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for projects
-- ----------------------------
DROP TABLE IF EXISTS `projects`;
CREATE TABLE `projects` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `no` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '单据编号（LN-YYYYMMDD-XXX）',
  `unit` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用工单位',
  `activity_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '活动名称',
  `location` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用工地点',
  `party_a_owner` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '甲方负责人',
  `party_a_phone` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '甲方联系电话',
  `start_date` date DEFAULT NULL COMMENT '用工开始日期',
  `end_date` date DEFAULT NULL COMMENT '用工结束日期',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_projects_no` (`no`),
  KEY `idx_period` (`start_date`,`end_date`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目（出勤表表头信息）';

-- ----------------------------
-- Records of projects
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for report_equipment
-- ----------------------------
DROP TABLE IF EXISTS `report_equipment`;
CREATE TABLE `report_equipment` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `report_id` bigint unsigned NOT NULL COMMENT '所属日报 ID',
  `equip_type` enum('FORK3','FORK5','FORK7','CRANE','RESCUE') COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '设备类型：FORK3 叉车3吨 / FORK5 叉车5吨 / FORK7 叉车7吨 / CRANE 吊车 / RESCUE 救援车',
  `shift_type` enum('DAY','EVENING','DAWN') COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '班次：DAY 白天 / EVENING 晚上 / DAWN 凌晨',
  `qty` bigint NOT NULL DEFAULT '0' COMMENT '台班数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_report_equip_shift` (`report_id`,`equip_type`,`shift_type`),
  UNIQUE KEY `uk_res` (`report_id`,`equip_type`,`shift_type`),
  CONSTRAINT `fk_re_report` FOREIGN KEY (`report_id`) REFERENCES `daily_reports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=106 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='④ 设备台班';

-- ----------------------------
-- Records of report_equipment
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for report_meals
-- ----------------------------
DROP TABLE IF EXISTS `report_meals`;
CREATE TABLE `report_meals` (
  `report_id` bigint unsigned NOT NULL,
  `breakfast_qty` bigint NOT NULL DEFAULT '0' COMMENT '早点份数',
  `lunch_qty` bigint NOT NULL DEFAULT '0' COMMENT '中餐份数',
  `dinner_qty` bigint NOT NULL DEFAULT '0' COMMENT '晚餐份数',
  `supper_qty` bigint NOT NULL DEFAULT '0' COMMENT '夜宵份数',
  `water_qty` bigint NOT NULL DEFAULT '0' COMMENT '水份数',
  PRIMARY KEY (`report_id`),
  CONSTRAINT `fk_rm_report` FOREIGN KEY (`report_id`) REFERENCES `daily_reports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='③ 餐费';

-- ----------------------------
-- Records of report_meals
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for report_module_drafts
-- ----------------------------
DROP TABLE IF EXISTS `report_module_drafts`;
CREATE TABLE `report_module_drafts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `project_id` bigint unsigned NOT NULL COMMENT '所属项目 ID',
  `report_date` date NOT NULL COMMENT '出勤日期',
  `module_key` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模块键（names/labor/meal/fork/out/note）',
  `content` json NOT NULL COMMENT '模块暂存内容（JSON 对象，字段名→值）',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_draft` (`project_id`,`report_date`,`module_key`),
  CONSTRAINT `fk_rmd_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='⑥ 模块暂存草稿';

-- ----------------------------
-- Records of report_module_drafts
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for report_outsourcing
-- ----------------------------
DROP TABLE IF EXISTS `report_outsourcing`;
CREATE TABLE `report_outsourcing` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `report_id` bigint unsigned NOT NULL COMMENT '所属日报 ID',
  `unit_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '外包单位名称',
  `day_count` bigint NOT NULL DEFAULT '0' COMMENT '白天人数',
  `evening_count` bigint NOT NULL DEFAULT '0' COMMENT '晚上人数',
  `dawn_count` bigint NOT NULL DEFAULT '0' COMMENT '凌晨人数',
  `day_ot_count` bigint NOT NULL DEFAULT '0' COMMENT '白班加班人数',
  `day_ot_start` time DEFAULT NULL COMMENT '白班加班开始时间，可空',
  `day_ot_end` time DEFAULT NULL COMMENT '白班加班结束时间，可空',
  `eve_ot_count` smallint unsigned NOT NULL DEFAULT '0' COMMENT '晚班加班人数',
  `eve_ot_start` time DEFAULT NULL COMMENT '晚班加班开始，可空',
  `eve_ot_end` time DEFAULT NULL COMMENT '晚班加班结束，可空',
  `dawn_ot_count` bigint NOT NULL DEFAULT '0' COMMENT '凌晨加班人数',
  `dawn_ot_start` time DEFAULT NULL COMMENT '凌晨加班开始时间，可空',
  `dawn_ot_end` time DEFAULT NULL COMMENT '凌晨加班结束时间，可空',
  `ot_start` time DEFAULT NULL COMMENT '旧版通用加班开始时间（历史兼容，不再写入）',
  `ot_end` time DEFAULT NULL COMMENT '旧版通用加班结束时间（历史兼容，不再写入）',
  `overnight` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否跨夜：1 表示加班结束时间在次日',
  `sort_order` bigint unsigned NOT NULL DEFAULT '0' COMMENT '排列顺序',
  `evening_ot_count` bigint NOT NULL DEFAULT '0' COMMENT '晚班加班人数',
  `evening_ot_start` time DEFAULT NULL COMMENT '晚班加班开始时间，可空',
  `evening_ot_end` time DEFAULT NULL COMMENT '晚班加班结束时间，可空',
  PRIMARY KEY (`id`),
  KEY `idx_report` (`report_id`),
  KEY `idx_report_outsourcing_report_id` (`report_id`),
  CONSTRAINT `fk_ro_report` FOREIGN KEY (`report_id`) REFERENCES `daily_reports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='⑤ 外包';

-- ----------------------------
-- Records of report_outsourcing
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for report_shifts
-- ----------------------------
DROP TABLE IF EXISTS `report_shifts`;
CREATE TABLE `report_shifts` (
  `report_id` bigint unsigned NOT NULL,
  `day_count` bigint NOT NULL DEFAULT '0' COMMENT '白天人数',
  `evening_count` bigint NOT NULL DEFAULT '0' COMMENT '晚上人数',
  `dawn_count` bigint NOT NULL DEFAULT '0' COMMENT '凌晨人数',
  PRIMARY KEY (`report_id`),
  CONSTRAINT `fk_rs_report` FOREIGN KEY (`report_id`) REFERENCES `daily_reports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='② 用工人数';

-- ----------------------------
-- Records of report_shifts
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for report_workers
-- ----------------------------
DROP TABLE IF EXISTS `report_workers`;
CREATE TABLE `report_workers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `report_id` bigint unsigned NOT NULL COMMENT '所属日报 ID',
  `worker_id` bigint unsigned NOT NULL COMMENT '人员 ID（关联 workers.id）',
  `worker_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '下单时的姓名快照，人员改名不影响历史单据',
  `start_time` time NOT NULL COMMENT '上班时间 HH:MM',
  `end_time` time NOT NULL COMMENT '下班时间 HH:MM',
  `overnight` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否跨夜：1 表示下班时间在次日',
  `sort_order` bigint unsigned NOT NULL DEFAULT '0' COMMENT '名单内排序序号（决定考勤表行序）',
  PRIMARY KEY (`id`),
  KEY `idx_report` (`report_id`),
  KEY `idx_worker` (`worker_id`),
  KEY `idx_report_workers_report_id` (`report_id`),
  KEY `idx_report_workers_worker_id` (`worker_id`),
  CONSTRAINT `fk_rw_report` FOREIGN KEY (`report_id`) REFERENCES `daily_reports` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_rw_worker` FOREIGN KEY (`worker_id`) REFERENCES `workers` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='① 用工名单';

-- ----------------------------
-- Records of report_workers
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `username` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '登录名',
  `password` longtext COLLATE utf8mb4_unicode_ci COMMENT '登录密码（bcrypt 哈希）',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员账号';

-- ----------------------------
-- Records of users
-- ----------------------------
BEGIN;
INSERT INTO `users` (`id`, `username`, `password`, `created_at`) VALUES (1, 'admin', '$2a$10$bUM1XJL0yW0K4tJ0PD1IGuthjRp3CVJCP5l224JeceGUcum0DZlLu', '2026-10-28 23:40:17.680');
COMMIT;

-- ----------------------------
-- Table structure for workers
-- ----------------------------
DROP TABLE IF EXISTS `workers`;
CREATE TABLE `workers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '姓名（唯一）',
  `phone` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `job_type` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工种',
  `note` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `pinyin_full` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '姓名全拼，小写无空格，用于搜索',
  `pinyin_initial` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '姓名拼音首字母，用于搜索',
  `active` tinyint(1) DEFAULT '1' COMMENT '是否在职：1 在职，0 停用',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `company` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '所属公司（员工表导入）',
  `id_card` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '身份证号（员工表导入）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_workers_name` (`name`),
  KEY `idx_workers_pinyin_full` (`pinyin_full`),
  KEY `idx_workers_pinyin_initial` (`pinyin_initial`)
) ENGINE=InnoDB AUTO_INCREMENT=156 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='人员档案';

-- ----------------------------
-- Records of workers
-- ----------------------------
BEGIN;
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (8, '常晓龙', '', '', '', 'changxiaolong', 'cxl', 1, '2026-10-29 11:27:13.012', '力楷', '130726198908216117');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (9, '陈亚斌', '', '', '', 'chenyabin', 'cyb', 1, '2026-10-29 11:27:13.017', '力楷', '411403198611195818');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (10, '陈月鹏', '', '', '', 'chenyuepeng', 'cyp', 1, '2026-10-29 11:27:13.018', '艺德利', '411403199401025815');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (11, '代海军', '', '', '', 'daihaijun', 'dhj', 1, '2026-10-29 11:27:13.019', '艺德利', '622628199407184671');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (12, '代仁海', '', '', '', 'dairenhai', 'drh', 1, '2026-10-29 11:27:13.021', '艺德利', '622628199009064674');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (13, '代小蹦', '', '', '', 'daixiaobeng', 'dxb', 1, '2026-10-29 11:27:13.021', '瀚楷', '412725198509263031');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (14, '豆官政', '', '', '', 'douguanzheng', 'dgz', 1, '2026-10-29 11:27:13.023', '力能', '411424199909070512');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (15, '段宝源', '', '', '', 'duanbaoyuan', 'dby', 1, '2026-10-29 11:27:13.024', '瀚楷', '130726199011092812');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (16, '段世义', '', '', '', 'duanshiyi', 'dsy', 1, '2026-10-29 11:27:13.025', '力楷', '371421198811174632');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (17, '范长水', '', '', '', 'fanzhangshui', 'fzs', 1, '2026-10-29 11:27:13.026', '力楷', '622628197701184673');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (18, '付圣忠', '', '', '', 'fushengzhong', 'fsz', 1, '2026-10-29 11:27:13.027', '艺德利', '372421197203102973');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (19, '郭龙宝', '', '', '', 'guolongbao', 'glb', 1, '2026-10-29 11:27:13.028', '力楷', '371421197906132238');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (20, '郭伟刚', '', '', '', 'guoweigang', 'gwg', 1, '2026-10-29 11:27:13.029', '力楷', '41052619770922487X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (21, '韩雪奎', '', '', '', 'hanxuekui', 'hxk', 1, '2026-10-29 11:27:13.030', '瀚楷', '622628199605284497');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (22, '何飞', '', '', '', 'hefei', 'hf', 1, '2026-10-29 11:27:13.032', '瀚楷', '513022199008131518');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (23, '侯丁', '', '', '', 'houding', 'hd', 1, '2026-10-29 11:27:13.034', '力楷', '142622199302127512');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (24, '黄金祥', '', '', '', 'huangjinxiang', 'hjx', 1, '2026-10-29 11:27:13.034', '艺德利', '622628198903154696');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (25, '蒋文辉', '', '', '', 'jiangwenhui', 'jwh', 1, '2026-10-29 11:27:13.036', '瀚楷', '62262819981110467X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (26, '李代哲', '', '', '', 'lidaizhe', 'ldz', 1, '2026-10-29 11:27:13.039', '力能', '12022120020818351X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (27, '李飞', '', '', '', 'lifei', 'lf', 1, '2026-10-29 11:27:13.041', '力楷', '130635198101275133');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (28, '李高贵', '', '', '', 'ligaogui', 'lgg', 1, '2026-10-29 11:27:13.042', '艺德利', '622628198606104676');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (29, '李会军', '', '', '', 'lihuijun', 'lhj', 1, '2026-10-29 11:27:13.044', '力楷', '622628199005144677');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (30, '李晶鑫', '', '', '', 'lijingxin', 'ljx', 1, '2026-10-29 11:27:13.045', '力楷', '230128199202281257');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (31, '李玲玲', '', '', '', 'lilingling', 'lll', 1, '2026-10-29 11:27:13.046', '瀚楷', '622628199607104672');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (32, '李奇', '', '', '', 'liqi', 'lq', 1, '2026-10-29 11:27:13.047', '瀚楷', '622628199907164677');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (33, '李献伟', '', '', '', 'lixianwei', 'lxw', 1, '2026-10-29 11:27:13.048', '力楷', '411424198510075456');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (34, '李晓东', '', '', '', 'lixiaodong', 'lxd', 1, '2026-10-29 11:27:13.049', '艺德利', '132525197801200035');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (35, '李扬', '', '', '', 'liyang', 'ly', 1, '2026-10-29 11:27:13.051', '力楷', '120225198209190493');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (36, '李义生', '', '', '', 'liyisheng', 'lys', 1, '2026-10-29 11:27:13.052', '力楷', '412327197212305479');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (37, '李引军', '', '', '', 'liyinjun', 'lyj', 1, '2026-10-29 11:27:13.053', '艺德利', '622628199510104692');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (38, '李现召', '', '', '', 'lixianzhao', 'lxz', 1, '2026-10-29 11:27:13.054', '力楷', '412327198310205476');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (39, '李志军', '', '', '', 'lizhijun', 'lzj', 1, '2026-10-29 11:27:13.055', '力楷', '150430198909223417');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (40, '梁俊合', '', '', '', 'liangjunhe', 'ljh', 1, '2026-10-29 11:27:13.057', '瀚楷', '142621198209251032');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (41, '梁镇伟', '', '', '', 'liangzhenwei', 'lzw', 1, '2026-10-29 11:27:13.058', '艺德利', '41272619860816123X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (42, '刘兵', '', '', '', 'liubing', 'lb', 1, '2026-10-29 11:27:13.059', '艺德利', '622628199804064470');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (43, '刘朝刚', '', '', '', 'liuchaogang', 'lcg', 1, '2026-10-29 11:27:13.060', '力楷', '410526198601273770');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (44, '刘晨辉', '', '', '', 'liuchenhui', 'lch', 1, '2026-10-29 11:27:13.061', '力楷', '41052620040109003X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (45, '刘冬', '', '', '', 'liudong', 'ld', 1, '2026-10-29 11:27:13.061', '力楷', '130726199009152812');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (46, '刘高林', '', '', '', 'liugaolin', 'lgl', 1, '2026-10-29 11:27:13.062', '力楷', '41272519870328221X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (47, '刘衡', '', '', '', 'liuheng', 'lh', 1, '2026-10-29 11:27:13.063', '力能', '371526199202070813');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (48, '刘龙龙', '', '', '', 'liulonglong', 'lll', 1, '2026-10-29 11:27:13.064', '力能', '130529199111113818');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (49, '刘明超', '', '', '', 'liumingchao', 'lmc', 1, '2026-10-29 11:27:13.065', '力能', '371526198706250450');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (50, '刘宁', '', '', '', 'liuning', 'ln', 1, '2026-10-29 11:27:13.066', '力楷', '371526198712103713');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (51, '刘庆强', '', '', '', 'liuqingqiang', 'lqq', 1, '2026-10-29 11:27:13.066', '力楷', '371526198708173719');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (52, '刘少聪', '', '', '', 'liushaocong', 'lsc', 1, '2026-10-29 11:27:13.067', '力楷', '410526199210013896');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (53, '刘守飞', '', '', '', 'liushoufei', 'lsf', 1, '2026-10-29 11:27:13.068', '艺德利', '412726197904101210');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (54, '刘岁邦', '', '', '', 'liusuibang', 'lsb', 1, '2026-10-29 11:27:13.068', '艺德利', '622628197803214693');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (55, '路传宾', '', '', '', 'luchuanbin', 'lcb', 1, '2026-10-29 11:27:13.069', '艺德利', '412726197712301234');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (56, '马郭宁', '', '', '', 'maguoning', 'mgn', 1, '2026-10-29 11:27:13.070', '瀚楷', '622628199804104479');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (57, '马甲坤', '', '', '', 'majiakun', 'mjk', 1, '2026-10-29 11:27:13.071', '力楷', '372923198707120838');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (58, '马江迎', '', '', '', 'majiangying', 'mjy', 1, '2026-10-29 11:27:13.073', '瀚楷', '622628198601124678');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (59, '马坤磊', '', '', '', 'makunlei', 'mkl', 1, '2026-10-29 11:27:13.074', '力楷', '411424199808218418');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (60, '马宁', '', '', '', 'maning', 'mn', 1, '2026-10-29 11:27:13.075', '力楷', '141022199009130014');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (61, '秦洋洋', '', '', '', 'qinyangyang', 'qyy', 1, '2026-10-29 11:27:13.075', '力楷', '14262519930714111X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (62, '邱建豪', '', '', '', 'qiujianhao', 'qjh', 1, '2026-10-29 11:27:13.076', '艺德利', '411628200008063034');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (63, '任爱军', '', '', '', 'renaijun', 'raj', 1, '2026-10-29 11:27:13.077', '力楷', '62262819921011467X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (64, '任广民', '', '', '', 'renguangmin', 'rgm', 1, '2026-10-29 11:27:13.078', '力能', '371421197605052218');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (65, '任现东', '', '', '', 'renxiandong', 'rxd', 1, '2026-10-29 11:27:13.079', '艺德利', '372421197911242213');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (66, '任宪超', '', '', '', 'renxianchao', 'rxc', 1, '2026-10-29 11:27:13.080', '艺德利', '37142120060815197X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (67, '宋军军', '', '', '', 'songjunjun', 'sjj', 1, '2026-10-29 11:27:13.081', '瀚楷', '371526198812033716');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (68, '苏永光', '', '', '', 'suyongguang', 'syg', 1, '2026-10-29 11:27:13.082', '力楷', '412725198207103032');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (69, '孙爱国', '', '', '', 'sunaiguo', 'sag', 1, '2026-10-29 11:27:13.084', '艺德利', '412725198012183036');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (70, '孙保良', '', '', '', 'sunbaoliang', 'sbl', 1, '2026-10-29 11:27:13.084', '艺德利', '41272519760529305X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (71, '孙成钢', '', '', '', 'sunchenggang', 'scg', 1, '2026-10-29 11:27:13.085', '力楷', '371421199210082219');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (72, '孙海波', '', '', '', 'sunhaibo', 'shb', 1, '2026-10-29 11:27:13.087', '力楷', '412725199412243039');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (73, '孙军令', '', '', '', 'sunjunling', 'sjl', 1, '2026-10-29 11:27:13.087', '艺德利', '412725199604073214');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (74, '孙天文', '', '', '', 'suntianwen', 'stw', 1, '2026-10-29 11:27:13.088', '艺德利', '130635197504204810');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (75, '孙永纯', '', '', '', 'sunyongchun', 'syc', 1, '2026-10-29 11:27:13.089', '艺德利', '412725197708043010');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (76, '孙玉东', '', '', '', 'sunyudong', 'syd', 1, '2026-10-29 11:27:13.090', '艺德利', '412725196608043051');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (77, '索军伙', '', '', '', 'suojunhuo', 'sjh', 1, '2026-10-29 11:27:13.091', '力能', '622628198811154672');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (78, '索军军', '', '', '', 'suojunjun', 'sjj', 1, '2026-10-29 11:27:13.092', '力楷', '622628199302144673');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (79, '唐小冲', '', '', '', 'tangxiaochong', 'txc', 1, '2026-10-29 11:27:13.092', '艺德利', '622628199302123530');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (80, '陶艳松', '', '', '', 'taoyansong', 'tys', 1, '2026-10-29 11:27:13.094', '力能', '371421198904211973');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (81, '王爱平', '', '', '', 'wangaiping', 'wap', 1, '2026-10-29 11:27:13.095', '艺德利', '622628198107094696');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (82, '王东红', '', '', '', 'wangdonghong', 'wdh', 1, '2026-10-29 11:27:13.096', '力楷', '622628199602174671');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (83, '王杜顺', '', '', '', 'wangdushun', 'wds', 1, '2026-10-29 11:27:13.096', '瀚楷', '622628198408104675');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (84, '王福勤', '', '', '', 'wangfuqin', 'wfq', 1, '2026-10-29 11:27:13.097', '力楷', '622628198405124670');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (85, '王吉良', '', '', '', 'wangjiliang', 'wjl', 1, '2026-10-29 11:27:13.098', '力能', '371526198901143714');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (86, '王吉振', '', '', '', 'wangjizhen', 'wjz', 1, '2026-10-29 11:27:13.099', '瀚楷', '371421198804102219');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (87, '王建银', '', '', '', 'wangjianyin', 'wjy', 1, '2026-10-29 11:27:13.099', '力能', '371526199005012470');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (88, '王金龙', '', '', '', 'wangjinlong', 'wjl', 1, '2026-10-29 11:27:13.100', '瀚楷', '622628198810014491');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (89, '王金仲', '', '', '', 'wangjinzhong', 'wjz', 1, '2026-10-29 11:27:13.101', '艺德利', '622628198005164673');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (90, '王群豪', '', '', '', 'wangqunhao', 'wqh', 1, '2026-10-29 11:27:13.102', '瀚楷', '411628200404083053');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (91, '王瑞兵', '', '', '', 'wangruibing', 'wrb', 1, '2026-10-29 11:27:13.102', '瀚楷', '622628199709114679');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (92, '王腾飞', '', '', '', 'wangtengfei', 'wtf', 1, '2026-10-29 11:27:13.103', '力楷', '41162819950912301X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (93, '王文国', '', '', '', 'wangwenguo', 'wwg', 1, '2026-10-29 11:27:13.104', '艺德利', '372421197605012217');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (94, '王羊', '', '', '', 'wangyang', 'wy', 1, '2026-10-29 11:27:13.104', '艺德利', '622628199907154671');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (95, '王义田', '', '', '', 'wangyitian', 'wyt', 1, '2026-10-29 11:27:13.105', '艺德利', '412725197404213033');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (96, '王永志', '', '', '', 'wangyongzhi', 'wyz', 1, '2026-10-29 11:27:13.106', '艺德利', '411628200104123017');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (97, '王云飞', '', '', '', 'wangyunfei', 'wyf', 1, '2026-10-29 11:27:13.107', '艺德利', '412725199807213010');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (98, '王在红', '', '', '', 'wangzaihong', 'wzh', 1, '2026-10-29 11:27:13.108', '艺德利', '622628198407154670');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (99, '王张想', '', '', '', 'wangzhangxiang', 'wzx', 1, '2026-10-29 11:27:13.109', '力楷', '622628198108124674');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (100, '王正山', '', '', '', 'wangzhengshan', 'wzs', 1, '2026-10-29 11:27:13.109', '力能', '372527198109212437');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (101, '武成刚', '', '', '', 'wuchenggang', 'wcg', 1, '2026-10-29 11:27:13.110', '力楷', '37142119811206221X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (102, '向竹林', '', '', '', 'xiangzhulin', 'xzl', 1, '2026-10-29 11:27:13.111', '艺德利', '513022198602252975');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (103, '肖刚', '', '', '', 'xiaogang', 'xg', 1, '2026-10-29 11:27:13.111', '瀚楷', '622628199602084676');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (104, '谢海涛', '', '', '', 'xiehaitao', 'xht', 1, '2026-10-29 11:27:13.112', '艺德利', '412725197601153439');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (105, '谢坤山', '', '', '', 'xiekunshan', 'xks', 1, '2026-10-29 11:27:13.113', '力楷', '412725197512113435');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (106, '谢新杰', '', '', '', 'xiexinjie', 'xxj', 1, '2026-10-29 11:27:13.114', '艺德利', '412725197903043411');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (107, '闫盼飞', '', '', '', 'yanpanfei', 'ypf', 1, '2026-10-29 11:27:13.116', '力楷', '142623199504054610');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (108, '杨彬', '', '', '', 'yangbin', 'yb', 1, '2026-10-29 11:27:13.117', '力楷', '132825198406160214');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (109, '杨利军', '', '', '', 'yanglijun', 'ylj', 1, '2026-10-29 11:27:13.118', '力楷', '622628199111234676');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (110, '杨想平', '', '', '', 'yangxiangping', 'yxp', 1, '2026-10-29 11:27:13.119', '力楷', '62262819870912467X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (111, '杨延会', '', '', '', 'yangyanhui', 'yyh', 1, '2026-10-29 11:27:13.120', '瀚楷', '142621199103101030');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (112, '杨永良', '', '', '', 'yangyongliang', 'yyl', 1, '2026-10-29 11:27:13.121', '艺德利', '622628198703254676');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (113, '尹明建', '', '', '', 'yinmingjian', 'ymj', 1, '2026-10-29 11:27:13.122', '瀚楷', '513022199407233538');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (114, '游森', '', '', '', 'yousen', 'ys', 1, '2026-10-29 11:27:13.123', '瀚楷', '372923199201040835');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (115, '俞小伟', '', '', '', 'yuxiaowei', 'yxw', 1, '2026-10-29 11:27:13.124', '艺德利', '622628199203044693');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (116, '张强', '', '', '', 'zhangqiang', 'zq', 1, '2026-10-29 11:27:13.124', '艺德利', '412725197711183831');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (117, '张天喂', '', '', '', 'zhangtianwei', 'ztw', 1, '2026-10-29 11:27:13.125', '力能', '622628199502104676');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (118, '张小春', '', '', '', 'zhangxiaochun', 'zxc', 1, '2026-10-29 11:27:13.126', '力能', '622628199211124677');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (119, '张永伟', '', '', '', 'zhangyongwei', 'zyw', 1, '2026-10-29 11:27:13.127', '瀚楷', '622628198904244677');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (120, '张玉', '', '', '', 'zhangyu', 'zy', 1, '2026-10-29 11:27:13.127', '力楷', '370832199005287055');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (121, '张泽江', '', '', '', 'zhangzejiang', 'zzj', 1, '2026-10-29 11:27:13.128', '力楷', '142623197105185710');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (122, '章大卫', '', '', '', 'zhangdawei', 'zdw', 1, '2026-10-29 11:27:13.129', '力能', '371422198708057717');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (123, '赵芳军', '', '', '', 'zhaofangjun', 'zfj', 1, '2026-10-29 11:27:13.130', '瀚楷', '622628199405134670');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (124, '赵吉海', '', '', '', 'zhaojihai', 'zjh', 1, '2026-10-29 11:27:13.131', '力能', '371526198902103714');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (125, '赵玉平', '', '', '', 'zhaoyuping', 'zyp', 1, '2026-10-29 11:27:13.131', '力能', '622628199408174678');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (126, '甄进良', '', '', '', 'zhenjinliang', 'zjl', 1, '2026-10-29 11:27:13.132', '瀚楷', '622628197911124679');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (127, '周奥迪', '', '', '', 'zhouaodi', 'zad', 1, '2026-10-29 11:27:13.132', '艺德利', '41162820050510305X');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (128, '朱月壮', '', '', '', 'zhuyuezhuang', 'zyz', 1, '2026-10-29 11:27:13.133', '瀚楷', '371727200104100516');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (129, '董树刚', '', '', '', 'dongshugang', 'dsg', 1, '2026-10-29 11:27:13.134', '力能', '372425197401208176');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (130, '刘春霞', '', '', '', 'liuchunxia', 'lcx', 1, '2026-10-29 11:27:13.135', '力能', '412327197702055425');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (131, '尹明超', '', '', '', 'yinmingchao', 'ymc', 1, '2026-10-29 11:27:13.136', '艺德利', '513022199203053535');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (132, '刘都玉', '', '', '', 'liudouyu', 'ldy', 1, '2026-10-29 11:27:13.136', '瀚楷', '622628199006124678');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (133, '左雕蛇', '', '', '', 'zuodiaoshe', 'zds', 1, '2026-10-29 11:27:13.137', '瀚楷', '622628198706044674');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (134, '王勇', '', '', '', 'wangyong', 'wy', 1, '2026-10-29 11:27:13.138', '瀚楷', '513022198310136972');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (135, '李军军', '', '', '', 'lijunjun', 'ljj', 1, '2026-10-29 11:27:13.139', '力能', '622628199003184675');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (136, '孙高明', '', '', '', 'sungaoming', 'sgm', 1, '2026-10-29 11:27:13.140', '瀚楷', '622628198911124673');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (137, '张克珍', '', '', '', 'zhangkezhen', 'zkz', 1, '2026-10-29 11:27:13.141', '瀚楷', '622628198512074672');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (138, '赵成', '', '', '', 'zhaocheng', 'zc', 1, '2026-10-29 11:27:13.142', '瀚楷', '132525197806250437');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (139, '邓高峰', '', '', '', 'denggaofeng', 'dgf', 1, '2026-10-29 11:27:13.142', '力能', '142623199011085718');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (140, '孟祥超', '', '', '', 'mengxiangchao', 'mxc', 1, '2026-10-29 11:27:13.143', '艺德利', '371425200301144615');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (141, '王立艇', '', '', '', 'wangliting', 'wlt', 1, '2026-10-29 11:27:13.144', '力能', '372425197904158174');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (142, '顾元波', '', '', '', 'guyuanbo', 'gyb', 1, '2026-10-29 11:27:13.144', '艺德利', '372425196312168614');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (143, '闫庆春', '', '', '', 'yanqingchun', 'yqc', 1, '2026-10-29 11:27:13.145', '力能', '110111197501068647');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (144, '孙博义', '', '', '', 'sunboyi', 'sby', 1, '2026-10-29 11:27:13.146', '艺德利', '410928198312274210');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (145, '孟伟蒙', '', '', '', 'mengweimeng', 'mwm', 1, '2026-10-29 11:27:13.147', '艺德利', '410928198809154337');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (146, '刘军', '', '', '', 'liujun', 'lj', 1, '2026-10-29 11:27:13.148', '艺德利', '510322198009247714');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (147, '王月娇', '', '', '', 'wangyuejiao', 'wyj', 1, '2026-10-29 11:27:13.150', '力楷', '130281198802042363');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (148, '刘进库', '', '', '', 'liujinku', 'ljk', 1, '2026-10-29 11:27:13.151', '瀚楷', '130726198704113116');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (149, '肖平安', '', '', '', 'xiaopingan', 'xpa', 1, '2026-10-29 11:27:13.152', '瀚楷', '500225199301254133');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (150, '李博', '', '', '', 'libo', 'lb', 1, '2026-10-29 11:27:13.153', '力能', '');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (151, '郭迎春', '', '', '', 'guoyingchun', 'gyc', 1, '2026-10-29 11:27:13.154', '力能', '');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (152, '马杰', '', '', '', 'majie', 'mj', 1, '2026-10-29 11:27:13.155', '瀚楷', '621226200508211472');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (153, '刘双迎', '', '', '', 'liushuangying', 'lsy', 1, '2026-10-29 11:27:13.156', '', '');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (154, '刘东辉', '', '', '', 'liudonghui', 'ldh', 1, '2026-10-29 11:27:13.158', '', '');
INSERT INTO `workers` (`id`, `name`, `phone`, `job_type`, `note`, `pinyin_full`, `pinyin_initial`, `active`, `created_at`, `company`, `id_card`) VALUES (155, '测试员甲', '', '', '', 'ceshiyuanjia', 'csyj', 1, '2026-10-30 16:13:56.669', '', '');
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
