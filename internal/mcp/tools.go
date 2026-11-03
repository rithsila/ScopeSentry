package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Autumn-27/ScopeSentry/internal/constants"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/app"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/asset"
	assetCommon "github.com/Autumn-27/ScopeSentry/internal/services/assets/common"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/crawler"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/dirscan"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/ip"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/mp"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/page_monitoring"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/root_domain"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/sensitive"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/subdomain"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/url"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/vulnerability"
	"github.com/Autumn-27/ScopeSentry/internal/services/node"
	"github.com/Autumn-27/ScopeSentry/internal/services/plugin"
	"github.com/Autumn-27/ScopeSentry/internal/services/project"
	taskCommon "github.com/Autumn-27/ScopeSentry/internal/services/task/common"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/task"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/template"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type deps struct {
	projectService    project.Service
	taskService       task.Service
	taskCommonService taskCommon.Service
	templateService   template.Service
	assetService      asset.Service
	rootDomainService root_domain.Service
	subdomainService  subdomain.Service
	appService        app.Service
	mpService         mp.Service
	urlService        url.Service
	sensitiveService  sensitive.Service
	dirscanService    dirscan.Service
	crawlerService    crawler.Service
	vulnService       vulnerability.Service
	ipService         ip.Service
	pageMonService    page_monitoring.Service
	commonService     assetCommon.Service
	nodeService       node.Service
	pluginService     plugin.Service
}

var (
	depsInstance *deps
	depsOnce     sync.Once
)

func getDeps() *deps {
	depsOnce.Do(func() {
		depsInstance = &deps{
			projectService:    project.NewService(),
			taskService:       task.NewService(),
			taskCommonService: taskCommon.NewService(),
			templateService:   template.NewService(),
			assetService:      asset.NewService(),
			rootDomainService: root_domain.NewService(),
			subdomainService:  subdomain.NewService(),
			appService:        app.NewService(),
			mpService:         mp.NewService(),
			urlService:        url.NewService(),
			sensitiveService:  sensitive.NewService(),
			dirscanService:    dirscan.NewService(),
			crawlerService:    crawler.NewService(),
			vulnService:       vulnerability.NewService(),
			ipService:         ip.NewService(),
			pageMonService:    page_monitoring.NewService(),
			commonService:     assetCommon.NewService(),
			nodeService:       node.NewService(),
			pluginService:     plugin.NewService(),
		}
	})
	return depsInstance
}

func registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List projects grouped by tag",
	}, listProjects)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects_data",
		Description: "Paginated list of projects with search support",
	}, listProjectsData)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project",
		Description: "Get project details by project ID",
	}, getProject)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_project",
		Description: "Create a new project. Requires name and valid target; optional tag, template, node, etc. If project name exists and cleaned target has targets, replaces project scope.",
	}, createProject)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tasks",
		Description: "Paginated list of scan tasks",
	}, listTasks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Description: "Get scan task details by task ID",
	}, getTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "stop_scan_task",
		Description: "Request stopping one scan task by ObjectID. Success means service acceptance, not scanner acknowledgement.",
	}, stopScanTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "resume_scan_task",
		Description: "Request resuming one stopped scan task by ObjectID. Success means service acceptance, not scanner acknowledgement.",
	}, resumeScanTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_scan_templates",
		Description: "Paginated list of scan templates",
	}, listScanTemplates)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_scan_template",
		Description: "Get scan template details by template ID",
	}, getScanTemplate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_plugin_modules",
		Description: "List all scan template module names (pipeline stages). Scan templates consist of these modules, each mounting plugins.",
	}, listPluginModules)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_plugins",
		Description: "List available scan plugins, filterable by module. Returns plugin hash, name, module, default parameter. When creating templates, module fields reference plugin hashes.",
	}, listPlugins)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_scan_template",
		Description: createScanTemplateToolDesc,
	}, createScanTemplate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_scan_task",
		Description: createScanTaskToolDesc,
	}, createScanTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_assets",
		Description: listAssetsToolDesc,
	}, listAssets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "count_assets",
		Description: countAssetsToolDesc,
	}, countAssets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_asset_detail",
		Description: "Get asset details; asset_type supports asset or vulnerability",
	}, getAssetDetail)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_asset_tag",
		Description: "Add a tag to an asset",
	}, addAssetTag)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nodes",
		Description: "List scan nodes; online_only=true returns only online nodes",
	}, listNodes)
}

