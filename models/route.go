package models

import (
	"strings"
)

type Route struct {
	path       string             // The segment of the URL stored in this node
	children   []*Route           // Child nodes
	handlers   map[string]Handler // method -> handler (e.g., "GET" -> func)
	isParam    bool               // Flag for :id style segments
	isWildcard bool
	paramName  string // Stores "id" if the path was ":id"
}

func (n *Route) Insert(method, path string, handler Handler) {
	path = strings.Trim(path, "/")
	segments := []string{}
	if path != "" {
		segments = strings.Split(path, "/")
	}
	curr := n

	for _, seg := range segments {
		var next *Route
		// Check if segment is a parameter
		isParam := strings.HasPrefix(seg, ":")
		isWildcard := seg == "*" // <--- Check for wildcard

		// Find existing child
		for _, child := range curr.children {
			if child.path == seg {
				next = child
				break
			}
		}

		// Create new node if it doesn't exist
		if next == nil {
			next = &Route{
				path:       seg,
				isParam:    isParam,
				isWildcard: isWildcard, // <--- Set flag
				handlers:   make(map[string]Handler),
			}
			if isParam {
				next.paramName = seg[1:] // strip the ":"
			}
			curr.children = append(curr.children, next)
		}
		curr = next

		// If this was a wildcard, we stop here.
		// Nothing can come after a wildcard.
		if isWildcard {
			break
		}
	}
	curr.handlers[method] = handler
}

func (n *Route) Search(method, path string) (bool, Handler, map[string]string) {
	path = strings.Trim(path, "/")
	segments := []string{}
	if path != "" {
		segments = strings.Split(path, "/")
	}
	params := make(map[string]string)

	curr := n

	for i, seg := range segments {
		var next *Route
		var wildcardChild *Route

		for _, child := range curr.children {
			// Exact match
			if child.path == seg {
				next = child
				break
			}
			// Wildcard match
			if child.isWildcard {
				wildcardChild = child
			}
			// Param match
			if child.isParam {
				next = child
			}

		}

		// If no exact/param match, but we have a wildcard, we are done!
		if next == nil && wildcardChild != nil {
			params["*"] = strings.Join(segments[i:], "/")
			handler, ok := wildcardChild.handlers[method]
			return ok, handler, params
		}

		if next == nil {
			return false, nil, nil
		}

		if next.isParam {
			params[next.paramName] = seg
		}

		curr = next

		// If the node we just moved to is a wildcard node,
		// it consumes the rest of the segments.
		if curr.isWildcard {
			params["*"] = strings.Join(segments[i:], "/")
			break
		}
	}

	if curr.handlers[method] == nil {
		for _, child := range curr.children {
			if child.isWildcard {
				params["*"] = "" // Empty path inside the directory
				h, ok := child.handlers[method]
				return ok, h, params
			}
		}
	}

	handler, ok := curr.handlers[method]
	return ok, handler, params
}
