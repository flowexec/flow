//nolint:lll
package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func addServerPrompts(s *server.MCPServer) {
	generateExecutable := mcp.NewPrompt("generate_executable",
		mcp.WithPromptDescription("Generate a Flow executable configuration"),
		mcp.WithArgument("task_description",
			mcp.ArgumentDescription("What task should this executable perform?"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("verb",
			mcp.ArgumentDescription("Preferred action verb (build, test, deploy, run, etc.)"),
		),
		mcp.WithArgument("execution_type",
			mcp.ArgumentDescription("e.g. single command, shell file, workflow / multiple steps, REST, launch, etc."),
		),
		mcp.WithArgument("context",
			mcp.ArgumentDescription("Technology stack, environment, constraints"),
		),
	)
	s.AddPrompt(generateExecutable, generateExecutablePrompt)

	generateProjectExecutables := mcp.NewPrompt("generate_project_executables",
		mcp.WithPromptDescription("Generate a complete set of executables for a project"),
		mcp.WithArgument("project_type",
			mcp.ArgumentDescription("e.g. web app, API, CLI tool, mobile app, etc."),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("tech_stack",
			mcp.ArgumentDescription("e.g., 'Node.js + React', 'Go + Docker', 'Python + FastAPI'"),
		),
		mcp.WithArgument("development_stage",
			mcp.ArgumentDescription("e.g. 'new' (starting fresh), 'existing' (has some automation), 'mature' (complex workflows)"),
		),
	)
	s.AddPrompt(generateProjectExecutables, generateProjectExecutablesPrompt)

	debugExecutable := mcp.NewPrompt("debug_executable",
		mcp.WithPromptDescription("Debug a failing Flow executable"),
		mcp.WithArgument("executable_ref",
			mcp.ArgumentDescription("The executable that's failing (e.g., 'build app', 'deploy prod')"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("failure_description",
			mcp.ArgumentDescription("Brief description of what's going wrong"),
		),
	)
	s.AddPrompt(debugExecutable, debugExecutablePrompt)

	migrateAutomation := mcp.NewPrompt("migrate_automation",
		mcp.WithPromptDescription("Convert existing automation (Makefile, npm scripts, shell scripts, etc.) to Flow executables"),
		mcp.WithArgument("automation_type",
			mcp.ArgumentDescription("Current automation system (Makefile, package.json scripts, shell scripts, GitHub Actions, etc.)"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("current_tasks",
			mcp.ArgumentDescription("List or describe the tasks you currently run (e.g., 'build, test, deploy to staging')"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("pain_points",
			mcp.ArgumentDescription("What's frustrating about your current setup?"),
		),
	)
	s.AddPrompt(migrateAutomation, migrateAutomationPrompt)

	explainFlow := mcp.NewPrompt("explain_flow",
		mcp.WithPromptDescription("Explain Flow concepts and help with getting started"),
		mcp.WithArgument("topic",
			mcp.ArgumentDescription("What to explain: 'basics', 'workspaces', 'executables', 'secrets', 'templates', 'getting-started'"),
		),
		mcp.WithArgument("user_background",
			mcp.ArgumentDescription("Your experience level: 'new' (never used Flow), 'beginner' (basic usage), 'intermediate' (regular user)"),
		),
	)
	s.AddPrompt(explainFlow, explainFlowPrompt)
}

func generateExecutablePrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := request.Params.Arguments
	taskDescription := getArgOrDefault(args, "task_description", "")
	verb := getArgOrDefault(args, "verb", "")
	executionType := getArgOrDefault(args, "execution_type", "")
	promptContext := getArgOrDefault(args, "context", "")

	fmtStr := `I need to create a Flow executable for this task:

**Task**: %s
**Preferred Verb**: %s
**Execution Type**: %s
**Context**: %s

Please generate a complete Flow executable configuration that:

1. **Determines the Best Approach**:
   - Choose the most appropriate executable type (exec, serial, parallel, request, launch)
   - Select a suitable verb if none was provided
   - Design the proper parameter and argument structure

2. **Creates Production-Ready Configuration**:
   - Valid YAML syntax following Flow conventions
   - Appropriate error handling (timeouts, retries where applicable)
   - Clear documentation and descriptions
   - Secure parameter handling (secrets, environment variables)

3. **Includes Usage Guidance**:
   - How to run the executable
   - What parameters or arguments are needed
   - Expected behavior and outputs
   - Any prerequisites or setup required

4. **Follows Best Practices**:
   - Proper naming conventions
   - Appropriate visibility settings
   - Helpful tags for organization
   - Integration considerations

If you need clarification about the task requirements, ask specific questions to ensure the generated executable meets the intended use case.`

	return mcp.NewGetPromptResult(
		"Generate Flow Executable",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(fmt.Sprintf(fmtStr, taskDescription, verb, executionType, promptContext)),
			),
		},
	), nil
}

func generateProjectExecutablesPrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := request.Params.Arguments
	projectType := getArgOrDefault(args, "project_type", "")
	techStack := getArgOrDefault(args, "tech_stack", "")
	developmentStage := getArgOrDefault(args, "development_stage", "new")

	fmtStr := `I want to set up Flow executables for my project:

**Project Type**: %s
**Technology Stack**: %s
**Development Stage**: %s

Please generate a comprehensive set of Flow executables that covers:

1. **Core Development Workflow**:
   - Build/compile processes
   - Testing (unit, integration, e2e as applicable)
   - Development server/watch modes
   - Code quality checks (linting, formatting)

2. **Deployment & Operations**:
   - Local environment setup
   - Staging/production deployment
   - Environment configuration management
   - Health checks and monitoring

3. **Project-Specific Tasks**:
   - Tasks specific to %s projects
   - Technology stack integrations for %s
   - Common maintenance and utility tasks

4. **Organization & Structure**:
   - Logical grouping with namespaces
   - Proper executable naming and descriptions
   - Dependencies and execution order
   - Parameter sharing between related tasks

5. **Documentation**:
   - Clear descriptions for each executable
   - Usage examples and common workflows
   - Getting started guide for team members
   - Integration with existing project structure

Provide complete YAML configurations ready to use, organized in a way that makes sense for a project type and development stage.`

	return mcp.NewGetPromptResult(
		"Generate Project Executables",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(fmt.Sprintf(fmtStr, projectType, techStack, developmentStage, projectType, techStack)),
			),
		},
	), nil
}

func debugExecutablePrompt(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := req.Params.Arguments
	execRef := getArgOrDefault(args, "executable_ref", "")
	failureDescription := getArgOrDefault(args, "failure_description", "")

	fmtStr := `I need help debugging a Flow executable that's failing:

**Executable**: %s
**Problem Description**: %s

Please help me troubleshoot by:

1. **Initial Analysis**:
   - Use the available MCP tools to check the executable configuration
   - Retrieve recent logs to understand what's happening
   - Identify common patterns in the failure

2. **Root Cause Investigation**:
   - Analyze the executable's configuration for issues
   - Check for common Flow executable problems
   - Consider environment and dependency issues
   - Review parameter and argument handling

3. **Diagnostic Commands**:
   - Specific Flow CLI commands to gather more information
   - How to test components individually
   - Ways to isolate the problem

4. **Solution Recommendations**:
   - Specific fixes for identified issues
   - Configuration adjustments needed
   - Alternative approaches if current method is problematic
   - Prevention strategies for similar issues

5. **Verification Steps**:
   - How to test that fixes work
   - Commands to validate the executable
   - Signs that everything is working correctly

Start by using the available tools to gather information about the executable and its recent execution history, then provide targeted debugging advice.`

	return mcp.NewGetPromptResult(
		"Debug Flow Executable",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(fmt.Sprintf(fmtStr, execRef, failureDescription)),
			),
		},
	), nil
}

func migrateAutomationPrompt(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := req.Params.Arguments
	automationType := getArgOrDefault(args, "automation_type", "")
	currentTasks := getArgOrDefault(args, "current_tasks", "")
	painPoints := getArgOrDefault(args, "pain_points", "")

	fmtStr := `I want to migrate my existing automation to Flow:

**Current System**: %s
**Tasks I Run**: %s
**Pain Points**: %s

Please help me convert this to Flow by providing:

1. **Migration Strategy**:
   - How to organize these tasks as Flow executables
   - Recommended workspace and namespace structure
   - Gradual migration approach to minimize disruption

2. **Flow Equivalents**:
   - Convert each existing task to a Flow executable
   - Preserve all current functionality
   - Improve error handling and logging where possible
   - Maintain compatibility during transition

3. **Enhanced Capabilities**:
   - Take advantage of Flow's parameter and secret management
   - Add proper documentation and discoverability
   - Implement conditional logic and workflows where beneficial
   - Better organization and team sharing

4. **Specific Improvements**:
   - Address the pain points I mentioned
   - Flow-specific enhancements over current approach
   - Better maintenance and debugging capabilities
   - Integration opportunities with other tools

5. **Implementation Plan**:
   - Step-by-step migration instructions
   - How to test converted workflows
   - Rollback strategy if needed

Provide concrete YAML configurations and specific commands to make the migration as smooth as possible. 
Focus on solving the current pain points while maintaining familiar workflows.`

	return mcp.NewGetPromptResult(
		"Migrate Automation to Flow",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(fmt.Sprintf(fmtStr, automationType, currentTasks, painPoints)),
			),
		},
	), nil
}