type listProjectsDataInput struct {
	Search    string `json:"search,omitempty" jsonschema:"Keyword search on project name"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number starting from 1, default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Number of items per page, default 20"`
}

func listProjects(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	c := ginContext(ctx)
	result, err := getDeps().projectService.GetProjectsByTag(c)
	if err != nil {
		return errorResult("failed to list projects", err)
	}
	return jsonToolResult(ginH{"list": result})
}

func listProjectsData(ctx context.Context, _ *mcp.CallToolRequest, input listProjectsDataInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	c := ginContext(ctx)
	result, err := getDeps().projectService.GetProjectsData(c, input.Search, input.PageIndex, input.PageSize)
	if err != nil {
		return errorResult("failed to get projects data", err)
	}
	return jsonToolResult(result)
}

type getProjectInput struct {
	ID string `json:"id" jsonschema:"Project MongoDB ObjectID"`
}

func getProject(ctx context.Context, _ *mcp.CallToolRequest, input getProjectInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id cannot be empty", nil)
	}
	c := ginContext(ctx)
	result, err := getDeps().projectService.GetProjectContent(c, input.ID)
	if err != nil {
		return errorResult("failed to get project details", err)
	}
	if result == nil {
		return errorResult("project not found", nil)
	}
	return jsonToolResult(result)
}

type createProjectInput struct {
	Name           string   `json:"name" jsonschema:"Project name, required"`
	Tag            string   `json:"tag,omitempty" jsonschema:"Project tag for grouping"`
	Target         string   `json:"target" jsonschema:"Scan targets, required. Domains/IPs/CIDRs/URLs separated by actual newline characters, not commas or literal backslash-n"`
	Template       string   `json:"template,omitempty" jsonschema:"Associated scan template ObjectID"`
	Node           []string `json:"node,omitempty" jsonschema:"List of designated scanning node names"`
	AllNode        bool     `json:"allNode,omitempty" jsonschema:"Whether to use all available nodes"`
	Ignore         string   `json:"ignore,omitempty" jsonschema:"Targets to ignore, format same as target"`
	Duplicates     string   `json:"duplicates,omitempty" jsonschema:"Deduplication policy"`
	ScheduledTasks bool     `json:"scheduledTasks,omitempty" jsonschema:"Whether to enable recurring scheduled scan"`
	Hour           int      `json:"hour,omitempty" jsonschema:"Interval in hours for scheduled scan, valid when scheduledTasks is true"`
}

func createProject(ctx context.Context, _ *mcp.CallToolRequest, input createProjectInput) (*mcp.CallToolResult, any, error) {
	if input.Name == "" || strings.TrimSpace(input.Target) == "" {
		return errorResult("name and target cannot be empty", nil)
	}
	p := &models.Project{
		Name:           input.Name,
		Tag:            input.Tag,
		Target:         input.Target,
		Template:       input.Template,
		Node:           input.Node,
		AllNode:        input.AllNode,
		Ignore:         input.Ignore,
		Duplicates:     input.Duplicates,
		ScheduledTasks: input.ScheduledTasks,
		Hour:           input.Hour,
		Tp:             "project",
	}
	c := ginContext(ctx)
	id, replaced, err := getDeps().projectService.ApplyProjectScope(c, p)
	if err != nil {
		return errorResult("failed to create project", err)
	}
	if replaced {
		return jsonToolResult(ginH{
			"success":  true,
			"message":  "project scope replaced",
			"id":       id,
			"replaced": true,
		})
	}
	return jsonToolResult(ginH{
		"success":  true,
		"message":  "project created successfully",
		"id":       id,
		"replaced": false,
	})
}

