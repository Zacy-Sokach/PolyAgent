#!/bin/bash

# 构建 Rust 库
cd "$(dirname "$0")/rust_markdown"

# 创建 release 构建
cargo build --release

# 复制静态库到上级目录
cp target/release/librust_markdown.a ../