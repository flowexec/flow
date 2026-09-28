package executable

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const (
	ExecutionGroupID = "execute"
	RetrievalGroupID = "retrieve"
	ViewGroupID      = "view"
	ConfigGroupID    = "config"
	UpdateGroupID    = "update"
	EvaluateGroupID  = "evaluate"

	CustomVerbPattern = `^[a-z][a-z0-9-]*$`
)

var (
	DefaultVerbAliasGroup = map[Verb]string{
		VerbExec:    ExecutionGroupID,
		VerbRun:     ExecutionGroupID,
		VerbExecute: ExecutionGroupID,

		VerbRetrieve: RetrievalGroupID,
		VerbFetch:    RetrievalGroupID,
		VerbGet:      RetrievalGroupID,

		VerbView: ViewGroupID,
		VerbShow: ViewGroupID,
		VerbList: ViewGroupID,

		VerbSetup:     ConfigGroupID,
		VerbConfigure: ConfigGroupID,

		VerbUpdate:  UpdateGroupID,
		VerbUpgrade: UpdateGroupID,

		VerbEvaluate: EvaluateGroupID,
		VerbEval:     EvaluateGroupID,
	}

	customVerbs      []Verb
	customVerbRegexp = regexp.MustCompile(CustomVerbPattern)
)

func ValidVerbs() []Verb {
	return append(builtinVerbs(), customVerbs...)
}

func RegisterCustomVerbs(verbs ...string) {
	registered := make([]Verb, 0, len(verbs))
	for _, v := range verbs {
		if ValidateCustomVerbName(v) != nil || slices.Contains(registered, Verb(v)) {
			continue
		}
		registered = append(registered, Verb(v))
	}
	customVerbs = registered
}

func IsBuiltinVerb(v Verb) bool {
	return slices.Contains(builtinVerbs(), v)
}

func ValidateCustomVerbName(name string) error {
	if !customVerbRegexp.MatchString(name) {
		return fmt.Errorf(
			"invalid verb %q: must contain only lowercase letters, numbers, and hyphens",
			name,
		)
	}
	if IsBuiltinVerb(Verb(name)) {
		return fmt.Errorf("verb %q is already a built-in verb", name)
	}
	return nil
}

//nolint:funlen
func builtinVerbs() []Verb {
	return []Verb{
		VerbAbort,
		VerbActivate,
		VerbAdd,
		VerbAnalyze,
		VerbApply,
		VerbArchive,
		VerbAudit,
		VerbBackup,
		VerbBenchmark,
		VerbBuild,
		VerbBundle,
		VerbCheck,
		VerbClean,
		VerbClear,
		VerbCommit,
		VerbCompile,
		VerbCompress,
		VerbConfigure,
		VerbConnect,
		VerbCopy,
		VerbCreate,
		VerbDeactivate,
		VerbDebug,
		VerbDecompress,
		VerbDecrypt,
		VerbDelete,
		VerbDeploy,
		VerbDestroy,
		VerbDiff,
		VerbDisable,
		VerbDisconnect,
		VerbDownload,
		VerbEdit,
		VerbEnable,
		VerbEncrypt,
		VerbErase,
		VerbEval,
		VerbEvaluate,
		VerbExec,
		VerbExecute,
		VerbExport,
		VerbExpose,
		VerbFetch,
		VerbFix,
		VerbFlush,
		VerbFormat,
		VerbGenerate,
		VerbGet,
		VerbImport,
		VerbIndex,
		VerbInfer,
		VerbInit,
		VerbInspect,
		VerbInstall,
		VerbJoin,
		VerbKill,
		VerbLaunch,
		VerbLint,
		VerbList,
		VerbLoad,
		VerbLock,
		VerbLogin,
		VerbLogout,
		VerbManage,
		VerbMerge,
		VerbMigrate,
		VerbModify,
		VerbMonitor,
		VerbMount,
		VerbNew,
		VerbNotify,
		VerbOpen,
		VerbPackage,
		VerbPartition,
		VerbPatch,
		VerbPause,
		VerbPing,
		VerbPlan,
		VerbPrefetch,
		VerbPreload,
		VerbProfile,
		VerbProvision,
		VerbPublish,
		VerbPull,
		VerbPurge,
		VerbPush,
		VerbQueue,
		VerbReboot,
		VerbRecover,
		VerbRefresh,
		VerbRelease,
		VerbReload,
		VerbRemove,
		VerbRender,
		VerbRequest,
		VerbReset,
		VerbRestart,
		VerbRestore,
		VerbResume,
		VerbRetrieve,
		VerbReview,
		VerbRollback,
		VerbRun,
		VerbSave,
		VerbScale,
		VerbScan,
		VerbSchedule,
		VerbSeed,
		VerbSend,
		VerbServe,
		VerbSet,
		VerbSetup,
		VerbShow,
		VerbSnapshot,
		VerbStart,
		VerbStash,
		VerbStop,
		VerbSummarize,
		VerbTag,
		VerbTeardown,
		VerbTerminate,
		VerbTest,
		VerbTidy,
		VerbTrace,
		VerbTrain,
		VerbTransform,
		VerbTrigger,
		VerbTune,
		VerbTunnel,
		VerbUndeploy,
		VerbUninstall,
		VerbUnlock,
		VerbUnmount,
		VerbUnset,
		VerbUpdate,
		VerbUpgrade,
		VerbUpload,
		VerbValidate,
		VerbVerify,
		VerbView,
		VerbWatch,
	}
}

func SortedValidVerbs() []string {
	verbs := make([]string, 0)
	for _, v := range ValidVerbs() {
		verbs = append(verbs, v.String())
	}
	slices.Sort(verbs)
	return verbs
}

func (v Verb) String() string {
	return string(v)
}

func (v Verb) Validate() error {
	if !slices.Contains(ValidVerbs(), v) {
		return fmt.Errorf("invalid executable verb %s", v)
	}
	return nil
}

func (v Verb) Equals(other Verb) bool {
	return v.String() == other.String() ||
		(DefaultVerbAliasGroup[v] != "" && DefaultVerbAliasGroup[v] == DefaultVerbAliasGroup[other])
}

func RelatedVerbs(verb Verb) []Verb {
	verbs := make([]Verb, 0)
	for _, v := range ValidVerbs() {
		if DefaultVerbAliasGroup[v] != "" && DefaultVerbAliasGroup[v] == DefaultVerbAliasGroup[verb] {
			verbs = append(verbs, v)
		}
	}
	return slices.Compact(verbs)
}

func NewRef(id string, verb Verb) Ref {
	if verb == "" {
		return ""
	}
	if id == "" {
		return Ref(verb.String())
	}
	return Ref(fmt.Sprintf("%s %s", verb, id))
}

func (r Ref) String() string {
	return string(r)
}

func (r Ref) Verb() Verb {
	refParts := strings.Split(string(r), " ")
	return Verb(refParts[0])
}

func (r Ref) ID() string {
	refParts := strings.Split(string(r), " ")
	if len(refParts) == 2 {
		return refParts[1]
	}
	return ""
}

func (r Ref) Namespace() string {
	id := r.ID()
	_, ns, _ := MustParseExecutableID(id)
	return ns
}

func (r Ref) Workspace() string {
	id := r.ID()
	ws, _, _ := MustParseExecutableID(id)
	return ws
}

func (r Ref) Equals(other Ref) bool {
	rVerb := r.Verb()
	oVerb := other.Verb()
	if !rVerb.Equals(oVerb) {
		return false
	}

	return r.ID() == other.ID()
}
