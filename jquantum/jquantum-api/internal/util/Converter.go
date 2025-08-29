package util

import (
	"encoding/json"
	"fmt"
	"jian-unified-system/jus-core/types/mq/jquantum"
	"log"
	"os"
)

type Converter struct {
	QbitIndex int
	Params    []interface{}
}

var supportGates = []string{
	"h", "x", "y", "z", "s", "sdg", "t", "tdg", "sx", "sxdg", "rx", "ry", "rz", "p", "u", "cx",
	"cz", "cy", "ch", "swap", "iswap", "cu3", "crx", "cry", "crz", "cu1", "rxx", "ryy",
	"rzz", "rzx", "ccx", "cswap", "cswap", "mcp", "mcx",
}

// 安全的类型断言辅助函数
func safeGetString(element interface{}) string {
	if str, ok := element.(string); ok {
		return str
	}
	return ""
}

func safeGetMap(element interface{}) map[string]interface{} {
	if m, ok := element.(map[string]interface{}); ok {
		return m
	}
	return nil
}

// 解析元素（保留嵌套结构）
func (c *Converter) resolveElementStructured(element jquantum.Element, patterns map[string]jquantum.PatternContent) interface{} {
	if str := safeGetString(element); str != "" {
		return str
	}

	if m := safeGetMap(element); m != nil {
		ref, refExists := m["ref"].(string)
		if !refExists {
			return "invalid_ref"
		}

		count := 1
		if cnt, ok := m["count"].(float64); ok {
			count = int(cnt)
		}

		if pattern, exists := patterns[ref]; exists {
			result := make([][]interface{}, count)
			for i := 0; i < count; i++ {
				inner := make([]interface{}, len(pattern.Content))
				for j, item := range pattern.Content {
					inner[j] = c.resolveElementStructured(item, patterns)
				}
				result[i] = inner
			}
			return result
		} else {
			// 普通引用（不是 pattern），直接展开为列表
			result := make([]string, count)
			for i := 0; i < count; i++ {
				result[i] = ref
			}
			return result
		}
	}

	return fmt.Sprintf("unknown_element_%v", element)
}

// 生成模式代码 - 增强健壮性
func (c *Converter) generatePatternsCode(patterns map[string]jquantum.PatternContent) string {
	if patterns == nil || len(patterns) == 0 {
		return "// 无重复模式\n"
	}

	patternsHead := "// 重复模式函数声明\n"
	patternsContent := "\n// 重复模式函数实现\n"
	paramIndex := 0

	for name, pattern := range patterns {
		patternsHead += fmt.Sprintf("void %s(Qureg& qureg, const Params& pattern_params);\n", name)
		patternsContent += fmt.Sprintf("void %s(Qureg& qureg, const Params& pattern_params) {\n", name)
		patternsContent += "\tParams params;\n"

		// 检查Content是否为空
		if pattern.Content == nil || len(pattern.Content) == 0 {
			patternsContent += "\t// 空模式内容\n"
		} else {
			for idx, item := range pattern.Content {
				switch {
				case safeGetString(item) != "":
					gateName := safeGetString(item)
					patternsContent += "\n\tparams = {};\n"
					patternsContent += fmt.Sprintf("\tparams.push_back(pattern_params[%d]);\n", paramIndex)
					patternsContent += NameToCode(gateName, 1, "0")
					paramIndex++
				case safeGetMap(item) != nil:
					m := safeGetMap(item)
					ref, refExists := m["ref"].(string)
					if !refExists {
						patternsContent += fmt.Sprintf("\t// 错误: 第%d项缺少ref字段\n", idx)
						continue
					}

					count := 1
					if cnt, ok := m["count"].(float64); ok {
						count = int(cnt)
					}

					patternsContent += "\n\tparams = {};\n"
					patternsContent += fmt.Sprintf("\tfor (int i = %d; i < %d; i++) {\n", paramIndex, paramIndex+count)
					patternsContent += "\t\tparams.push_back(pattern_params[i]);\n"
					patternsContent += "\t};\n"
					patternsContent += fmt.Sprintf("    for(int i = 0; i < %d; i++) {\n", count)
					patternsContent += NameToCode(ref, 2, "i")
					patternsContent += "    }\n"
					paramIndex += count
				default:
					patternsContent += fmt.Sprintf("\t// 未知类型的项: %v\n", item)
				}
			}
		}
		patternsContent += "}\n\n"
	}

	return patternsHead + patternsContent
}

