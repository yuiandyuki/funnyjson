package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// JsonElement 抽象组件接口
type JsonElement interface {
	SetIcon(iconFamily IconFamily)
}

// JsonObject 组合组件
type JsonObject struct {
	keys   []string
	values []JsonElement
}

func NewJsonObject() *JsonObject {
	return &JsonObject{}
}

func (j *JsonObject) Add(key string, value JsonElement) {
	j.keys = append(j.keys, key)
	j.values = append(j.values, value)
}

func (j *JsonObject) GetKeys() []string {
	return j.keys
}

func (j *JsonObject) GetValues() []JsonElement {
	return j.values
}

func (j *JsonObject) SetIcon(iconFamily IconFamily) {
	internalIcon := iconFamily.GetInternalNodeIcon()
	leafIcon := iconFamily.GetLeafNodeIcon()
	for i, value := range j.values {
		prefix := leafIcon
		if _, ok := value.(*JsonObject); ok {
			prefix = internalIcon
			value.SetIcon(iconFamily)
		}
		j.keys[i] = prefix + j.keys[i]
	}
}

// JsonValue 叶子组件
type JsonValue struct {
	value string
}

func NewJsonValue(value string) *JsonValue {
	return &JsonValue{value: value}
}

func (j *JsonValue) GetValue() string {
	return j.value
}

func (j *JsonValue) SetIcon(iconFamily IconFamily) {
	// JsonValue不需要设置图标
}

// IconFamily 接口定义（抽象产品）
type IconFamily interface {
	GetInternalNodeIcon() string
	GetLeafNodeIcon() string
}

// PokerFaceIconFamily 具体产品实现
type PokerFaceIconFamily struct{}

func (p *PokerFaceIconFamily) GetInternalNodeIcon() string {
	return "♢"
}

func (p *PokerFaceIconFamily) GetLeafNodeIcon() string {
	return "♤"
}

// JsonIconFamily 具体产品实现
type JsonIconFamily struct {
	internalNodeIcon string
	leafNodeIcon     string
}

func NewJsonIconFamily(internalNodeIcon, leafNodeIcon string) *JsonIconFamily {
	return &JsonIconFamily{
		internalNodeIcon: internalNodeIcon,
		leafNodeIcon:     leafNodeIcon,
	}
}

func (j *JsonIconFamily) GetInternalNodeIcon() string {
	return j.internalNodeIcon
}

func (j *JsonIconFamily) GetLeafNodeIcon() string {
	return j.leafNodeIcon
}

// Style 接口定义
type Style interface {
	Render(jsonData JsonElement) string
}

// TreeStyle 具体产品
type TreeStyle struct{}

func (t *TreeStyle) Render(jsonData JsonElement) string {
	return t.renderTree(jsonData, "")
}

func (t *TreeStyle) renderTree(data JsonElement, prefix string) string {
	var result string
	if obj, ok := data.(*JsonObject); ok {
		keys := obj.GetKeys()
		values := obj.GetValues()
		for i, key := range keys {
			value := values[i]
			if childObj, ok := value.(*JsonObject); ok {
				result += prefix
				newPrefix := prefix
				if i == len(keys)-1 {
					result += "└─"
					newPrefix += "   "
				} else {
					result += "├─"
					newPrefix += "│  "
				}
				result += key + "\n"
				result += t.renderTree(childObj, newPrefix)
			} else if childValue, ok := value.(*JsonValue); ok {
				result += prefix
				if i == len(keys)-1 {
					result += "└─"
				} else {
					result += "├─"
				}
				result += key
				if childValue.GetValue() != "null" {
					result += ": " + childValue.GetValue()
				}
				result += "\n"
			} else {
				result += prefix
				if i == len(keys)-1 {
					result += "└─"
				} else {
					result += "├─"
				}
				result += key
				result += "\n"
			}
		}
	}
	return result
}

// RectangleStyle 具体产品
type RectangleStyle struct {
	displayLength int
}

