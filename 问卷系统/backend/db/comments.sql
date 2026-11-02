-- 一次性迁移：为 lineng 库所有表/字段补充 COMMENT（幂等，可重复执行）
-- 生成自本地库实际结构；新增字段请同步 internal/model/model.go 的 gorm comment 标签

SET NAMES utf8mb4;

ALTER TABLE `daily_reports`
  MODIFY COLUMN `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  MODIFY COLUMN `project_id` bigint unsigned NOT NULL COMMENT '所属项目 ID',
  MODIFY COLUMN `report_date` date NOT NULL COMMENT '出勤日期',
  MODIFY COLUMN `is_outsourced` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否外包：1 是，0 否（与是否存在外包明细等价）',
  MODIFY COLUMN `notes` varchar(500) NULL COMMENT '备注',
  MODIFY COLUMN `created_at` datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN `updated_at` datetime(3) NULL COMMENT '更新时间';

ALTER TABLE `daily_reports` COMMENT = '日报主表（一项目一天一单，保存即 upsert）';

ALTER TABLE `project_prices`
  MODIFY COLUMN `project_id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '项目 ID（主键，与 projects.id 对应）',
  MODIFY COLUMN `labor_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '人工单价，单位 元/人',
  MODIFY COLUMN `meal_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '餐费单价，单位 元/份',
  MODIFY COLUMN `water_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '水单价，单位 元/份',
  MODIFY COLUMN `fork3_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '叉车 3 吨单价，单位 元/台班',
  MODIFY COLUMN `fork5_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '叉车 5 吨单价，单位 元/台班',
  MODIFY COLUMN `fork7_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '叉车 7 吨单价，单位 元/台班',
  MODIFY COLUMN `crane_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '吊车单价，单位 元/台班',
  MODIFY COLUMN `rescue_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '救援车单价，单位 元/台班',
  MODIFY COLUMN `tax_rate` decimal(5,2) NOT NULL DEFAULT 6.00 COMMENT '税率百分比，6 表示 6%',
  MODIFY COLUMN `updated_at` datetime(3) NULL COMMENT '更新时间';

ALTER TABLE `project_prices` COMMENT = '项目默认单价（与 projects 一对一）';

ALTER TABLE `projects`
  MODIFY COLUMN `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  MODIFY COLUMN `no` varchar(32) NULL COMMENT '单据编号（LN-YYYYMMDD-XXX）',
  MODIFY COLUMN `unit` varchar(128) NULL COMMENT '用工单位',
  MODIFY COLUMN `activity_name` varchar(128) NULL COMMENT '活动名称',
  MODIFY COLUMN `location` varchar(128) NULL COMMENT '用工地点',
  MODIFY COLUMN `party_a_owner` varchar(64) NULL COMMENT '甲方负责人',
  MODIFY COLUMN `party_a_phone` varchar(32) NULL COMMENT '甲方联系电话',
  MODIFY COLUMN `start_date` date NULL COMMENT '用工开始日期',
  MODIFY COLUMN `end_date` date NULL COMMENT '用工结束日期',
  MODIFY COLUMN `created_at` datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN `updated_at` datetime(3) NULL COMMENT '更新时间';

ALTER TABLE `projects` COMMENT = '项目（出勤表表头信息）';

ALTER TABLE `report_equipment`
  MODIFY COLUMN `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  MODIFY COLUMN `report_id` bigint unsigned NOT NULL COMMENT '所属日报 ID',
  MODIFY COLUMN `equip_type` enum('FORK3','FORK5','FORK7','CRANE','RESCUE') NOT NULL COMMENT '设备类型：FORK3 叉车3吨 / FORK5 叉车5吨 / FORK7 叉车7吨 / CRANE 吊车 / RESCUE 救援车',
  MODIFY COLUMN `shift_type` enum('DAY','EVENING','DAWN') NOT NULL COMMENT '班次：DAY 白天 / EVENING 晚上 / DAWN 凌晨',
  MODIFY COLUMN `qty` bigint NOT NULL DEFAULT 0 COMMENT '台班数';

ALTER TABLE `report_equipment` COMMENT = '④ 日报-设备台班（5 种设备 × 3 个班次）';

