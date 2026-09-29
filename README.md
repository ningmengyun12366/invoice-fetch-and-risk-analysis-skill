# 柠檬云发票获取及风险分析

本公开源码包只包含 `invoice-fetch-and-risk-analysis` 业务 Skill。公共的 Go 实现、Node.js 打包脚本和依赖代码保留在 `go-script/` 中，便于查看、构建和维护。

## 描述

**名字：** 发票获取_发票附件归集_发票风险分析_发票分析_发票整理_发票扫描_税务风险分析_发票扫描_进销项统计_会计票据管理_财税发票分析_柠檬云发票

**描述：**

柠檬云AI财税开放平台：财税发票归集及税务风险分析Skill，专为财务、会计、财税人员打造，一键获取企业指定期间发票及电子附件，自动完成票据整理、进销项数据汇总，结合纳税人身份、行业特性智能扫描税务风险，分角色输出可落地的财务分析结论与涉税整改方案。
✅发票批量自动归集存档：输入企业名称与时间区间，自动拉取对应发票清单，同步下载数电发票电子附件，建立发票与附件关联索引，帮助会计、财务完成全量票据归集、电子档案留存；
✅多维度财务数据自动统计：自动汇总进销项金额、发票类型、月度营收、上下游合作方等六大维度财税数据，一键输出标准化票据统计结果，减轻财务做账统计工作量；
✅行业化财税风险智能诊断：结合企业所属行业、纳税人身份，智能识别税负异常、进项抵扣不匹配等涉税风险，分岗位角色输出可落地风险分析报告与财税整改建议。

详细的 Skill 规则、使用前准备和安装更新流程见 [`invoice-fetch-and-risk-analysis/README.md`](./invoice-fetch-and-risk-analysis/README.md)、[`invoice-fetch-and-risk-analysis/SKILL.md`](./invoice-fetch-and-risk-analysis/SKILL.md) 和 [`invoice-fetch-and-risk-analysis/invoice-fetch-and-risk-analysis-install.md`](./invoice-fetch-and-risk-analysis/invoice-fetch-and-risk-analysis-install.md)。

## 目录结构

- `invoice-fetch-and-risk-analysis/`：当前 Skill 的说明、安装文档、引用资料和资源。
- `go-script/`：Go 命令行实现、公共库和生产打包脚本。

## 生产构建

构建要求 Node.js 20+ 和 Go 1.24+。请在仓库根目录执行：

```powershell
npm ci
node go-script/build.js --env prod invoice-fetch-and-risk-analysis
```

生产发布物生成在 `dist/Skills/invoice-fetch-and-risk-analysis/`。运行时 ZIP 包不包含 `bin/**` 二进制文件，安装时按安装文档下载当前平台对应的二进制。

## 配置与数据安全

- API 凭据由使用者在本地工作区配置，不应写入源码、文档或日志。
- 生产接口通过 `go-script/config/prod.json` 注入发布包。
- 安全问题请按照 [SECURITY.md](SECURITY.md) 私下报告。

## 许可证

本项目自有代码、文档和资源文件采用 [MIT License](LICENSE)；第三方组件与字体许可见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。