type listTasksInput struct {
	Search    string `json:"search,omitempty" jsonschema:"Keyword search on task name"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number starting from 1, default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Number of items per page, default 20"`
}

func listTasks(ctx context.Context, _ *mcp.CallToolRequest, input listTasksInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	c := ginContext(ctx)
	tasks, total, err := getDeps().taskService.List(c, input.Search, input.PageIndex, input.PageSize)
	if err != nil {
		return errorResult("failed to list tasks", err)
	}
	return jsonToolResult(ginH{"list": tasks, "total": total})
}

type getTaskInput struct {
	ID string `json:"id" jsonschema:"Task MongoDB ObjectID"`
}

func getTask(ctx context.Context, _ *mcp.CallToolRequest, input getTaskInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id cannot be empty", nil)
	}
	c := ginContext(ctx)
	result, err := getDeps().taskService.GetTaskDetail(c, input.ID)
	if err != nil {
		return errorResult("failed to get task details", err)
	}
	if result == nil {
		return errorResult("task not found", nil)
	}
	return jsonToolResult(result)
}

type listScanTemplatesInput struct {
	Query     string `json:"query,omitempty" jsonschema:"Keyword search on template name"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number starting from 1, default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Number of items per page, default 20"`
}

func listScanTemplates(ctx context.Context, _ *mcp.CallToolRequest, input listScanTemplatesInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	result, err := getDeps().templateService.List(ctx, input.PageIndex, input.PageSize, input.Query)
	if err != nil {
		return errorResult("failed to list scan templates", err)
	}
	return jsonToolResult(result)
}

type getScanTemplateInput struct {
	ID string `json:"id" jsonschema:"Scan template MongoDB ObjectID"`
}

func getScanTemplate(ctx context.Context, _ *mcp.CallToolRequest, input getScanTemplateInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id cannot be empty", nil)
	}
	c := ginContext(ctx)
	result, err := getDeps().templateService.Detail(c, input.ID)
	if err != nil {
		return errorResult("failed to get scan template details", err)
	}
	return jsonToolResult(result)
}

type createScanTemplateInput struct {
	Name         string                       `json:"name" jsonschema:"Template name, required"`
	Modules      map[string][]string          `json:"modules,omitempty" jsonschema:"Mapping of module name to plugin hash array (e.g. {\"SubdomainScan\":[\"d60ba73c...\"]})"`
	Parameters   map[string]map[string]string `json:"parameters,omitempty" jsonschema:"Optional parameter overrides: {module: {pluginHash: paramString}}"`
	VulList      []string                     `json:"vullist,omitempty" jsonschema:"Optional Nuclei POC template IDs for VulnerabilityScan"`
	TemplateJSON string                       `json:"template_json,omitempty" jsonschema:"Optional complete ScanTemplate JSON, takes precedence over modules"`
}

func createScanTemplate(ctx context.Context, _ *mcp.CallToolRequest, input createScanTemplateInput) (*mcp.CallToolResult, any, error) {
	if input.Name == "" && input.TemplateJSON == "" {
		return errorResult("name or template_json must be provided", nil)
	}

	var tmpl *models.ScanTemplate

	switch {
	case input.TemplateJSON != "":
		tmpl = &models.ScanTemplate{}
		if err := json.Unmarshal([]byte(input.TemplateJSON), tmpl); err != nil {
			return errorResult("invalid template_json format", err)
		}
		if tmpl.Name == "" {
			tmpl.Name = input.Name
		}
	case len(input.Modules) > 0:
		built, err := buildTemplateFromModules(ctx, input)
		if err != nil {
			return errorResult(err.Error(), nil)
		}
		tmpl = built
	default:
		tmpl = &models.ScanTemplate{Name: input.Name}
	}

	if tmpl.Name == "" {
		return errorResult("template name cannot be empty", nil)
	}

	id, err := getDeps().templateService.Save(ctx, "", tmpl)
	if err != nil {
		return errorResult("failed to create scan template", err)
	}
	return jsonToolResult(ginH{"success": true, "id": id, "message": "template created successfully, can be used for create_scan_task template parameter"})
}

