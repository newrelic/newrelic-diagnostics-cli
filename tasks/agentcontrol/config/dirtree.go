package config

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// AgentControlDirTree collects a directory tree of agent-control installation paths.
type AgentControlDirTree struct{}

func (p AgentControlDirTree) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("AgentControl/Config/DirTree")
}

func (p AgentControlDirTree) Explain() string {
	return "Collect directory tree of New Relic agent-control installation paths"
}

func (p AgentControlDirTree) Dependencies() []string {
	return []string{"AgentControl/Config/Agent"}
}

func (p AgentControlDirTree) Execute(_ tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["AgentControl/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Agent Control not detected on system.",
		}
	}

	roots := acCollectLinux
	if runtime.GOOS == "windows" {
		roots = acCollectWindows
	}
	var dirs []string
	for _, r := range roots {
		dirs = append(dirs, r.base)
	}

	var sb strings.Builder
	anyFound := false

	for _, dir := range dirs {
		if !tasks.FileExists(dir) {
			log.Debug("agent-control directory not found: " + dir)
			sb.WriteString(dir + " (not found)\n\n")
			continue
		}
		anyFound = true
		sb.WriteString(buildDirTree(dir))
		sb.WriteString("\n")
	}

	if !anyFound {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: fmt.Sprintf("No agent-control directories found in %v.", dirs),
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(sb.String(), stream)

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Collected agent-control directory tree.",
		FilesToCopy: []tasks.FileCopyEnvelope{{
			Path:       "agent-control-dirtree.txt",
			Stream:     stream,
			Identifier: "AgentControl/Config/DirTree",
		}},
	}
}

type treeNode struct {
	name     string
	isDir    bool
	size     int64
	err      error
	children []*treeNode
}

func buildDirTree(root string) string {
	rootNode := &treeNode{name: root, isDir: true}
	nodeByPath := map[string]*treeNode{root: rootNode}

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if path == root {
			return nil
		}
		parent := filepath.Dir(path)
		parentNode, ok := nodeByPath[parent]
		if !ok {
			return nil
		}
		node := &treeNode{name: d.Name(), isDir: d.IsDir(), err: err}
		if err == nil && !d.IsDir() {
			if info, infoErr := d.Info(); infoErr == nil {
				node.size = info.Size()
			}
		}
		nodeByPath[path] = node
		parentNode.children = append(parentNode.children, node)
		return nil
	})

	var sb strings.Builder
	sb.WriteString(root + "\n")
	renderNode(&sb, rootNode, "")
	return sb.String()
}

func renderNode(sb *strings.Builder, node *treeNode, prefix string) {
	for i, child := range node.children {
		isLast := i == len(node.children)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}
		label := child.name
		if child.err != nil {
			label += fmt.Sprintf(" (error: %s)", child.err.Error())
		} else if child.isDir {
			label += "/"
		} else {
			label += fmt.Sprintf(" (%s)", formatSize(child.size))
		}
		sb.WriteString(prefix + connector + label + "\n")
		renderNode(sb, child, childPrefix)
	}
}

func formatSize(b int64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	}
}
