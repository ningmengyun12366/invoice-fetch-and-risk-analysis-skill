package report

import (
	"regexp"
	"strconv"
	"strings"
)

var reportNumberToken = regexp.MustCompile(`[+-]?\d[\d,]*(?:\.\d+)?`)

// formatReportNarrativeAmounts adds thousands separators to monetary values
// written by the Agent. It deliberately requires amount context so years,
// percentages, tax IDs and classification codes are left unchanged.
func formatReportNarrativeAmounts(text string) string {
	if strings.TrimSpace(text) == "" {
		return text
	}
	matches := reportNumberToken.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text
	}
	var builder strings.Builder
	last := 0
	for _, match := range matches {
		start, end := match[0], match[1]
		builder.WriteString(text[last:start])
		raw := text[start:end]
		if format, ok := reportAmountToken(raw, text, start, end); ok {
			builder.WriteString(format)
		} else {
			builder.WriteString(raw)
		}
		last = end
	}
	builder.WriteString(text[last:])
	return builder.String()
}

func reportAmountToken(raw, text string, start, end int) (string, bool) {
	unsigned := strings.TrimLeft(raw, "+-")
	digits := strings.ReplaceAll(unsigned, ",", "")
	parts := strings.SplitN(digits, ".", 2)
	if len(parts[0]) < 4 || len(digits) == 0 {
		return "", false
	}
	// A number already grouped is still accepted, but its output remains stable.
	if end < len(text) && text[end] == '%' {
		return "", false
	}
	if end < len(text) && (strings.HasPrefix(text[end:], "年") || strings.HasPrefix(text[end:], "月") || strings.HasPrefix(text[end:], "日") || strings.HasPrefix(text[end:], "条") || strings.HasPrefix(text[end:], "笔") || strings.HasPrefix(text[end:], "家") || strings.HasPrefix(text[end:], "张") || strings.HasPrefix(text[end:], "项") || strings.HasPrefix(text[end:], "次") || strings.HasPrefix(text[end:], "户") || strings.HasPrefix(text[end:], "名")) {
		return "", false
	}
	if len(parts[0]) == 4 && end+3 <= len(text) && (text[end] == '-' || text[end] == '/') && text[end+1] >= '0' && text[end+1] <= '9' && text[end+2] >= '0' && text[end+2] <= '9' {
		return "", false
	}
	left := start - 24
	if left < 0 {
		left = 0
	}
	right := end + 24
	if right > len(text) {
		right = len(text)
	}
	context := text[left:right]
	if !containsAny(context, "金额", "合计", "元", "人民币", "税额", "税款", "收入", "销售额", "营业额", "成本", "费用", "利润", "发票", "采购", "购进", "销项", "进项", "差额", "净额", "余额", "实缴", "申报", "应补", "应纳", "现金流", "资本", "资产", "负债", "回款", "预付款", "欠税", "罚款", "处罚") {
		return "", false
	}
	negative := strings.HasPrefix(raw, "-")
	positive := strings.HasPrefix(raw, "+")
	integer := parts[0]
	for index := len(integer) - 3; index > 0; index -= 3 {
		integer = integer[:index] + "," + integer[index:]
	}
	formatted := integer
	if len(parts) == 2 {
		formatted += "." + parts[1]
	}
	if negative {
		formatted = "-" + formatted
	} else if positive {
		formatted = "+" + formatted
	}
	return formatted, true
}

// formatTableAmountCells applies the table's amount headers to bare numeric
// cells, covering legacy blocks that predate moneyText's grouped formatting.
func formatTableAmountCells(block Block) Block {
	amountColumn := make([]bool, len(block.Headers))
	for i, header := range block.Headers {
		amountColumn[i] = containsAny(header, "金额", "税额", "收入", "成本", "费用", "利润", "余额", "资本", "差额", "净额", "实缴", "申报")
	}
	for rowIndex := range block.Rows {
		for column := range block.Rows[rowIndex] {
			if column >= len(amountColumn) || !amountColumn[column] {
				continue
			}
			value := strings.TrimSpace(block.Rows[rowIndex][column])
			if number, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", ""), 64); err == nil {
				block.Rows[rowIndex][column] = groupedAmountText(number)
			}
		}
	}
	return block
}

func groupedAmountText(value float64) string {
	if value > -0.005 && value < 0.005 {
		value = 0
	}
	negative := value < 0
	if negative {
		value = -value
	}
	raw := strconv.FormatFloat(value, 'f', 2, 64)
	parts := strings.SplitN(raw, ".", 2)
	integer := parts[0]
	for index := len(integer) - 3; index > 0; index -= 3 {
		integer = integer[:index] + "," + integer[index:]
	}
	result := integer + "." + parts[1]
	if negative && result != "0.00" {
		return "-" + result
	}
	return result
}

// formatFinalModelAmounts applies display-only grouping immediately before the
// two report renderers run. Frozen calculations, AI evidence IDs and source
// snapshots retain their original numeric representations.
func formatFinalModelAmounts(model Model) Model {
	for sectionIndex := range model.Sections {
		section := &model.Sections[sectionIndex]
		for blockIndex := range section.Blocks {
			block := &section.Blocks[blockIndex]
			*block = formatTableAmountCells(*block)
			for rowIndex := range block.Rows {
				for columnIndex := range block.Rows[rowIndex] {
					header := ""
					if columnIndex < len(block.Headers) {
						header = block.Headers[columnIndex]
					}
					block.Rows[rowIndex][columnIndex] = formatReportTableCellAmount(header, block.Rows[rowIndex][columnIndex])
				}
			}
			for paragraphIndex := range block.Paragraphs {
				block.Paragraphs[paragraphIndex] = formatReportNarrativeAmounts(block.Paragraphs[paragraphIndex])
			}
			if block.Chart != nil {
				block.Chart.Summary = formatReportNarrativeAmounts(block.Chart.Summary)
				block.Chart.CenterValue = formatReportNarrativeAmounts(block.Chart.CenterValue)
			}
			formatFindingAmounts(block.Finding)
		}
	}
	for index := range model.Findings {
		formatFindingAmounts(&model.Findings[index])
	}
	for index := range model.Evaluations {
		model.Evaluations[index].Evidence = formatReportNarrativeAmounts(model.Evaluations[index].Evidence)
		formatFindingAmounts(model.Evaluations[index].Finding)
	}
	return model
}

func formatReportTableCellAmount(header, value string) string {
	if strings.TrimSpace(header) == "" {
		return formatReportNarrativeAmounts(value)
	}
	prefix := header + "："
	formatted := formatReportNarrativeAmounts(prefix + value)
	return strings.TrimPrefix(formatted, prefix)
}

func formatFindingAmounts(finding *Finding) {
	if finding == nil {
		return
	}
	finding.MajorImpactEvidence = formatReportNarrativeAmounts(finding.MajorImpactEvidence)
	finding.Evidence = formatReportNarrativeAmounts(finding.Evidence)
	finding.Recommendation = formatReportNarrativeAmounts(finding.Recommendation)
}
