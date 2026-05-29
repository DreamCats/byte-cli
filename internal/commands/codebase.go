package commands

import (
	"flag"
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/DreamCats/byte-cli/internal/httpclient"
)

const (
	repoAPIBase = "https://codebase-api.byted.org/unstable"
	mrAPIBase   = "https://code.byted.org/api/v2/"
)

type Department struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	EnName    string `json:"en_name"`
	TenantKey string `json:"tenant_key"`
}

type Repo struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Platform    string     `json:"platform"`
	ExternalID  int        `json:"external_id"`
	ExternalURL string     `json:"external_url"`
	GitURL      string     `json:"git_url"`
	GitSSHURL   string     `json:"git_ssh_url"`
	GitHTTPURL  string     `json:"git_http_url"`
	Type        string     `json:"type"`
	Level       string     `json:"level"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	IsMonorepo  bool       `json:"is_monorepo"`
	MergeMethod string     `json:"merge_method"`
	Squash      string     `json:"squash"`
	Department  Department `json:"department"`
	CreatedAt   string     `json:"created_at"`
	AuditStatus string     `json:"audit_status"`
}

type DisplayName struct {
	Content string `json:"content"`
	I18n    string `json:"i18n"`
}

type User struct {
	ID          int         `json:"id"`
	Username    string      `json:"username"`
	DisplayName DisplayName `json:"display_name"`
	Email       string      `json:"email"`
}

type MR struct {
	ID              int    `json:"id"`
	Number          int    `json:"number"`
	Status          string `json:"status"`
	SourceBranch    string `json:"source_branch_name"`
	TargetBranch    string `json:"target_branch_name"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	CreatedBy       User   `json:"created_by"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	ChangesCount    int    `json:"changes_count"`
	CommitsCount    int    `json:"commits_count"`
	MergeMethod     string `json:"merge_method"`
	Draft           bool   `json:"draft"`
	SquashCommits   bool   `json:"squash_commits"`
	AutoMerge       bool   `json:"auto_merge"`
	MergeInProgress bool   `json:"merge_in_progress"`
}

type Position struct {
	Type      string `json:"type"`
	Side      string `json:"side"`
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type Suggestion struct {
	Content         string `json:"content"`
	OriginalContent string `json:"original_content"`
	Applied         bool   `json:"applied"`
}

type Comment struct {
	ID          int          `json:"id"`
	Content     string       `json:"content"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
	CreatedBy   User         `json:"created_by"`
	Suggestions []Suggestion `json:"suggestions"`
}

type Thread struct {
	ID        int        `json:"id"`
	Status    string     `json:"status"`
	Outdated  bool       `json:"outdated"`
	Comments  []Comment  `json:"comments"`
	Positions []Position `json:"positions"`
}

func Codebase(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printHelp(out, `Codebase 仓库和 MR 查询

Usage:
  byte-cli codebase <command>

Commands:
  repo       仓库操作
  mr         MR 操作`)
		return 0, nil
	}
	switch args[0] {
	case "repo":
		return codebaseRepo(args[1:], out)
	case "mr":
		return codebaseMR(args[1:], out)
	default:
		return 1, fmt.Errorf("unknown codebase command: %s", args[0])
	}
}

