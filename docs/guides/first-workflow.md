---
title: Your First Workflow
---

# Your First Workflow

Now that you understand flow's [core concepts](concepts.md), let's build a real workflow that shows how executables, workspaces, and secrets work together.

We'll create a simple web project deployment workflow that:
- Builds a static site
- Runs tests
- Deploys to a server (simulated)

## Setup

Create a new workspace for this tutorial:

```shell
mkdir ~/flow-tutorial
cd ~/flow-tutorial
flow workspace add tutorial . --set
```

## Step 1: Create the Project Structure

Let's simulate a simple web project:

```shell
# Create project files
mkdir -p src tests
echo "<h1>My Website</h1>" > src/index.html
echo "console.log('Testing...');" > tests/test.js
echo "build/" > .gitignore
```

## Step 2: Build Your First Workflow

Create a `deploy.flow` file:

```yaml
# deploy.flow
executables:
  - verb: build
    name: site
    exec:
      cmd: |
        echo "Building website..."
        mkdir -p build
        cp -r src/* build/
        echo "✅ Build complete"

  - verb: test
    name: site
    exec:
      cmd: |
        echo "Running tests..."
        node tests/test.js
        echo "✅ Tests passed"

  - verb: deploy
    name: full
    serial:
      execs:
        - ref: build site
        - ref: test site
        - cmd: |
            echo "Deploying to server..."
            echo "📦 Deployment complete!"
```

## Step 3: Run Individual Steps

Sync and try each step:

```shell
flow sync

# Try each step individually
flow build site
flow test site
```

## Step 4: Run the Full Workflow

Now run the complete deployment:

```shell
flow deploy full
```

You'll see each step run in sequence. This is your first multi-step workflow!

## Step 5: Add Configuration with Secrets

Real deployments need configuration. Let's add some secrets:

```shell
# Create a vault for this project and set the generated key in the default environment variable
export FLOW_VAULT_KEY="$(flow vault create tutorial-vault --set --log-level fatal)"

# Add some deployment secrets
flow secret set server-url "https://my-server.com"
flow secret set api-key "your-secret-key-here"
```

## Step 6: Use Secrets in Your Workflow

Update your `deploy.flow` to use secrets:

```yaml
# deploy.flow
executables:
  # Build and test steps remain the same
  
  # Update deploy step to use secrets
  - verb: deploy
    name: full
    serial:
      execs:
        - ref: build site
        - ref: test site
        - cmd: |
            echo "Deploying to $SERVER_URL..."
            echo "Using API key: ${API_KEY:0:8}..."
            echo "📦 Deployment complete!"
      params:
        - secretRef: server-url
          envKey: SERVER_URL
        - secretRef: api-key
          envKey: API_KEY
```

Run it again:

```shell
flow deploy full
```

Now your workflow uses secure configuration!

## Step 7: Add Interactive Elements

Let's make the workflow more interactive by adding prompts:

```yaml
# Add this executable to deploy.flow
  - verb: deploy
    name: interactive
    exec:
      params:
        - prompt: "Which environment? (dev/staging/prod)"
          envKey: ENVIRONMENT
        - prompt: "Run tests first? (y/n)"
          envKey: RUN_TESTS
        - secretRef: server-url
          envKey: SERVER_URL
      cmd: |
        echo "Deploying to $ENVIRONMENT environment..."
        
        if [ "$RUN_TESTS" = "y" ]; then
          echo "Running tests first..."
          node tests/test.js
        fi
        
        echo "Deploying to $SERVER_URL..."
        echo "🚀 $ENVIRONMENT deployment complete!"
```

Try the interactive version:

```shell
flow sync
flow deploy interactive
```

## Step 8: Browse Your Workflows

Use the TUI to explore what you've built:

```shell
flow browse
```

Navigate through your executables, view their details, and run them directly from the interface.

## Recap

Congratulations! You've built a workflow that demonstrates:

✅ **Multi-step workflows** with `serial` executables  
✅ **Executable references** with `ref` to reuse steps  
✅ **Secret management** with secure configuration  
✅ **Interactive prompts** for runtime customization

### More Examples

Want to see more workflow patterns? Check out:
- [flow examples repo](https://github.com/flowexec/examples) - Collection of workflow patterns and scaffolding
- [flow project itself](https://github.com/flowexec/flow/tree/main/.execs) - Real development workflows for building flow

Both contain flow files you can explore and adapt for your own projects.

## Next Steps

Ready to level up your flow skills?

- **Secure more workflows** → [Working with secrets](secrets.md)
- **Build complex automations** → [Advanced workflows](advanced.md)
- **Generate new projects** → [Templates & code generation](templating.md)
- **Customize your experience** → [Interactive UI](interactive.md)
