FROM python:3.10-slim

WORKDIR /app

# 除非你的 requirements.txt 中有需要编译的 C 扩展包，否则不需要安装 build-essential
# 如果后续安装报错提示缺少编译器，再取消下面几行的注释并使用国内镜像源
# RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list && \
#     apt-get update && apt-get install -y --no-install-recommends build-essential && \
#     rm -rf /var/lib/apt/lists/*

# 复制依赖文件
COPY requirements.txt .

# 安装 Python 依赖
RUN pip install --no-cache-dir -r requirements.txt

# 复制源代码
COPY . .

# 启动命令
CMD ["python", "main.py"]