// buildTemplateFromModules assembles a scan template from modules and plugin hashes, populating default parameters.
func buildTemplateFromModules(ctx context.Context, input createScanTemplateInput) (*models.ScanTemplate, error) {
	validModule := make(map[string]bool, len(constants.PLUGINSMODULES))
	for _, m := range constants.PLUGINSMODULES {
		validModule[m] = true
	}

	templateMap := map[string]any{
		"name": input.Name,
	}
	paramMap := map[string]map[string]string{}

	for module, hashes := range input.Modules {
		if !validModule[module] {
			return nil, fmt.Errorf("invalid module name: %s (use list_plugin_modules for valid modules)", module)
		}
		if len(hashes) == 0 {
			continue
		}
		templateMap[module] = hashes

		modParams := map[string]string{}
		for _, hash := range hashes {
			if input.Parameters != nil {
				if mp, ok := input.Parameters[module]; ok {
					if v, ok := mp[hash]; ok {
						modParams[hash] = v
						continue
					}
				}
			}
			plg, err := getDeps().pluginService.GetPluginByHash(ctx, hash)
			if err != nil || plg == nil {
				return nil, fmt.Errorf("plugin hash not found: %s (module %s)", hash, module)
			}
			if plg.Module != module {
				return nil, fmt.Errorf("plugin %s (hash=%s) belongs to module %s, cannot place in module %s", plg.Name, hash, plg.Module, module)
			}
			modParams[hash] = plg.Parameter
		}
		if len(modParams) > 0 {
			paramMap[module] = modParams
		}
	}

	if len(paramMap) > 0 {
		templateMap["Parameters"] = paramMap
	}
	if len(input.VulList) > 0 {
		templateMap["vullist"] = input.VulList
	}

	data, err := json.Marshal(templateMap)
	if err != nil {
		return nil, fmt.Errorf("failed to build template: %w", err)
	}
	tmpl := &models.ScanTemplate{}
	if err := json.Unmarshal(data, tmpl); err != nil {
		return nil, fmt.Errorf("failed to build template: %w", err)
	}
	return tmpl, nil
}

type createScanTaskInput struct {
	Name           string              `json:"name" jsonschema:"Task name, required and unique"`
	Target         string              `json:"target,omitempty" jsonschema:"Scan targets separated by actual newline characters, required when targetSource is general; not commas or literal backslash-n"`
	Node           []string            `json:"node,omitempty" jsonschema:"Online scanning node names, required unless allNode is true"`
	Template       string              `json:"template,omitempty" jsonschema:"Scan template ObjectID, required to execute scan"`
	AllNode        bool                `json:"allNode,omitempty" jsonschema:"Whether to automatically include all online nodes"`
	Ignore         string              `json:"ignore,omitempty" jsonschema:"Targets to ignore, format same as target"`
	Duplicates     string              `json:"duplicates,omitempty" jsonschema:"Deduplication policy"`
	Project        []string            `json:"project,omitempty" jsonschema:"Project ObjectIDs, required when targetSource is project"`
	TargetSource   string              `json:"targetSource,omitempty" jsonschema:"Target source: general, project, asset, RootDomain, subdomain, UrlScan, etc. Default general"`
	TargetTp       string              `json:"targetTp,omitempty" jsonschema:"Selection type for asset source: search or select"`
	Search         string              `json:"search,omitempty" jsonschema:"Search DSL expression when filtering targets from asset database"`
	Filter         map[string][]string `json:"filter,omitempty" jsonschema:"Exact filter map, filter.project contains project ObjectIDs"`
	TargetNumber   int                 `json:"targetNumber,omitempty" jsonschema:"Max target count limit in search mode, 0 for unlimited"`
	TargetIds      []string            `json:"targetIds,omitempty" jsonschema:"Selected asset ObjectIDs in select mode"`
	BindProject    string              `json:"bindProject,omitempty" jsonschema:"Associated project ObjectID to bind findings to"`
	ScheduledTasks bool                `json:"scheduledTasks,omitempty" jsonschema:"Unsupported: true is rejected; MCP supports one-shot tasks only"`
	Hour           int                 `json:"hour,omitempty" jsonschema:"Scheduled interval - hours"`
	Minute         int                 `json:"minute,omitempty" jsonschema:"Scheduled interval - minutes"`
	Day            int                 `json:"day,omitempty" jsonschema:"Scheduled interval - days"`
	Week           int                 `json:"week,omitempty" jsonschema:"Scheduled interval - week (for weekly cycle)"`
	CycleType      string              `json:"cycleType,omitempty" jsonschema:"Cycle type: nhours, daily, weekly"`
}

