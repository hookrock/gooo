package gooo

type trie struct {
	part     string
	pattern  string // 完整路径模式
	isWild   bool
	children []*trie
}

// 匹配子节点
func (t *trie) matchChild(part string) *trie {
	for _, child := range t.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 匹配所有子节点
// 保留此方法用于未来可能的扩展功能，如：
// - 多匹配路由
// - 路由回溯
// - 更复杂的通配符处理
// func (t *trie) matchChildren(part string) []*trie {
// 	nodes := make([]*trie, 0)
// 	for _, child := range t.children {
// 		if child.part == part || child.isWild {
// 			nodes = append(nodes, child)
// 		}
// 	}
// 	return nodes
// }

// 插入路径到前缀树
func (t *trie) insert(pattern string, parts []string) {
	node := t
	for i, part := range parts {
		child := node.matchChild(part)
		if child == nil {
			// 验证通配符节点只能出现在末尾
			if part[0] == '*' && i != len(parts)-1 {
				panic("通配符路由必须位于路径末尾")
			}
			child = &trie{
				part:     part,
				isWild:   part[0] == ':' || part[0] == '*',
				children: make([]*trie, 0),
			}
			node.children = append(node.children, child)
		} else {
			// 检查路由冲突
			if child.part != part && (!child.isWild && !(part[0] == ':' || part[0] == '*')) {
				panic("路由冲突: " + pattern + " 与 " + child.pattern)
			}
		}
		node = child
	}
	// 检查是否重复注册相同路径
	if node.pattern != "" {
		panic("重复注册路由: " + pattern)
	}
	node.pattern = pattern
}

// 搜索匹配的路由节点
func (t *trie) search(parts []string) *trie {
	node := t
	for _, part := range parts {
		child := node.matchChild(part)
		if child == nil {
			return nil
		}
		node = child

		// 遇到通配符节点时提前返回
		if node.part[0] == '*' {
			return node
		}
	}

	// 只有完整匹配的节点才返回
	if node.pattern == "" {
		return nil
	}
	return node
}
