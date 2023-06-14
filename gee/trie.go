package gee

import "strings"

type node struct {
	pattern  string  // 待匹配路由，如 /p/:lang
	part     string  // 路由中的一部分，如 :lang
	children []*node // 子节点，如[doc, tutorial, intro]
	isWild   bool    // 是否精确匹配，part 含有 : 或 * 为true
}

// matchChild 第一个匹配成功的节点，用于插入，只用于 insert
// 用于在当前节点的子节点集合中查找与指定字符串 part 匹配的节点，
// 并返回匹配到的第一个节点。如果没有匹配成功，则返回 nil
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild { // 节点相同，或者含有: 或*
			return child
		}
	}
	return nil
}

// insert 插入功能，递归查找每一层节点，没有匹配到当前part的节点，则新建一个
// 先将输入的路径切割成一个字符串切片 parts，然后逐层遍历节点，在每一级上寻找能够匹配当前路径元素的子节点。
// 如果找到了能够匹配当前路径元素的子节点，则继续递归向下查找，否则则新建一个节点并将其插入到当前节点的子节点集合中
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height { // 达到目的位置，返回
		n.pattern = pattern
		return
	}

	part := parts[height]       // 提取当前part
	child := n.matchChild(part) // 根据part匹配子节点
	if child == nil {           // 没有子节点就创建添加
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// matchChildren 所有匹配成功的节点，用于查找，只在 search 函数中使用
// 在当前节点的子节点集合中查找与 part 匹配的所有节点，
// 并将其保存在切片 nodes 中，并返回此切片
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child) // 找出所有的匹配
		}
	}
	return nodes
}

// search 查询，递归查询每一层的节点， 直到匹配到*，匹配失败，或者匹配到最深层
// 在 trie 树中查找与给定路径匹配的最深节点。
func (n *node) search(parts []string, height int) *node {
	// 首先判断当前路径元素是否为最后一个路径元素，或者当前节点是否为通配符 *。
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" { // 待匹配路由为空
			return nil
		}
		return n // 返回当前节点
	}
	// 否则在子节点集合中查找所有能够匹配当前路径元素的节点，并依次递归进入下一层级查找
	part := parts[height]
	children := n.matchChildren(part) // 对子节点进行匹配，如果是相同的，或者: 或者* 则加入新集合，继续匹配
	for _, child := range children {
		result := child.search(parts, height+1)
		// 如果找到了匹配的叶子节点，则返回该节点，否则返回 nil
		if result != nil {
			return result
		}
	}
	return nil
}

// travel 用于遍历 trie 树上所有叶子节点
// 接收一个指向节点切片的指针作为参数
// 最终会将遍历得到的所有叶子节点找出存储在节点切片中返回给调用者
func (n *node) travel(list *[]*node) {
	if n.pattern != "" { // 遍历所有子节点，非空存入节点切片
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list) // 对当前节点的每个子节点进行递归调用，并将节点切片指针传入下一层级函数中
	}
}