func createScanTask(ctx context.Context, _ *mcp.CallToolRequest, input createScanTaskInput) (*mcp.CallToolResult, any, error) {
	if input.ScheduledTasks {
		return errorResult("scheduledTasks=true is unsupported; MCP supports one-shot scans only", nil)
	}
	if strings.TrimSpace(input.Name) == "" {
		return errorResult("name cannot be empty", nil)
	}
	if _, err := primitive.ObjectIDFromHex(input.Template); err != nil {
		return errorResult("template must be a valid scan template ObjectID; use list_scan_templates", nil)
	}
	c := ginContext(ctx)
	tmpl, err := getDeps().templateService.Detail(c, input.Template)
	if err != nil {
		return errorResult("failed to get scan template; use list_scan_templates", err)
	}
	if tmpl == nil {
		return errorResult("scan template not found; use list_scan_templates", nil)
	}
	onlineNodes, err := getDeps().nodeService.GetNodeData(ctx, true)
	if err != nil {
		return errorResult("failed to list online nodes", err)
	}
	onlineNames := make(map[string]bool, len(onlineNodes))
	for _, n := range onlineNodes {
		if strings.TrimSpace(n.Name) != "" {
			onlineNames[n.Name] = true
		}
	}
	for _, name := range input.Node {
		if strings.TrimSpace(name) == "" {
			return errorResult("node name cannot be empty", nil)
		}
		if !onlineNames[name] {
			return errorResult(fmt.Sprintf("node %q is offline or unknown; use list_nodes with online_only=true", name), nil)
		}
	}
	if input.AllNode {
		input.Node = make([]string, 0, len(onlineNodes))
		seen := make(map[string]bool, len(onlineNodes))
		for _, n := range onlineNodes {
			if onlineNames[n.Name] && !seen[n.Name] {
				input.Node = append(input.Node, n.Name)
				seen[n.Name] = true
			}
		}
	}
	if len(input.Node) == 0 {
		return errorResult("at least one online node is required; use list_nodes with online_only=true", nil)
	}
	exists, err := getDeps().taskService.CheckTaskNameExists(c, input.Name)
	if err != nil {
		return errorResult("failed to check task name", err)
	}
	if exists {
		return errorResult("task name already exists", nil)
	}

	targetSource := input.TargetSource
	if targetSource == "" {
		targetSource = "general"
	}

	taskModel := &models.Task{
		Name:     input.Name,
		Target:   input.Target,
		Node:     input.Node,
		Template: input.Template,
		// Nodes were resolved above. Do not let background dispatch add offline nodes.
		AllNode:        false,
		Ignore:         input.Ignore,
		Duplicates:     input.Duplicates,
		Project:        input.Project,
		TargetSource:   targetSource,
		TargetTp:       input.TargetTp,
		Search:         input.Search,
		TargetNumber:   input.TargetNumber,
		TargetIds:      input.TargetIds,
		BindProject:    input.BindProject,
		ScheduledTasks: input.ScheduledTasks,
		Hour:           input.Hour,
		Minute:         input.Minute,
		Day:            input.Day,
		Week:           input.Week,
		CycleType:      input.CycleType,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		taskModel.Filter = filter
	}
	taskID, err := getDeps().taskCommonService.Insert(ctx, taskModel)
	if err != nil {
		return errorResult("failed to create scan task", err)
	}
	return jsonToolResult(ginH{"success": true, "id": taskID})
}

