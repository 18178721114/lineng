-- =============================================================
-- 力能 出勤/结算系统 数据库结构 V2
-- 规范化设计：无 JSON、无 TEXT，一个业务模块一张表
-- MySQL 8.0 / InnoDB / utf8mb4_unicode_ci
--
-- 表清单：
--   users               账号（单管理员）
--   projects            项目（对应纸质《出勤表》表头）
--   project_prices      项目默认单价（1:1，由 projects.prices TEXT 拆出）
--   workers             人员档案（拼音字段用于搜索）
--   daily_reports       日报主表（只留标量与外包标记）
--   report_workers      ① 用工名单：先时间后多人
--   report_shifts       ② 用工人数：白天/晚上/凌晨
--   report_meals        ③ 餐费：早点/中餐/晚餐/夜宵/水
--   report_equipment    ④ 设备台班：叉车3/5/7吨、吊车、救援车 × 白/晚/凌晨
--   report_outsourcing  ⑤ 外包：一单位一行，可多套
-- =============================================================

SET NAMES utf8mb4;

CREATE DATABASE IF NOT EXISTS lineng
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE lineng;

SET FOREIGN_KEY_CHECKS = 0;

-- -------------------------------------------------------------
-- 1. 账号
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username   VARCHAR(64)     NOT NULL                COMMENT '登录名',
  password   VARCHAR(255)    NOT NULL                COMMENT 'bcrypt 哈希',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员账号';

