package main

// Workspace state types — serialized as JSON string inside the TOML config's `workspace` field.
// Single source of truth: ~/.zeno.toml

type WorkspacePane struct {
	Type      string         `json:"type"`
	SessionID string         `json:"sessionId,omitempty"`
	Direction string         `json:"direction,omitempty"`
	Ratio     float64        `json:"ratio,omitempty"`
	First     *WorkspacePane `json:"first,omitempty"`
	Second    *WorkspacePane `json:"second,omitempty"`
}

type WorkspaceTab struct {
	Title    string         `json:"title"`
	RootNode *WorkspacePane `json:"rootNode"`
}

type WorkspaceState struct {
	Version        int            `json:"version"`
	ActiveTabIndex int            `json:"activeTabIndex"`
	Tabs           []WorkspaceTab `json:"tabs"`
}