type listAssetsInput struct {
	AssetType        string              `json:"asset_type" jsonschema:"Asset type, required (e.g. asset, RootDomain, subdomain, app, mp, UrlScan, SensitiveResult, DirScanResult, crawler, vulnerability, PageMonitoring, IPAsset, SubdomainTakerResult)"`
	PageIndex        int                 `json:"pageIndex,omitempty" jsonschema:"Page number starting from 1, default 1"`
	PageSize         int                 `json:"pageSize,omitempty" jsonschema:"Number of items per page, default 20"`
	SearchExpression string              `json:"search,omitempty" jsonschema:"Search DSL expression, see list_assets tool description"`
	Filter           map[string][]string `json:"filter,omitempty" jsonschema:"Exact filter JSON map. filter.project contains project ObjectIDs; filter.task contains task names"`
	Sort             map[string]string   `json:"sort,omitempty" jsonschema:"Sort map. Only UrlScan and DirScanResult support length: ascending or descending"`
	Sid              string              `json:"sid,omitempty" jsonschema:"Sensitive rule name (sid), only valid when asset_type is SensitiveResult"`
}

type countAssetsInput struct {
	AssetType        string              `json:"asset_type" jsonschema:"Asset type, required, same values as list_assets"`
	SearchExpression string              `json:"search,omitempty" jsonschema:"Search DSL expression, same syntax as list_assets"`
	Filter           map[string][]string `json:"filter,omitempty" jsonschema:"Exact filter JSON map, filter.project contains project ObjectIDs"`
	Sid              string              `json:"sid,omitempty" jsonschema:"Sensitive rule name (sid), only valid when asset_type is SensitiveResult"`
}

func countAssets(ctx context.Context, _ *mcp.CallToolRequest, input countAssetsInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}

	query := models.SearchRequest{
		Index:            index,
		SearchExpression: input.SearchExpression,
		Sid:              input.Sid,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		query.Filter = filter
	}

	total, err := getDeps().commonService.TotalData(ctx, &query)
	if err != nil {
		return errorResult("failed to count assets", err)
	}
	return jsonToolResult(ginH{"total": total})
}

func listAssets(ctx context.Context, _ *mcp.CallToolRequest, input listAssetsInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	query := models.SearchRequest{
		PageIndex:        input.PageIndex,
		PageSize:         input.PageSize,
		Index:            index,
		SearchExpression: input.SearchExpression,
		Sort:             input.Sort,
		Sid:              input.Sid,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		query.Filter = filter
	}

	c := ginContext(ctx)
	data, err := queryAssets(ctx, c, index, query)
	if err != nil {
		return errorResult("failed to query assets", err)
	}
	return jsonToolResult(data)
}

type getAssetDetailInput struct {
	AssetType string `json:"asset_type" jsonschema:"Asset type, supports asset or vulnerability"`
	ID        string `json:"id" jsonschema:"Asset MongoDB ObjectID; for vulnerability use hash"`
}