ALTER TABLE `report_meals`
  MODIFY COLUMN `report_id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '所属日报 ID（主键）',
  MODIFY COLUMN `breakfast_qty` bigint NOT NULL DEFAULT 0 COMMENT '早点份数',
  MODIFY COLUMN `lunch_qty` bigint NOT NULL DEFAULT 0 COMMENT '中餐份数',
  MODIFY COLUMN `dinner_qty` bigint NOT NULL DEFAULT 0 COMMENT '晚餐份数',
  MODIFY COLUMN `supper_qty` bigint NOT NULL DEFAULT 0 COMMENT '夜宵份数',
  MODIFY COLUMN `water_qty` bigint NOT NULL DEFAULT 0 COMMENT '水份数';

ALTER TABLE `report_meals` COMMENT = '③ 日报-餐费与水（早点/中餐/晚餐/夜宵/水）';

ALTER TABLE `report_outsourcing`
  MODIFY COLUMN `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  MODIFY COLUMN `report_id` bigint unsigned NOT NULL COMMENT '所属日报 ID',
  MODIFY COLUMN `unit_name` varchar(128) NOT NULL COMMENT '外包单位名称',
  MODIFY COLUMN `day_count` bigint NOT NULL DEFAULT 0 COMMENT '白天人数',
  MODIFY COLUMN `evening_count` bigint NOT NULL DEFAULT 0 COMMENT '晚上人数',
  MODIFY COLUMN `dawn_count` bigint NOT NULL DEFAULT 0 COMMENT '凌晨人数',
  MODIFY COLUMN `ot_start` time NULL COMMENT '旧版通用加班开始时间（历史兼容，不再写入）',
  MODIFY COLUMN `ot_end` time NULL COMMENT '旧版通用加班结束时间（历史兼容，不再写入）',
  MODIFY COLUMN `overnight` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否跨夜：1 表示加班结束时间在次日',
  MODIFY COLUMN `sort_order` bigint unsigned NOT NULL DEFAULT 0 COMMENT '排列顺序',
  MODIFY COLUMN `day_ot_count` bigint NOT NULL DEFAULT 0 COMMENT '白班加班人数',
  MODIFY COLUMN `day_ot_start` time NULL COMMENT '白班加班开始时间，可空',
  MODIFY COLUMN `day_ot_end` time NULL COMMENT '白班加班结束时间，可空',
  MODIFY COLUMN `evening_ot_count` bigint NOT NULL DEFAULT 0 COMMENT '晚班加班人数',
  MODIFY COLUMN `evening_ot_start` time NULL COMMENT '晚班加班开始时间，可空',
  MODIFY COLUMN `evening_ot_end` time NULL COMMENT '晚班加班结束时间，可空',
  MODIFY COLUMN `dawn_ot_count` bigint NOT NULL DEFAULT 0 COMMENT '凌晨加班人数',
  MODIFY COLUMN `dawn_ot_start` time NULL COMMENT '凌晨加班开始时间，可空',
  MODIFY COLUMN `dawn_ot_end` time NULL COMMENT '凌晨加班结束时间，可空';

ALTER TABLE `report_outsourcing` COMMENT = '⑤ 日报-外包用工（一个外包单位一行）';

ALTER TABLE `report_shifts`
  MODIFY COLUMN `report_id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '所属日报 ID（主键）',
  MODIFY COLUMN `day_count` bigint NOT NULL DEFAULT 0 COMMENT '白天人数',
  MODIFY COLUMN `evening_count` bigint NOT NULL DEFAULT 0 COMMENT '晚上人数',
  MODIFY COLUMN `dawn_count` bigint NOT NULL DEFAULT 0 COMMENT '凌晨人数';

ALTER TABLE `report_shifts` COMMENT = '② 日报-用工人数（白天/晚上/凌晨）';

ALTER TABLE `report_workers`
  MODIFY COLUMN `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  MODIFY COLUMN `report_id` bigint unsigned NOT NULL COMMENT '所属日报 ID',
  MODIFY COLUMN `worker_id` bigint unsigned NOT NULL COMMENT '人员 ID（关联 workers.id）',
  MODIFY COLUMN `worker_name` varchar(64) NOT NULL COMMENT '下单时的姓名快照，人员改名不影响历史单据',
  MODIFY COLUMN `start_time` time NOT NULL COMMENT '上班时间 HH:MM',
  MODIFY COLUMN `end_time` time NOT NULL COMMENT '下班时间 HH:MM',
  MODIFY COLUMN `overnight` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否跨夜：1 表示下班时间在次日',
  MODIFY COLUMN `sort_order` bigint unsigned NOT NULL DEFAULT 0 COMMENT '名单内排序序号（决定考勤表行序）';

ALTER TABLE `report_workers` COMMENT = '① 日报-用工名单（先时间后多人，一行一工）';

ALTER TABLE `users`
  MODIFY COLUMN `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  MODIFY COLUMN `username` varchar(64) NULL COMMENT '登录名',
  MODIFY COLUMN `password` longtext NULL COMMENT '登录密码（bcrypt 哈希）',
  MODIFY COLUMN `created_at` datetime(3) NULL COMMENT '创建时间';

ALTER TABLE `users` COMMENT = '管理员账号';

ALTER TABLE `workers`
  MODIFY COLUMN `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  MODIFY COLUMN `name` varchar(64) NULL COMMENT '姓名（唯一）',
  MODIFY COLUMN `phone` varchar(32) NULL COMMENT '联系电话',
  MODIFY COLUMN `job_type` varchar(64) NULL COMMENT '工种',
  MODIFY COLUMN `note` varchar(255) NULL COMMENT '备注',
  MODIFY COLUMN `pinyin_full` varchar(255) NULL COMMENT '姓名全拼，小写无空格，用于搜索',
  MODIFY COLUMN `pinyin_initial` varchar(64) NULL COMMENT '姓名拼音首字母，用于搜索',
  MODIFY COLUMN `active` tinyint(1) NULL DEFAULT 1 COMMENT '是否在职：1 在职，0 停用',
  MODIFY COLUMN `created_at` datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN `company` varchar(64) NULL COMMENT '所属公司（员工表导入）',
  MODIFY COLUMN `id_card` varchar(32) NULL COMMENT '身份证号（员工表导入）';

ALTER TABLE `workers` COMMENT = '人员档案';