// 生成序列代码 - 增强健壮性
func (c *Converter) generateSequenceCode(sequenceData []jquantum.Element, patterns map[string]jquantum.PatternContent) string {
	if sequenceData == nil || len(sequenceData) == 0 {
		return "\t// 无量子门序列\n"
	}

	sequenceCode := ""
	instructionIndex := 0

	for _, item := range sequenceData {
		switch {
		case safeGetString(item) != "":
			gateName := safeGetString(item)
			sequenceCode += fmt.Sprintf("\tparams = get_gate_params(compressed, %d);\n", instructionIndex)
			sequenceCode += NameToCode(gateName, 1, "0") + "\n"
			instructionIndex++
		case safeGetMap(item) != nil:
			m := safeGetMap(item)
			ref, refExists := m["ref"].(string)
			if !refExists {
				sequenceCode += "\t// 错误: 序列项缺少ref字段\n"
				continue
			}

			count := 1
			if cnt, ok := m["count"].(float64); ok {
				count = int(cnt)
			}

			// 检查是否是模式引用
			if pattern, exists := patterns[ref]; exists {
				patternsTotalGatesNum := pattern.Total * count
				sequenceCode += fmt.Sprintf("    for(int i = %d; i < %d; i+=%d) {\n",
					instructionIndex, instructionIndex+patternsTotalGatesNum, pattern.Total)
				sequenceCode += fmt.Sprintf("        params = get_gate_params(compressed, i, %d);\n", pattern.Total)
				sequenceCode += "        " + ref + "(qureg, params);\n"
				sequenceCode += "    }\n\n"
				instructionIndex += patternsTotalGatesNum
			} else {
				// 普通门引用
				if count > 1 {
					sequenceCode += fmt.Sprintf("\tparams = get_repeats_params(get_gate_params(compressed, %d, %d), %d);\n",
						instructionIndex, count, count)
					sequenceCode += fmt.Sprintf("    for(int i = 0; i < %d; i++) {\n", count)
					sequenceCode += NameToCode(ref, 2, "i")
					sequenceCode += "    }\n\n"
					instructionIndex += count
				} else {
					sequenceCode += fmt.Sprintf("\tparams = get_gate_params(compressed, %d);\n", instructionIndex)
					sequenceCode += NameToCode(ref, 1, "0")
					instructionIndex++
				}
			}
		default:
			sequenceCode += fmt.Sprintf("\t// 未知序列项类型: %v\n", item)
		}
	}

	return sequenceCode
}

// 将量子电路转换为QuEST C++模拟代码
func (c *Converter) circuitToQuestJSON(result jquantum.ResultJSON) string {
	numQubits := result.NumQubits
	patterns := result.Patterns

	// 创建代码模板
	code := CodeTemplateBeginning

	// 添加模式代码（即使为空也会处理）
	code += c.generatePatternsCode(patterns)

	// 主函数
	code += MainCodeTemplateBeginning(numQubits, "./user")

	// 添加序列代码
	sequenceCode := c.generateSequenceCode(result.Sequence, patterns)
	code += sequenceCode

	code += MainCodeTemplateEnding(numQubits)

	return code
}

// 读取JSON文件内容
func readJSONFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("无法读取JSON文件 %s: %v", filename, err)
	}
	return data, nil
}

// 主函数：从JSON文件加载数据并生成代码
func main() {
	//if len(os.Args) < 2 {
	//	fmt.Println("用法: go-program <json文件路径>")
	//	fmt.Println("示例: go-program circuit.json")
	//	os.Exit(1)
	//}

	//jsonFile := os.Args[1]

	// 读取JSON文件
	//jsonData, err := readJSONFile(jsonFile)
	//if err != nil {
	//	log.Fatalf("错误: %v", err)
	//}

	jsonData := "{\"num_qubits\": 18, \"patterns\": {\"pattern_1\": {\"content\": [{\"count\": 18, \"ref\": \"h\"},\n                                        {\"count\": 13, \"ref\": \"x\"},\n                                        \"h\",\n                                        \"mcx\",\n                                        \"h\",\n                                        {\"count\": 26, \"ref\": \"x\"},\n                                        \"h\",\n                                        \"mcx\",\n                                        \"h\",\n                                        {\"count\": 13, \"ref\": \"x\"},\n                                        {\"count\": 18, \"ref\": \"h\"},\n                                        {\"count\": 18, \"ref\": \"x\"},\n                                        \"h\",\n                                        \"mcx\",\n                                        \"h\",\n                                        {\"count\": 18, \"ref\": \"x\"}],\n                            \"count\": 16,\n                            \"total\": 133}},\n \"sequence\": [{\"count\": 284, \"ref\": \"pattern_1\"}, {\"count\": 18, \"ref\": \"h\"}]}"

	// 解析JSON数据
	var result jquantum.ResultJSON
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		log.Fatalf("解析JSON失败: %v", err)
	}

	// 设置默认值（如果某些字段缺失）
	if result.NumQubits == 0 {
		result.NumQubits = 1 // 默认值
	}
	if result.Sequence == nil {
		result.Sequence = []jquantum.Element{}
	}
	if result.Patterns == nil {
		result.Patterns = map[string]jquantum.PatternContent{}
	}

	// 输出解析的信息
	fmt.Printf("解析成功: %d 量子比特, %d 模式, %d 序列项\n",
		result.NumQubits, len(result.Patterns), len(result.Sequence))

	// 创建转换器并生成代码
	converter := &Converter{}
	questCode := converter.circuitToQuestJSON(result)

	// 输出生成的代码到文件
	outputFile := "generated_quest.cpp"
	err = os.WriteFile(outputFile, []byte(questCode), 0644)
	if err != nil {
		log.Fatalf("写入输出文件失败: %v", err)
	}

	fmt.Printf("代码已生成到: %s\n", outputFile)
	fmt.Printf("量子比特数: %d\n", result.NumQubits)
	if len(result.Patterns) == 0 {
		fmt.Println("提示: 电路较小，未检测到重复模式")
	}
}