func explainFlowPrompt(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := req.Params.Arguments
	topic := getArgOrDefault(args, "topic", "basics")
	userBackground := getArgOrDefault(args, "user_background", "new")

	fmtStr := `Please explain Flow concepts to me:

**Topic**: %s
**My Experience Level**: %s

Please provide a clear explanation that:

1. **Matches My Level**:
   - Adjust complexity and detail for users with my experience level
   - Use appropriate examples and analogies
   - Build on concepts I likely already understand

2. **Covers Key Concepts**:
   - Core ideas and terminology for the desired topic
   - How it fits into the broader Flow ecosystem
   - Common use cases and patterns

3. **Provides Practical Examples**:
   - Real-world scenarios and configurations
   - Common commands and workflows
   - Hands-on exercises if appropriate

4. **Addresses Common Questions**:
   - Typical confusion points for users with my experience level
   - How this relates to other automation tools
   - Best practices and gotchas to avoid

5. **Next Steps**:
   - What to learn or try next
   - Resources for deeper understanding
   - How to get started with practical implementation

Make the explanation conversational and practical, focusing on helping me understand not just what something is, 
but why it's useful and how to use it effectively. You should fetch any additional information you need directly
from the Flow documentation at https://flowexec.io/#/README

Do not add confusion by mentioning other tools or platforms unless directly relevant to the explanation. Clearly state
when you do not have enough information to provide a complete answer, and suggest where I can find more details.`

	return mcp.NewGetPromptResult(
		"Explain Flow Concepts",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(fmt.Sprintf(fmtStr, topic, userBackground)),
			),
		},
	), nil
}

func getArgOrDefault(args map[string]string, key, defaultVal string) string {
	if val, exists := args[key]; exists {
		return val
	}
	return defaultVal
}