func (r *RectangleStyle) Render(jsonData JsonElement) string {
	r.renderRectangle(jsonData, "", true)
	result := r.renderRectangle(jsonData, "", false)
	result = strings.Replace(result, "├", "┌", 1)
	result = strings.Replace(result, "┤", "┐", 1)

	// 找到最后一行的起始位置
	lastLineIndex := strings.LastIndex(result, "\n")
	secondLastLineIndex := strings.LastIndex(result[:lastLineIndex], "\n")
	lastLine := result[secondLastLineIndex+1:]
	modifiedLastLine := strings.ReplaceAll(lastLine, "├─", "└─")
	modifiedLastLine = strings.ReplaceAll(modifiedLastLine, "│ ", "└─")
	//result = strings.Replace(result, "┤", "┘", -1)
	result = result[:secondLastLineIndex+1] + modifiedLastLine
	lastIndex := strings.LastIndex(result, "┤")

	if lastIndex != -1 {
		result = result[:lastIndex] + "┘" + result[lastIndex+len("┤"):]
	}
	return result

}

func (r *RectangleStyle) renderRectangle(data JsonElement, prefix string, getDisplayLength bool) string {
	var result string
	if obj, ok := data.(*JsonObject); ok {
		keys := obj.GetKeys()
		values := obj.GetValues()
		for i, key := range keys {
			value := values[i]
			var curRow string
			curRow += prefix
			curRow += "├─"
			curRow += key
			if childValue, ok := value.(*JsonValue); ok {
				if childValue.GetValue() != "null" {
					curRow += ": " + childValue.GetValue()
				}
			}
			if getDisplayLength {
				r.displayLength = max(r.displayLength, r.calculateDisplayWidth(curRow))
			} else {
				num := r.displayLength + 10 - r.calculateDisplayWidth(curRow)
				for k := 0; k < num; k++ {
					curRow += "─"
				}
			}
			curRow += "┤\n"
			result += curRow
			if childObj, ok := value.(*JsonObject); ok {
				result += r.renderRectangle(childObj, prefix+"│  ", getDisplayLength)
			}
		}
	}
	return result
}

func (r *RectangleStyle) calculateDisplayWidth(str string) int {
	width := 0.0
	for _, ch := range str {
		if ch > 127 {
			width++
		} else {
			width++
		}
	}
	return int(width + 1.0/3)
}

// StyleFactory 接口定义
type StyleFactory interface {
	CreateStyle() Style
}

// TreeStyleFactory 具体工厂
type TreeStyleFactory struct{}

func (t *TreeStyleFactory) CreateStyle() Style {
	return &TreeStyle{}
}

// RectangleStyleFactory 具体工厂
type RectangleStyleFactory struct{}

func (r *RectangleStyleFactory) CreateStyle() Style {
	return &RectangleStyle{}
}

// IconFamilyFactory 接口定义
type IconFamilyFactory interface {
	CreateIconFamily() IconFamily
}

// PokerFaceIconFamilyFactory 具体工厂
type PokerFaceIconFamilyFactory struct{}

func (p *PokerFaceIconFamilyFactory) CreateIconFamily() IconFamily {
	return &PokerFaceIconFamily{}
}

// JsonIconFamilyFactory 具体工厂
type JsonIconFamilyFactory struct {
	internalNodeIcon string
	leafNodeIcon     string
}

func NewJsonIconFamilyFactory(internalNodeIcon, leafNodeIcon string) *JsonIconFamilyFactory {
	return &JsonIconFamilyFactory{
		internalNodeIcon: internalNodeIcon,
		leafNodeIcon:     leafNodeIcon,
	}
}

func (j *JsonIconFamilyFactory) CreateIconFamily() IconFamily {
	return NewJsonIconFamily(j.internalNodeIcon, j.leafNodeIcon)
}

// JsonLoader 负责加载和解析JSON文件
type JsonLoader struct{}

func (loader *JsonLoader) LoadJson(filePath string) (JsonElement, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open JSON file: %v", err)
	}

	strData := string(data)
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(strData), &obj); err != nil {
		return nil, err
	}
	return parseJsonObject(obj), nil
}

func parseJsonObject(data map[string]interface{}) *JsonObject {
	obj := NewJsonObject()
	for key, value := range data {
		switch value.(type) {
		case string:
			obj.Add(key, NewJsonValue(value.(string)))
		case map[string]interface{}:
			obj.Add(key, parseJsonObject(value.(map[string]interface{})))
		case nil:
			obj.Add(key, nil)
		}
	}

	return obj
}

// VisualizationBuilder 建造者接口
type VisualizationBuilder interface {
	SetStyle(style Style)
	SetIconFamily(iconFamily IconFamily)
	SetJsonData(jsonData JsonElement)
	Build() (string, error)
}