func getAssetDetail(ctx context.Context, _ *mcp.CallToolRequest, input getAssetDetailInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id cannot be empty", nil)
	}
	c := ginContext(ctx)
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}

	switch index {
	case "asset":
		result, err := getDeps().assetService.GetAssetByID(c, input.ID)
		if err != nil {
			return errorResult("failed to get asset details", err)
		}
		if result == nil {
			return errorResult("asset not found", nil)
		}
		return jsonToolResult(result)
	case "vulnerability":
		result, err := getDeps().vulnService.GetVulnerabilityDetailByHash(c, input.ID)
		if err != nil {
			return errorResult("failed to get vulnerability details", err)
		}
		return jsonToolResult(result)
	default:
		return errorResult("asset_type only supports asset or vulnerability for detail query", nil)
	}
}

type addAssetTagInput struct {
	AssetType string `json:"asset_type" jsonschema:"Asset type, same values as list_assets"`
	ID        string `json:"id" jsonschema:"Asset MongoDB ObjectID"`
	Tag       string `json:"tag" jsonschema:"Tag name to add"`
}

func addAssetTag(ctx context.Context, _ *mcp.CallToolRequest, input addAssetTagInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}
	if input.ID == "" || input.Tag == "" {
		return errorResult("id and tag cannot be empty", nil)
	}
	c := ginContext(ctx)
	req := &models.TagRequest{Type: index, ID: input.ID, Tag: input.Tag}
	if err := getDeps().commonService.AddTag(c, req); err != nil {
		return errorResult("failed to add tag", err)
	}
	return jsonToolResult(ginH{"success": true})
}

type listNodesInput struct {
	OnlineOnly bool `json:"online_only,omitempty" jsonschema:"When true, only returns online nodes. Default false returns all"`
}

func listNodes(ctx context.Context, _ *mcp.CallToolRequest, input listNodesInput) (*mcp.CallToolResult, any, error) {
	result, err := getDeps().nodeService.GetNodeData(ctx, input.OnlineOnly)
	if err != nil {
		return errorResult("failed to list nodes", err)
	}
	return jsonToolResult(ginH{"list": result})
}

func listPluginModules(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return jsonToolResult(ginH{"modules": constants.PLUGINSMODULES})
}

type listPluginsInput struct {
	Module string `json:"module,omitempty" jsonschema:"Filter by module name; empty returns all plugins. Module names from list_plugin_modules"`
	Search string `json:"search,omitempty" jsonschema:"Keyword search on plugin name (only applied when module is empty)"`
}

type pluginBrief struct {
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	Parameter string `json:"parameter"`
	Type      string `json:"type"`
	Status    bool   `json:"status"`
}

func toPluginBriefs(plugins []models.Plugin) []pluginBrief {
	briefs := make([]pluginBrief, 0, len(plugins))
	for _, p := range plugins {
		if p.Type == "server" {
			continue
		}
		briefs = append(briefs, pluginBrief{
			Hash:      p.Hash,
			Name:      p.Name,
			Module:    p.Module,
			Parameter: p.Parameter,
			Type:      p.Type,
			Status:    p.Status,
		})
	}
	return briefs
}

func listPlugins(ctx context.Context, _ *mcp.CallToolRequest, input listPluginsInput) (*mcp.CallToolResult, any, error) {
	c := ginContext(ctx)

	if input.Module != "" {
		plugins, err := getDeps().pluginService.ListByModule(c, input.Module)
		if err != nil {
			return errorResult("failed to list plugins", err)
		}
		return jsonToolResult(ginH{"list": toPluginBriefs(plugins)})
	}

	resp, err := getDeps().pluginService.List(c, &models.PluginListRequest{
		PageIndex: 1,
		PageSize:  500,
		Search:    input.Search,
	})
	if err != nil {
		return errorResult("failed to list plugins", err)
	}
	return jsonToolResult(ginH{"list": toPluginBriefs(resp.List), "total": resp.Total})
}

