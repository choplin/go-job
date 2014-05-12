package report

type ReporterConfig struct {
	Reporters        string
	FluentdHost      string
	FluentdPort      int
	FluentdTagPrefix string
	FileDirectory    string
}