func codebaseRepo(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printHelp(out, `仓库操作

Usage:
  byte-cli codebase repo <command>

Commands:
  info <repo-name>   查看仓库信息，repo-name 格式: group/project`)
		return 0, nil
	}
	if args[0] != "info" {
		return 1, fmt.Errorf("codebase repo requires info")
	}
	if wantsHelp(args[1:]) {
		printHelp(out, `查看仓库信息

Usage:
  byte-cli codebase repo info [options] <repo-name>

Options:
  --json       JSON 格式输出
  -h, --help   显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("codebase repo info", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	if err := fs.Parse(args[1:]); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("repo info requires REPO_NAME")
	}
	info, err := getRepoInfo(fs.Arg(0))
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, info)
	}
	fmt.Fprintf(out.Out, "仓库: %s\n", info.Name)
	fmt.Fprintf(out.Out, "ID: %d\n", info.ID)
	fmt.Fprintf(out.Out, "描述: %s\n", defaultString(info.Description, "(无)"))
	fmt.Fprintf(out.Out, "类型: %s\n", info.Type)
	fmt.Fprintf(out.Out, "状态: %s\n", info.Status)
	fmt.Fprintf(out.Out, "Git SSH: %s\n", info.GitSSHURL)
	fmt.Fprintf(out.Out, "Git HTTP: %s\n", info.GitHTTPURL)
	fmt.Fprintf(out.Out, "合并方式: %s\n", info.MergeMethod)
	fmt.Fprintf(out.Out, "创建时间: %s\n", info.CreatedAt)
	return 0, nil
}

func codebaseMR(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printHelp(out, `MR 操作

Usage:
  byte-cli codebase mr <command>

Commands:
  get <number>        查看 MR 详情
  comments <number>   查看 MR 评论`)
		return 0, nil
	}
	switch args[0] {
	case "get":
		return codebaseMRGet(args[1:], out)
	case "comments":
		return codebaseMRComments(args[1:], out)
	default:
		return 1, fmt.Errorf("unknown codebase mr command: %s", args[0])
	}
}

func codebaseMRGet(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查看 MR 详情

Usage:
  byte-cli codebase mr get [options] <number>

Options:
  -R, --repo <repo>   仓库名 group/project，默认从 git remote 推断
  --json              JSON 格式输出
  -h, --help          显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("codebase mr get", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	repo := fs.String("repo", "", "仓库名 group/project")
	fs.StringVar(repo, "R", "", "仓库名 group/project")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("mr get requires NUMBER")
	}
	repoName, err := resolveRepo(*repo, out)
	if err != nil {
		return 1, err
	}
	info, err := getRepoInfo(repoName)
	if err != nil {
		return authExitCode(err), err
	}
	mr, err := getMR(info.ID, atoiDefault(fs.Arg(0), 0))
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, mr)
	}
	fmt.Fprintf(out.Out, "MR !%d: %s\n", mr.Number, mr.Title)
	fmt.Fprintf(out.Out, "状态: %s\n", mr.Status)
	fmt.Fprintf(out.Out, "作者: %s (@%s)\n", mr.CreatedBy.DisplayName.Content, mr.CreatedBy.Username)
	fmt.Fprintf(out.Out, "源分支: %s → %s\n", mr.SourceBranch, mr.TargetBranch)
	fmt.Fprintf(out.Out, "变更: %d 文件, %d 提交\n", mr.ChangesCount, mr.CommitsCount)
	if mr.Draft {
		fmt.Fprintln(out.Out, "Draft: 是")
	}
	fmt.Fprintf(out.Out, "创建时间: %s\n", mr.CreatedAt)
	if mr.Description != "" {
		fmt.Fprintf(out.Out, "\n描述:\n%s\n", mr.Description)
	}
	return 0, nil
}

func codebaseMRComments(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查看 MR 评论

Usage:
  byte-cli codebase mr comments [options] <number>

Options:
  -R, --repo <repo>   仓库名 group/project，默认从 git remote 推断
  --unresolved        只显示未解决的评论
  --json              JSON 格式输出
  -h, --help          显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("codebase mr comments", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	repo := fs.String("repo", "", "仓库名 group/project")
	fs.StringVar(repo, "R", "", "仓库名 group/project")
	unresolved := fs.Bool("unresolved", false, "只显示未解决的评论")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("mr comments requires NUMBER")
	}
	repoName, err := resolveRepo(*repo, out)
	if err != nil {
		return 1, err
	}
	info, err := getRepoInfo(repoName)
	if err != nil {
		return authExitCode(err), err
	}
	mr, err := getMR(info.ID, atoiDefault(fs.Arg(0), 0))
	if err != nil {
		return authExitCode(err), err
	}
	threads, err := getMRThreads(info.ID, mr.ID)
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, threads)
	}
	for _, thread := range threads {
		if *unresolved && thread.Status == "resolved" {
			continue
		}
		outdated := ""
		if thread.Outdated {
			outdated = " [outdated]"
		}
		if len(thread.Positions) > 0 {
			pos := thread.Positions[0]
			fmt.Fprintf(out.Out, "\n● %s:%d%s\n", pos.Path, pos.StartLine, outdated)
		} else {
			fmt.Fprintf(out.Out, "\n● (general)%s\n", outdated)
		}
		for _, comment := range thread.Comments {
			printComment(comment, out)
		}
	}
	return 0, nil
}

func getRepoInfo(repo string) (Repo, error) {
	token, err := tokenFor("codebase")
	if err != nil {
		return Repo{}, err
	}
	resp, err := httpclient.Get(repoAPIBase+"/repos/"+url.PathEscape(repo), map[string]string{
		"Authorization":   "Codebase-User-JWT " + token,
		"Accept-Encoding": "gzip, deflate, br, zstd",
	})
	if err != nil {
		return Repo{}, err
	}
	var info Repo
	return info, readJSON(resp, &info)
}

func getMR(repoID, number int) (MR, error) {
	token, err := tokenFor("codebase")
	if err != nil {
		return MR{}, err
	}
	body := map[string]any{
		"RepoId": repoID,
		"Number": number,
		"Selector": map[string]bool{
			"Labels": true, "CurrentUser": true, "Version": true, "User": true,
		},
	}
	resp, err := httpclient.Post(mrAPIBase+"?Action=GetMergeRequest", body, map[string]string{"x-codebase-user-jwt": token})
	if err != nil {
		return MR{}, err
	}
	var mr MR
	return mr, readJSON(resp, &mr)
}

func getMRThreads(repoID, mrID int) ([]Thread, error) {
	token, err := tokenFor("codebase")
	if err != nil {
		return nil, err
	}
	body := map[string]any{"RepoId": repoID, "CommentableId": mrID, "CommentableType": "merge_request"}
	resp, err := httpclient.Post(mrAPIBase+"?Action=ListThreads", body, map[string]string{"x-codebase-user-jwt": token})
	if err != nil {
		return nil, err
	}
	var payload struct {
		Threads []Thread `json:"threads"`
	}
	if err := readJSON(resp, &payload); err != nil {
		return nil, err
	}
	return payload.Threads, nil
}

func resolveRepo(repo string, out Output) (string, error) {
	if repo != "" {
		return repo, nil
	}
	inferred := InferRepoName()
	if inferred == "" {
		return "", fmt.Errorf("无法推断仓库名，请在 git 仓库中运行或使用 -R 指定")
	}
	fmt.Fprintf(out.Out, "推断仓库: %s\n", inferred)
	return inferred, nil
}

func InferRepoName() string {
	raw, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return ParseRepoName(strings.TrimSpace(string(raw)))
}

func ParseRepoName(remoteURL string) string {
	if strings.HasPrefix(remoteURL, "ssh://") || strings.HasPrefix(remoteURL, "http://") || strings.HasPrefix(remoteURL, "https://") {
		parsed, err := url.Parse(remoteURL)
		if err != nil {
			return ""
		}
		return cleanRepoPath(parsed.Path)
	}
	if strings.Contains(remoteURL, "@") && strings.Contains(remoteURL, ":") {
		idx := strings.LastIndex(remoteURL, ":")
		return cleanRepoPath(remoteURL[idx+1:])
	}
	return ""
}

func cleanRepoPath(path string) string {
	path = strings.Trim(path, "/")
	path = strings.TrimSuffix(path, ".git")
	if !strings.Contains(path, "/") {
		return ""
	}
	return path
}

func printComment(comment Comment, out Output) {
	author := comment.CreatedBy.DisplayName.Content
	if author == "" {
		author = comment.CreatedBy.Username
	}
	header := fmt.Sprintf("  [%s] %s", author, comment.CreatedAt)
	if comment.CreatedBy.Username == "aime" {
		printAimeComment(comment, header, out)
		return
	}
	fmt.Fprintln(out.Out, header)
	preview := comment.Content
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	fmt.Fprintf(out.Out, "  %s\n\n", preview)
}

func printAimeComment(comment Comment, header string, out Output) {
	fmt.Fprintln(out.Out, header)
	description := extractMarkdownField(comment.Content, "**问题描述**:")
	priority := extractMarkdownField(comment.Content, "**优先级**:")
	category := extractMarkdownField(comment.Content, "**问题分类**:")
	if description != "" {
		fmt.Fprintf(out.Out, "  问题: %s\n", description)
	}
	if priority != "" {
		fmt.Fprintf(out.Out, "  优先级: %s\n", priority)
	}
	if category != "" {
		fmt.Fprintf(out.Out, "  分类: %s\n", category)
	}
	fmt.Fprintln(out.Out)
}

func extractMarkdownField(content, prefix string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	return ""
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func atoiDefault(value string, fallback int) int {
	var n int
	_, err := fmt.Sscanf(value, "%d", &n)
	if err != nil {
		return fallback
	}
	return n
}