func queryAssets(ctx context.Context, c *gin.Context, index string, query models.SearchRequest) (any, error) {
	d := getDeps()
	switch index {
	case "asset":
		list, err := d.assetService.GetAssets(c, query)
		return ginH{"list": list}, err
	case "RootDomain":
		return d.rootDomainService.GetRootDomainData(c, query)
	case "subdomain":
		list, err := d.subdomainService.GetSubdomains(ctx, query)
		return ginH{"list": list}, err
	case "app":
		return d.appService.GetAppData(c, query)
	case "mp":
		return d.mpService.GetMPData(c, query)
	case "UrlScan":
		list, err := d.urlService.GetURLs(ctx, query)
		return ginH{"list": list}, err
	case "SensitiveResult":
		list, err := d.sensitiveService.GetSensitiveInfo(ctx, query)
		return ginH{"list": list}, err
	case "DirScanResult":
		list, err := d.dirscanService.List(ctx, query)
		return ginH{"list": list}, err
	case "crawler":
		list, err := d.crawlerService.GetCrawlers(ctx, query)
		return ginH{"list": list}, err
	case "vulnerability":
		return d.vulnService.GetVulnerabilities(ctx, query)
	case "PageMonitoring":
		return d.pageMonService.GetResult(c, query)
	case "IPAsset":
		return d.ipService.GetIPAssets(c, query)
	case "SubdomainTakerResult":
		return querySubdomainTaker(ctx, query)
	default:
		return nil, fmt.Errorf("unsupported asset type: %s", index)
	}
}

func querySubdomainTaker(ctx context.Context, query models.SearchRequest) (any, error) {
	takerService := subdomain.NewTakerService()
	c := ginContext(ctx)
	return takerService.GetSubdomainTakerData(c, query)
}

func normalizeAssetIndex(assetType string) (string, error) {
	assetType = strings.TrimSpace(assetType)
	if assetType == "" {
		return "", fmt.Errorf("asset_type cannot be empty")
	}
	aliases := map[string]string{
		"asset":                "asset",
		"web":                  "asset",
		"rootdomain":           "RootDomain",
		"root_domain":          "RootDomain",
		"root-domain":          "RootDomain",
		"subdomain":            "subdomain",
		"app":                  "app",
		"mp":                   "mp",
		"miniprogram":          "mp",
		"mini_program":         "mp",
		"url":                  "UrlScan",
		"urlscan":              "UrlScan",
		"sensitive":            "SensitiveResult",
		"sensitiveresult":      "SensitiveResult",
		"sensitive_result":     "SensitiveResult",
		"dirscan":              "DirScanResult",
		"dir_scan":             "DirScanResult",
		"dirscanresult":        "DirScanResult",
		"directory":            "DirScanResult",
		"crawler":              "crawler",
		"vulnerability":        "vulnerability",
		"vuln":                 "vulnerability",
		"pagemonitoring":       "PageMonitoring",
		"page_monitoring":      "PageMonitoring",
		"ip":                   "IPAsset",
		"ipasset":              "IPAsset",
		"subdomaintaker":       "SubdomainTakerResult",
		"subdomain_taker":      "SubdomainTakerResult",
		"subdomaintakerresult": "SubdomainTakerResult",
	}
	key := strings.ToLower(assetType)
	if v, ok := aliases[key]; ok {
		return v, nil
	}
	if _, ok := aliases[strings.ReplaceAll(key, "-", "_")]; ok {
		return aliases[strings.ReplaceAll(key, "-", "_")], nil
	}
	valid := []string{"asset", "RootDomain", "subdomain", "app", "mp", "UrlScan",
		"SensitiveResult", "DirScanResult", "crawler", "vulnerability",
		"PageMonitoring", "IPAsset", "SubdomainTakerResult"}
	for _, v := range valid {
		if v == assetType {
			return v, nil
		}
	}
	return "", fmt.Errorf("unsupported asset_type: %s", assetType)
}

type ginH map[string]any
