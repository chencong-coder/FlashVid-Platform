# 多阶段构建：构建 + 运行
FROM node:20-alpine AS builder

WORKDIR /build

# 复制 package.json 和 package-lock.json
COPY package*.json ./

# 安装依赖
RUN npm ci

# 复制源代码
COPY . .

# 生产环境构建
RUN npm run build

# ============================================
# Nginx 运行镜像
# ============================================
FROM nginx:alpine

WORKDIR /usr/share/nginx/html

# 从 builder 复制构建产物
COPY --from=builder /build/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
