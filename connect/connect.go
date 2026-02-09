package connect

type DbConnStr struct {
	SrcHost      string
	SrcUserName  string
	SrcPassword  string
	SrcDatabase  string
	SrcSchema    string
	SrcPort      uint64
	SrcParams    map[string]string
	DestHost     string
	DestPort     int
	DestUserName string
	DestPassword string
	DestDatabase string
	DestParams   map[string]string
}
