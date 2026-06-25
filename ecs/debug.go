package ecs

type ArchetypeReport struct {
	Archetypes []ArchetypeInfo
}

type ArchetypeInfo struct {
	ID         uint32
	Components []string
	Entities   int
}

type QueryDebugInfo struct {
	System  string
	Query   string
	With    []string
	Without []string
	Reads   []string
	Writes  []string
}

type AccessReport struct {
	Stages []StageAccessReport
}

type StageAccessReport struct {
	Stage   StageID
	Systems []SystemAccessInfo
}

type SystemAccessInfo struct {
	System  string
	Reads   []string
	Writes  []string
	Queries []QueryDebugInfo
}
