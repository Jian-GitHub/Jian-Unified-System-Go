# JQuantum

**Quantum circuit parameter compression and generation toolkit / 量子电路参数压缩与生成工具包**

`JQuantum` is the Python component of **Jian Unified System**, designed for quantum circuit parameter compression and generation.  
It is mainly used for quantum computing simulation and algorithm research.  
This package depends on Qiskit and NumPy, and provides a unified `Client` class as the main entry point.

`JQuantum` 是 **Jian Unified System** 的 Python 组件，用于量子电路参数压缩和生成，主要面向量子计算模拟和算法研究。  
它依赖 Qiskit 和 NumPy，提供统一的 `Client` 类作为主要入口。

---

## 📦 Installation / 安装

Install from PyPI:

从 PyPI 安装：

```bash
pip install jquantum
```
Install development version from GitHub:
从 GitHub 安装开发版本：
```bash
pip install git+https://github.com/Jian-GitHub/Jian-Unified-System-Go.git#subdirectory=jquantum/jquantum-py
```

## ⚡ Quick Start / 快速使用
```python
from jquantum import Client

# Create a Client instance
# 创建 Client 实例
client = Client()

# Use Client methods
# 使用 Client 提供的方法
client.some_method(...)
```
All features are wrapped in the Client class. Users do not need to directly import internal modules.

所有功能都通过 Client 类封装，用户无需直接操作内部模块。

## 🔗 Dependencies / 依赖
- Python >= 3.10
- Qiskit
- NumPy

You can use pip install -r requirements.txt if you maintain a requirements file.

如果你维护了 requirements.txt，可以使用 pip install -r requirements.txt 安装依赖。

## 📁 File Structure / 文件结构
```
Jian-Unified-System/jquantum/jquantum-py
├── jquantum/            # package directory / 包目录
│   ├── __init__.py
│   ├── client.py        # main interface / 对外接口
│   ├── compress.py
│   ├── con.py
│   ├── params_compressor.py
│   ├── params_generator.py
│   └── pattern_code.py
├── README.md            # this file / 当前文件
├── LICENSE
└── pyproject.toml
```
Internal modules like compress, params_compressor are not meant to be directly imported. Users should use Client only.

内部模块如 compress、params_compressor 等无需直接导入，用户只使用 Client 即可。

## 📜 License / 许可
MIT License © 2025 Jian Qi

## 🐛 Issues & Feedback / 问题反馈
Submit issues at the repository / 问题反馈请到仓库提交：

https://github.com/Jian-GitHub/Jian-Unified-System-Go/issues