-- -------------------------------------------------------------
-- 2. 项目（去掉 prices TEXT，日期改用 DATE）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS projects (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  no            VARCHAR(32)     NOT NULL              COMMENT '单据编号 LN-YYYYMMDD-XXX',
  unit          VARCHAR(128)    NOT NULL DEFAULT ''   COMMENT '用工单位',
  activity_name VARCHAR(128)    NOT NULL DEFAULT ''   COMMENT '活动名称',
  location      VARCHAR(128)    NOT NULL DEFAULT ''   COMMENT '地点',
  party_a_owner VARCHAR(64)     NOT NULL DEFAULT ''   COMMENT '甲方负责人',
  party_a_phone VARCHAR(32)     NOT NULL DEFAULT ''   COMMENT '甲方电话',
  start_date    DATE            NOT NULL              COMMENT '用工开始',
  end_date      DATE            NOT NULL              COMMENT '用工结束',
  created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_no (no),
  KEY idx_period (start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目（出勤表表头信息）';

-- -------------------------------------------------------------
-- 2.1 项目默认单价（1:1）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS project_prices (
  project_id   BIGINT UNSIGNED NOT NULL,
  labor_price  DECIMAL(10,2)   NOT NULL DEFAULT 0     COMMENT '人工 元/人',
  meal_price   DECIMAL(10,2)   NOT NULL DEFAULT 0     COMMENT '餐费 元/份',
  water_price  DECIMAL(10,2)   NOT NULL DEFAULT 0     COMMENT '水 元/份',
  fork3_price  DECIMAL(10,2)   NOT NULL DEFAULT 0     COMMENT '叉车3吨 元/台班',
  fork5_price  DECIMAL(10,2)   NOT NULL DEFAULT 0     COMMENT '叉车5吨 元/台班',
  fork7_price  DECIMAL(10,2)   NOT NULL DEFAULT 0     COMMENT '叉车7吨 元/台班（原1吨）',
  crane_price  DECIMAL(10,2)   NOT NULL DEFAULT 0     COMMENT '吊车 元/台班（新增）',
  rescue_price DECIMAL(10,2)   NOT NULL DEFAULT 0     COMMENT '救援车 元/台班（新增）',
  tax_rate     DECIMAL(5,2)    NOT NULL DEFAULT 6.00   COMMENT '税率百分比，6 表示 6%',
  updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (project_id),
  CONSTRAINT fk_prices_project FOREIGN KEY (project_id)
    REFERENCES projects (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目默认单价';

-- -------------------------------------------------------------
-- 3. 人员档案
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS workers (
  id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name           VARCHAR(64)     NOT NULL              COMMENT '姓名',
  phone          VARCHAR(32)     NOT NULL DEFAULT ''   COMMENT '电话',
  job_type       VARCHAR(64)     NOT NULL DEFAULT ''   COMMENT '工种',
  note           VARCHAR(255)    NOT NULL DEFAULT ''   COMMENT '备注',
  pinyin_full    VARCHAR(255)    NOT NULL DEFAULT ''   COMMENT '全拼，小写无空格',
  pinyin_initial VARCHAR(64)     NOT NULL DEFAULT ''   COMMENT '拼音首字母',
  active         TINYINT(1)      NOT NULL DEFAULT 1    COMMENT '1 在职，0 停用',
  created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_name (name),
  KEY idx_pinyin_full (pinyin_full),
  KEY idx_pinyin_initial (pinyin_initial),
  KEY idx_active (active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='人员档案';

-- -------------------------------------------------------------
-- 4. 日报主表（一项目一天一单，保存即 upsert）
--    原 workers_json / shifts_json / meals_json / equipment_json
--    / outsourcing_json 五个 TEXT 列全部拆到下列子表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS daily_reports (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  project_id    BIGINT UNSIGNED NOT NULL,
  report_date   DATE            NOT NULL              COMMENT '出勤日期',
  is_outsourced TINYINT(1)      NOT NULL DEFAULT 0    COMMENT '是否外包（与 ⑤ 有行等价，便于筛选）',
  notes         VARCHAR(500)    NOT NULL DEFAULT ''   COMMENT '备注',
  created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_project_date (project_id, report_date),
  KEY idx_date (report_date),
  CONSTRAINT fk_report_project FOREIGN KEY (project_id)
    REFERENCES projects (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日报主表';

-- -------------------------------------------------------------
-- ① 用工名单：先填工作时间（开始~结束），再选人名（可多选）
--    一次勾选多人 = 多行共享同一 start_time/end_time
--    名单行数 = 工数，可直接 GROUP BY 出考勤表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS report_workers (
  id          BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
  report_id   BIGINT UNSIGNED   NOT NULL,
  worker_id   BIGINT UNSIGNED   NOT NULL              COMMENT '人员',
  worker_name VARCHAR(64)       NOT NULL DEFAULT ''   COMMENT '下单时姓名快照，改名不影响历史单据',
  start_time  TIME              NOT NULL              COMMENT '上班时间',
  end_time    TIME              NOT NULL              COMMENT '下班时间',
  overnight   TINYINT(1)        NOT NULL DEFAULT 0    COMMENT '1 = 下班时间在次日',
  sort_order  SMALLINT UNSIGNED NOT NULL DEFAULT 0    COMMENT '名单内顺序（考勤表行序）',
  PRIMARY KEY (id),
  KEY idx_report (report_id),
  KEY idx_worker (worker_id),
  CONSTRAINT fk_rw_report FOREIGN KEY (report_id)
    REFERENCES daily_reports (id) ON DELETE CASCADE,
  CONSTRAINT fk_rw_worker FOREIGN KEY (worker_id)
    REFERENCES workers (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='① 用工名单';

-- -------------------------------------------------------------
-- ② 用工人数：白天/晚上/凌晨（按名单自动带出，可手改）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS report_shifts (
  report_id     BIGINT UNSIGNED   NOT NULL,
  day_count     SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '白天人数',
  evening_count SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '晚上人数',
  dawn_count    SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '凌晨人数',
  PRIMARY KEY (report_id),
  CONSTRAINT fk_rs_report FOREIGN KEY (report_id)
    REFERENCES daily_reports (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='② 用工人数';

-- -------------------------------------------------------------
-- ③ 餐费：早点 / 中餐 / 晚餐 / 夜宵 + 水（新增）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS report_meals (
  report_id     BIGINT UNSIGNED   NOT NULL,
  breakfast_qty SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '早点份数',
  lunch_qty     SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '中餐份数',
  dinner_qty    SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '晚餐份数',
  supper_qty    SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '夜宵份数',
  water_qty     SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '水份数（新增）',
  PRIMARY KEY (report_id),
  CONSTRAINT fk_rm_report FOREIGN KEY (report_id)
    REFERENCES daily_reports (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='③ 餐费';

-- -------------------------------------------------------------
-- ④ 设备台班：叉车3吨/5吨/7吨、吊车、救援车 × 白班/晚班/凌晨
--    每单固定 15 行，0 值行用于 UPSERT 与报表取数
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS report_equipment (
  id         BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
  report_id  BIGINT UNSIGNED   NOT NULL,
  equip_type ENUM('FORK3','FORK5','FORK7','CRANE','RESCUE') NOT NULL
                                                    COMMENT '叉车3吨/5吨/7吨、吊车、救援车',
  shift_type ENUM('DAY','EVENING','DAWN')          NOT NULL
                                                    COMMENT '白班/晚班/凌晨',
  qty        SMALLINT UNSIGNED NOT NULL DEFAULT 0   COMMENT '台班数',
  PRIMARY KEY (id),
  UNIQUE KEY uk_report_equip_shift (report_id, equip_type, shift_type),
  CONSTRAINT fk_re_report FOREIGN KEY (report_id)
    REFERENCES daily_reports (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='④ 设备台班';

-- -------------------------------------------------------------
-- ⑤ 外包：是否外包 → 外包单位（名称 + 白/晚/凌 的人数 / 加班人数 / 加班起~止）
--    一个外包单位一行，可加多套；三个班次各自记录加班
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS report_outsourcing (
  id             BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
  report_id      BIGINT UNSIGNED   NOT NULL,
  unit_name      VARCHAR(128)      NOT NULL            COMMENT '外包单位名称',
  day_count      SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '白天人数',
  evening_count  SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '晚上人数',
  dawn_count     SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '凌晨人数',
  day_ot_count   SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '白班加班人数',
  day_ot_start   TIME              NULL                COMMENT '白班加班开始，可空',
  day_ot_end     TIME              NULL                COMMENT '白班加班结束，可空',
  eve_ot_count   SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '晚班加班人数',
  eve_ot_start   TIME              NULL                COMMENT '晚班加班开始，可空',
  eve_ot_end     TIME              NULL                COMMENT '晚班加班结束，可空',
  dawn_ot_count  SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '凌晨加班人数',
  dawn_ot_start  TIME              NULL                COMMENT '凌晨加班开始，可空',
  dawn_ot_end    TIME              NULL                COMMENT '凌晨加班结束，可空',
  ot_start       TIME              NULL                COMMENT '旧版通用加班开始（历史兼容，不再写入）',
  ot_end         TIME              NULL                COMMENT '旧版通用加班结束（历史兼容，不再写入）',
  overnight      TINYINT(1)        NOT NULL DEFAULT 0 COMMENT '1 = 加班结束在次日',
  sort_order     SMALLINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排列顺序',
  PRIMARY KEY (id),
  KEY idx_report (report_id),
  CONSTRAINT fk_ro_report FOREIGN KEY (report_id)
    REFERENCES daily_reports (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='⑤ 外包';

-- -------------------------------------------------------------
-- ⑥ 模块暂存草稿：H5「暂存本模块」入库（同项目同日期同模块一条，
--    有则更新、无则新增；提交日报后整组删除）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS report_module_drafts (
  id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  project_id  BIGINT UNSIGNED NOT NULL              COMMENT '所属项目',
  report_date DATE            NOT NULL              COMMENT '出勤日期',
  module_key  VARCHAR(32)     NOT NULL              COMMENT '模块键（names/labor/meal/fork/out/note）',
  content     JSON            NOT NULL              COMMENT '模块暂存内容（JSON 对象，字段名→值）',
  created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_draft (project_id, report_date, module_key),
  CONSTRAINT fk_rmd_project FOREIGN KEY (project_id)
    REFERENCES projects (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='⑥ 模块暂存草稿';

SET FOREIGN_KEY_CHECKS = 1;

-- =============================================================
-- 常用取数示例
-- =============================================================
-- 考勤表：某人某天出工次数（名单行数 = 工数）
-- SELECT wk.name, wk.pinyin_full, DAYOFMONTH(r.report_date) AS d, COUNT(*) AS gong
--   FROM report_workers w
--   JOIN daily_reports r ON r.id = w.report_id
--   JOIN workers wk      ON wk.id = w.worker_id
--  WHERE r.project_id = ? AND r.report_date BETWEEN ? AND ?
--  GROUP BY wk.name, wk.pinyin_full, r.report_date;
--
-- 结算单区间汇总：人工 / 餐水 / 设备
-- SELECT SUM(s.day_count + s.evening_count + s.dawn_count)                AS labor_cnt,
--        SUM(m.breakfast_qty + m.lunch_qty + m.dinner_qty + m.supper_qty) AS meal_cnt,
--        SUM(m.water_qty)                                                 AS water_cnt,
--        SUM(CASE WHEN e.equip_type='FORK7' THEN e.qty ELSE 0 END)         AS fork7_taiban
--   FROM daily_reports r
--   LEFT JOIN report_shifts    s ON s.report_id = r.id
--   LEFT JOIN report_meals     m ON m.report_id = r.id
--   LEFT JOIN report_equipment e ON e.report_id = r.id
--  WHERE r.project_id = ? AND r.report_date BETWEEN ? AND ?;
--
-- 人员工时排行
-- SELECT wk.name, COUNT(*) AS days,
--        SUM(TIME_TO_SEC(TIMEDIFF(w.end_time, w.start_time))
--            + IF(w.overnight, 86400, 0)) / 3600 AS hours
--   FROM report_workers w
--   JOIN workers wk ON wk.id = w.worker_id
--  GROUP BY wk.name
--  ORDER BY hours DESC;
