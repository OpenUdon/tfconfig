package tfconfig

const (
	// StaticV1 is the version string for the initial tfconfig static fact model.
	StaticV1 = "tfconfig.static.v1"

	// DefaultProducer identifies JSON written by this package when callers do
	// not provide a more specific producer name.
	DefaultProducer = "github.com/OpenUdon/tfconfig"
)

// Document is the top-level static Terraform/OpenTofu configuration fact model.
//
// The model is static and review-oriented. It represents source facts,
// expressions, references, diagnostics, and module structure without provider
// plugin execution, state, refresh, plan, apply, credential resolution, or
// Terraform/OpenTofu runtime evaluation.
type Document struct {
	Version     string       `json:"version"`
	Producer    string       `json:"producer,omitempty"`
	RootDir     string       `json:"root_dir"`
	SourceRoots []SourceRoot `json:"source_roots,omitempty"`
	SourceFiles []SourceFile `json:"source_files,omitempty"`
	Modules     []Module     `json:"modules,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// SourceRoot describes a normalized source root used by source files and
// ranges. The root module normally uses the empty module address.
type SourceRoot struct {
	ID            string `json:"id"`
	ModuleAddress string `json:"module_address"`
	Dir           string `json:"dir"`
}

type FileFormat string

const (
	FileFormatHCL  FileFormat = "hcl"
	FileFormatJSON FileFormat = "json"
)

type FileRole string

const (
	FileRolePrimary  FileRole = "primary"
	FileRoleOverride FileRole = "override"
	FileRoleTest     FileRole = "test"
)

// SourceFile describes a Terraform/OpenTofu source file that contributed facts
// or diagnostics to the model.
type SourceFile struct {
	ID            string     `json:"id"`
	ModuleAddress string     `json:"module_address"`
	Path          string     `json:"path"`
	Format        FileFormat `json:"format"`
	Role          FileRole   `json:"role"`
}

// SourceRange uses 1-based line and column numbers. Byte offsets are optional
// because HCL and HCL JSON parsers differ in what range information is
// available.
type SourceRange struct {
	SourceID string   `json:"source_id,omitempty"`
	Path     string   `json:"path,omitempty"`
	Start    Position `json:"start"`
	End      Position `json:"end"`
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Byte   int `json:"byte,omitempty"`
}

type ModuleStatus string

const (
	ModuleStatusRoot        ModuleStatus = "root"
	ModuleStatusLoaded      ModuleStatus = "loaded"
	ModuleStatusMissing     ModuleStatus = "missing"
	ModuleStatusRemote      ModuleStatus = "remote"
	ModuleStatusUnsupported ModuleStatus = "unsupported"
)

// Module represents one static module in the configuration tree.
type Module struct {
	Address           string                `json:"address"`
	Source            *Value                `json:"source,omitempty"`
	Dir               string                `json:"dir"`
	ParentAddress     string                `json:"parent_address,omitempty"`
	Status            ModuleStatus          `json:"status"`
	SourceFiles       []string              `json:"source_files,omitempty"`
	RequiredVersions  []Value               `json:"required_versions,omitempty"`
	RequiredProviders []ProviderRequirement `json:"required_providers,omitempty"`
	ProviderConfigs   []ProviderConfig      `json:"provider_configs,omitempty"`
	Variables         []Variable            `json:"variables,omitempty"`
	Locals            []Local               `json:"locals,omitempty"`
	Outputs           []Output              `json:"outputs,omitempty"`
	Resources         []Resource            `json:"resources,omitempty"`
	DataSources       []DataSource          `json:"data_sources,omitempty"`
	ModuleCalls       []ModuleCall          `json:"module_calls,omitempty"`
	Moved             []MovedBlock          `json:"moved,omitempty"`
	Imports           []ImportBlock         `json:"imports,omitempty"`
	Removed           []RemovedBlock        `json:"removed,omitempty"`
	Checks            []CheckBlock          `json:"checks,omitempty"`
	Tests             []TestFile            `json:"tests,omitempty"`
	Diagnostics       []Diagnostic          `json:"diagnostics,omitempty"`
	Range             *SourceRange          `json:"range,omitempty"`
}

type Variable struct {
	Name        string       `json:"name"`
	Type        *Value       `json:"type,omitempty"`
	Default     *Value       `json:"default,omitempty"`
	Description string       `json:"description,omitempty"`
	Sensitive   bool         `json:"sensitive,omitempty"`
	Nullable    *bool        `json:"nullable,omitempty"`
	References  []Reference  `json:"references,omitempty"`
	Range       *SourceRange `json:"range,omitempty"`
}

type Local struct {
	Name       string       `json:"name"`
	Value      *Value       `json:"value,omitempty"`
	References []Reference  `json:"references,omitempty"`
	Range      *SourceRange `json:"range,omitempty"`
}

type Output struct {
	Name        string       `json:"name"`
	Value       *Value       `json:"value,omitempty"`
	Description string       `json:"description,omitempty"`
	Sensitive   bool         `json:"sensitive,omitempty"`
	DependsOn   []Reference  `json:"depends_on,omitempty"`
	References  []Reference  `json:"references,omitempty"`
	Range       *SourceRange `json:"range,omitempty"`
}

type ProviderRequirement struct {
	LocalName          string       `json:"local_name"`
	Source             string       `json:"source,omitempty"`
	VersionConstraints []string     `json:"version_constraints,omitempty"`
	Range              *SourceRange `json:"range,omitempty"`
}

type ProviderConfig struct {
	LocalName  string       `json:"local_name"`
	Alias      string       `json:"alias,omitempty"`
	Address    string       `json:"address"`
	Config     []Attribute  `json:"config,omitempty"`
	References []Reference  `json:"references,omitempty"`
	Range      *SourceRange `json:"range,omitempty"`
}

type ProviderRef struct {
	LocalName string       `json:"local_name"`
	Alias     string       `json:"alias,omitempty"`
	Address   string       `json:"address,omitempty"`
	Range     *SourceRange `json:"range,omitempty"`
}

type Resource struct {
	Address    string       `json:"address"`
	Type       string       `json:"type"`
	Name       string       `json:"name"`
	Provider   *ProviderRef `json:"provider,omitempty"`
	Config     []Attribute  `json:"config,omitempty"`
	Lifecycle  *Lifecycle   `json:"lifecycle,omitempty"`
	DependsOn  []Reference  `json:"depends_on,omitempty"`
	Count      *Value       `json:"count,omitempty"`
	ForEach    *Value       `json:"for_each,omitempty"`
	References []Reference  `json:"references,omitempty"`
	Range      *SourceRange `json:"range,omitempty"`
}

type DataSource struct {
	Address    string       `json:"address"`
	Type       string       `json:"type"`
	Name       string       `json:"name"`
	Provider   *ProviderRef `json:"provider,omitempty"`
	Config     []Attribute  `json:"config,omitempty"`
	DependsOn  []Reference  `json:"depends_on,omitempty"`
	Count      *Value       `json:"count,omitempty"`
	ForEach    *Value       `json:"for_each,omitempty"`
	References []Reference  `json:"references,omitempty"`
	Range      *SourceRange `json:"range,omitempty"`
}

type ModuleCall struct {
	Address          string            `json:"address"`
	Name             string            `json:"name"`
	Source           *Value            `json:"source,omitempty"`
	Inputs           []Attribute       `json:"inputs,omitempty"`
	ProviderMappings []ProviderMapping `json:"provider_mappings,omitempty"`
	DependsOn        []Reference       `json:"depends_on,omitempty"`
	Count            *Value            `json:"count,omitempty"`
	ForEach          *Value            `json:"for_each,omitempty"`
	References       []Reference       `json:"references,omitempty"`
	Range            *SourceRange      `json:"range,omitempty"`
}

type ProviderMapping struct {
	ChildName string       `json:"child_name"`
	Provider  ProviderRef  `json:"provider"`
	Range     *SourceRange `json:"range,omitempty"`
}

type Attribute struct {
	Path  string       `json:"path"`
	Value Value        `json:"value"`
	Range *SourceRange `json:"range,omitempty"`
}

type Lifecycle struct {
	CreateBeforeDestroy *Value       `json:"create_before_destroy,omitempty"`
	PreventDestroy      *Value       `json:"prevent_destroy,omitempty"`
	IgnoreChanges       []Value      `json:"ignore_changes,omitempty"`
	ReplaceTriggeredBy  []Reference  `json:"replace_triggered_by,omitempty"`
	Preconditions       []CheckRule  `json:"preconditions,omitempty"`
	Postconditions      []CheckRule  `json:"postconditions,omitempty"`
	Diagnostics         []Diagnostic `json:"diagnostics,omitempty"`
	Range               *SourceRange `json:"range,omitempty"`
}

type MovedBlock struct {
	From  string       `json:"from"`
	To    string       `json:"to"`
	Range *SourceRange `json:"range,omitempty"`
}

type ImportBlock struct {
	To       string       `json:"to"`
	ID       *Value       `json:"id,omitempty"`
	Provider *ProviderRef `json:"provider,omitempty"`
	Range    *SourceRange `json:"range,omitempty"`
}

type RemovedBlock struct {
	From      string       `json:"from"`
	Lifecycle *Lifecycle   `json:"lifecycle,omitempty"`
	Range     *SourceRange `json:"range,omitempty"`
}

type CheckBlock struct {
	Name       string       `json:"name"`
	Assertions []CheckRule  `json:"assertions,omitempty"`
	Range      *SourceRange `json:"range,omitempty"`
}

type CheckRule struct {
	Condition    *Value       `json:"condition,omitempty"`
	ErrorMessage *Value       `json:"error_message,omitempty"`
	References   []Reference  `json:"references,omitempty"`
	Range        *SourceRange `json:"range,omitempty"`
}

type TestFile struct {
	Path  string       `json:"path"`
	Runs  []TestRun    `json:"runs,omitempty"`
	Range *SourceRange `json:"range,omitempty"`
}

type TestRun struct {
	Name       string       `json:"name"`
	Command    string       `json:"command,omitempty"`
	Variables  []Attribute  `json:"variables,omitempty"`
	Assertions []CheckRule  `json:"assertions,omitempty"`
	Range      *SourceRange `json:"range,omitempty"`
}

type ValueKind string

const (
	ValueKindNull       ValueKind = "null"
	ValueKindBool       ValueKind = "bool"
	ValueKindNumber     ValueKind = "number"
	ValueKindString     ValueKind = "string"
	ValueKindCollection ValueKind = "collection"
	ValueKindExpression ValueKind = "expression"
	ValueKindUnknown    ValueKind = "unknown"
	ValueKindRedacted   ValueKind = "redacted"
)

// Value represents either a static literal, a symbolic expression, an unknown
// value, or a redacted likely-secret literal. Public JSON never emits Literal
// when SensitiveCandidate is true or Redacted is true.
type Value struct {
	Kind               ValueKind           `json:"kind"`
	Literal            any                 `json:"literal,omitempty"`
	Expression         string              `json:"expression,omitempty"`
	References         []Reference         `json:"references,omitempty"`
	Sensitive          bool                `json:"sensitive,omitempty"`
	SensitiveCandidate *SensitiveCandidate `json:"sensitive_candidate,omitempty"`
	Redacted           bool                `json:"redacted,omitempty"`
	UnknownReason      string              `json:"unknown_reason,omitempty"`
	Range              *SourceRange        `json:"range,omitempty"`
}

type SensitiveCandidate struct {
	Reason        string `json:"reason"`
	AttributePath string `json:"attribute_path,omitempty"`
}

type Reference struct {
	Traversal string       `json:"traversal"`
	Subject   string       `json:"subject,omitempty"`
	Range     *SourceRange `json:"range,omitempty"`
}

type DiagnosticSeverity string

const (
	DiagnosticError   DiagnosticSeverity = "error"
	DiagnosticWarning DiagnosticSeverity = "warning"
	DiagnosticInfo    DiagnosticSeverity = "info"
)

type Diagnostic struct {
	Severity      DiagnosticSeverity `json:"severity"`
	Code          string             `json:"code"`
	Summary       string             `json:"summary"`
	Detail        string             `json:"detail,omitempty"`
	ModuleAddress string             `json:"module_address,omitempty"`
	Address       string             `json:"address,omitempty"`
	Range         *SourceRange       `json:"range,omitempty"`
}
