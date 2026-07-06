-- 初始化数据库 Schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS bloggers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(50),
    avatar VARCHAR(500),
    bio TEXT,
    email VARCHAR(100),
    blog_title VARCHAR(100),
    blog_description TEXT,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 表注释
COMMENT ON TABLE bloggers IS '博主信息表（单博主系统，仅一条记录）';

-- 字段注释
COMMENT ON COLUMN bloggers.id IS '博主唯一标识';
COMMENT ON COLUMN bloggers.username IS '登录用户名';
COMMENT ON COLUMN bloggers.password_hash IS 'bcrypt哈希密码';
COMMENT ON COLUMN bloggers.nickname IS '昵称';
COMMENT ON COLUMN bloggers.avatar IS '头像URL';
COMMENT ON COLUMN bloggers.bio IS '个人简介';
COMMENT ON COLUMN bloggers.email IS '邮箱（用于通知）';
COMMENT ON COLUMN bloggers.blog_title IS '博客标题';
COMMENT ON COLUMN bloggers.blog_description IS '博客描述';
COMMENT ON COLUMN bloggers.last_login_at IS '最后登录时间';
COMMENT ON COLUMN bloggers.created_at IS '创建时间';
COMMENT ON COLUMN bloggers.updated_at IS '更新时间';