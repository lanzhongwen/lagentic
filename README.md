# Coding-Oriented Multi-Agent Collaboration System. It's CLI style like Claude or OpenCode. Please refer to the design and implementation of Claude and OpenCode. each agent has independent context with different model.
# 代码开发场景｜多智能体协作标准角色划分
适用于：LLM 多智能体软件开发（AutoGen、LangGraph、OpenCode、Devin 类架构），分为**最小核心团队**、完整工程团队、辅助工具智能体三层，附带分工、职责、协作流程。

## 一、核心基础智能体（搭建最小可用编码多智能体系统必备）
### 1. 项目经理 / 协调智能体（Manager / Coordinator Agent）
**中枢调度，总负责人**
- 接收原始需求，拆解任务、分配子任务给各个智能体
- 管控执行流程、判断任务是否完成、冲突仲裁
- 汇总各智能体输出，把控整体进度，决定是否进入下一阶段
- 对外对接用户，收集澄清需求，传递反馈
> 作用：避免多个智能体并行乱执行，统一入口

### 2. 需求&架构智能体（Architect Agent）
负责顶层设计
- 梳理业务需求，识别边界、约束条件（性能、部署环境、技术栈）
- 设计系统架构、模块划分、接口定义、数据结构
- 输出架构图、模块职责、技术选型方案
- 制定编码规范、目录结构，约束开发智能体实现方案

### 3. 开发工程师智能体（Coder / Developer Agent）
核心编码角色
- 根据架构方案编写业务代码、函数、类、接口
- 实现功能逻辑，遵循约定的代码规范
- 处理基础异常、类型定义
- 接收调试反馈，迭代修改代码
> 大型系统可以拆分：前端开发Agent、后端开发Agent、算法开发Agent

### 4. 测试智能体（Tester Agent）
质量校验角色
- 编写单元测试、集成测试、接口测试用例
- 执行测试，捕获Bug、边界异常、逻辑错误
- 输出缺陷报告，反馈给开发Agent修复
- 验证修复后的代码是否通过用例

## 二、工程化进阶智能体（贴近真实软件开发流水线）
### 5. 代码审查智能体（Code Review Agent）
独立于测试，专注代码质量
- 检查代码规范、冗余代码、安全漏洞、性能隐患
- 识别潜在内存泄漏、SQL注入、硬编码、可读性问题
- 提出重构建议，不直接改代码，输出评审意见

### 6. 调试智能体（Debugger Agent）
专项排障
- 接收报错堆栈、日志、异常信息
- 定位根因，复现问题，给出修复思路
- 和开发Agent配合迭代修复复杂疑难Bug

### 7. 文档智能体（Docs Agent）
负责所有产出文档
- 生成接口文档（OpenAPI）、README、注释
- 撰写部署说明、使用示例、设计文档
- 同步更新文档，保证和代码同步

### 8. 构建&部署智能体（DevOps Agent）
工程运维侧
- 编写 Dockerfile、CI/CD脚本、启动脚本
- 处理依赖管理、环境配置、编译构建
- 执行本地启动、服务部署、端口/环境变量配置

## 三、工具工具型智能体（工具调用层，独立执行外部操作）
这类Agent不做逻辑推理，负责调用真实系统能力：
1. **文件操作智能体（File Agent）**：读写文件、创建目录、代码落地保存
2. **命令行智能体（Shell Agent）**：执行shell、pip安装、运行程序、执行测试命令
3. **检索智能体（Search Agent）**：查询技术文档、第三方API、开源方案
4. **Git智能体（Git Agent）**：提交代码、创建分支、查看diff

## 四、专项细分角色（复杂项目按需新增）
- **安全审计Agent**：扫描漏洞、依赖风险（SCA）
- **性能优化Agent**：分析耗时、优化算法、SQL调优
- **数据库Agent**：设计表结构、编写SQL、索引优化
- **产品Agent**：持续澄清模糊需求、校验功能是否符合预期

# 两种主流协作架构模式
## 模式1：串行流水线（最容易落地，LangGraph常用）
用户需求 → 协调Agent → 架构Agent → 开发Agent → CR评审Agent → 测试Agent → 文档Agent → DevOps部署
测试不通过 → 回传给开发Agent循环修复

## 模式2：并行辩论博弈架构（AutoGen经典）
Manager作为主席
- Developer：目标快速实现功能
- Tester：持续寻找漏洞、挑战实现方案
- CodeReview：持续提出优化意见
多方对话辩论，直到代码满足标准

# 【最简原型角色组合】快速上手推荐
如果你想快速搭建Demo，只保留5个：
1. Coordinator 协调者
2. Architect 架构师
3. Coder 开发
4. Tester 测试
5. File+Shell 工具智能体

# 示例交互流程
1. 用户：实现一个Go语言HTTP用户接口服务
2. Coordinator 接收需求，交给Architect设计模块与接口
3. Architect输出结构，下发任务给Coder
4. Coder编写代码，File Agent保存源码
5. Tester编写测试并运行，发现逻辑bug
6. 反馈至Coder修改，循环直至测试通过
7. 最终汇总输出完整工程

如果你需要，我可以直接提供：
1. AutoGen / LangGraph 各角色系统提示词模板
2. 多智能体消息流转流程图文本描述
3. 一套可直接运行的最小多智能体编码框架伪代码