// ConcreteVisualizationBuilder 具体的建造者
type ConcreteVisualizationBuilder struct {
	style      Style
	iconFamily IconFamily
	jsonData   JsonElement
}

func (b *ConcreteVisualizationBuilder) SetStyle(style Style) {
	b.style = style
}

func (b *ConcreteVisualizationBuilder) SetIconFamily(iconFamily IconFamily) {
	b.iconFamily = iconFamily
}

func (b *ConcreteVisualizationBuilder) SetJsonData(jsonData JsonElement) {
	b.jsonData = jsonData
}

func (b *ConcreteVisualizationBuilder) Build() (string, error) {
	if b.style == nil || b.iconFamily == nil || b.jsonData == nil {
		return "", fmt.Errorf("Style, icon family, and JSON data must be set before building.")
	}

	if obj, ok := b.jsonData.(*JsonObject); ok {
		obj.SetIcon(b.iconFamily)
	}

	renderedOutput := b.style.Render(b.jsonData)
	return renderedOutput, nil
}

// VisualizationDirector 指导者类
type VisualizationDirector struct {
	builder VisualizationBuilder
}

func (d *VisualizationDirector) SetBuilder(builder VisualizationBuilder) {
	d.builder = builder
}

func (d *VisualizationDirector) Construct(jsonData JsonElement) (string, error) {
	if d.builder == nil {
		return "", fmt.Errorf("Builder is not set.")
	}
	d.builder.SetJsonData(jsonData)
	return d.builder.Build()
}

func main() {
	if len(os.Args) < 7 {
		fmt.Println("Usage: fje -f <json file> -s <style> -i <icon family>")
		return
	}

	jsonFile := os.Args[2]
	styleName := os.Args[4]
	iconFamilyName := os.Args[6]

	// 获取json对象
	loader := &JsonLoader{}
	jsonData, err := loader.LoadJson(jsonFile)
	if err != nil {
		fmt.Println("Error loading JSON:", err)
		return
	}

	// 使用工厂方法模式创建风格对象
	var style Style
	switch styleName {
	case "tree":
		style = (&TreeStyleFactory{}).CreateStyle()
	case "rectangle":
		style = (&RectangleStyleFactory{}).CreateStyle()
	default:
		fmt.Println("Unknown style:", styleName)
		return
	}

	// 使用抽象工厂创建icon对象
	var iconFamily IconFamily
	switch iconFamilyName {
	case "poker-face":
		iconFamily = (&PokerFaceIconFamilyFactory{}).CreateIconFamily()
	case "json_defined":
		iconData, err := loader.LoadJson("icon.json")
		if err != nil {
			fmt.Println("Error loading icon JSON:", err)
			return
		}
		iconObj, ok := iconData.(*JsonObject)
		if !ok {
			fmt.Println("Invalid icon JSON format")
			return
		}
		keys := iconObj.GetKeys()
		values := iconObj.GetValues()
		var internalNodeIcon, leafNodeIcon string
		if keys[0] == "internalNodeIcon" && keys[1] == "leafNodeIcon" {
			internalNodeIcon = values[0].(*JsonValue).GetValue()
			leafNodeIcon = values[1].(*JsonValue).GetValue()
		} else if keys[0] == "leafNodeIcon" && keys[1] == "internalNodeIcon" {
			internalNodeIcon = values[1].(*JsonValue).GetValue()
			leafNodeIcon = values[0].(*JsonValue).GetValue()
		} else {
			internalNodeIcon = "+"
			leafNodeIcon = "-"
		}
		iconFamily = NewJsonIconFamilyFactory(internalNodeIcon, leafNodeIcon).CreateIconFamily()
	default:
		fmt.Println("Unknown icon family:", iconFamilyName)
		return
	}

	// 使用建造者模式创建可视化对象
	builder := &ConcreteVisualizationBuilder{}
	director := &VisualizationDirector{}
	director.SetBuilder(builder)

	// 设置建造者的参数
	builder.SetStyle(style)
	builder.SetIconFamily(iconFamily)

	result, err := director.Construct(jsonData)
	if err != nil {
		fmt.Println("Error constructing visualization:", err)
		return
	}

	fmt.Println(result)
}
