use std::result::Result;
use tauri::Manager;

pub mod cli;
pub mod commands;
pub mod types;

pub use types::{enriched, generated};

#[tauri::command]
async fn check_flow_binary() -> Result<(), String> {
    let cli = commands::core::CliCommand::new();
    cli.check_binary().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_config() -> Result<generated::config::Config, String> {
    let runner = cli::cli_executor();
    runner.config.get().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn set_config_theme(theme: String) -> Result<String, String> {
    let runner = cli::cli_executor();
    runner
        .config
        .set_theme(&theme)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn set_config_workspace_mode(mode: String) -> Result<String, String> {
    let runner = cli::cli_executor();
    runner
        .config
        .set_workspace_mode(&mode)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn set_config_log_mode(mode: String) -> Result<String, String> {
    let runner = cli::cli_executor();
    runner
        .config
        .set_log_mode(&mode)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn set_config_namespace(namespace: String) -> Result<String, String> {
    let runner = cli::cli_executor();
    runner
        .config
        .set_namespace(&namespace)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn set_config_timeout(timeout: String) -> Result<String, String> {
    let runner = cli::cli_executor();
    runner
        .config
        .set_timeout(&timeout)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_workspace(name: String) -> Result<enriched::Workspace, String> {
    let runner = cli::cli_executor();
    runner.workspace.get(&name).await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn set_workspace(name: String, fixed: bool) -> Result<String, String> {
    let runner = cli::cli_executor();
    runner
        .workspace
        .switch(&name, fixed)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn list_workspaces() -> Result<Vec<enriched::Workspace>, String> {
    let runner = cli::cli_executor();
    runner.workspace.list().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_executable(executable_ref: String) -> Result<enriched::Executable, String> {
    let runner = cli::cli_executor();
    runner
        .executable
        .get(&executable_ref)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn list_executables(
    workspace: Option<String>,
    namespace: Option<String>,
    tags: Option<Vec<String>>,
    verb: Option<String>,
    filter: Option<String>,
) -> Result<Vec<enriched::Executable>, String> {
    let runner = cli::cli_executor();

    // Convert owned Strings into borrowed &str slices expected by the command layer
    let tags_ref_vec: Option<Vec<&str>> = tags
        .as_ref()
        .map(|v| v.iter().map(|s| s.as_str()).collect());

    runner
        .executable
        .list(
            workspace.as_deref(),
            namespace.as_deref(),
            tags_ref_vec.as_deref(),
            verb.as_deref(),
            filter.as_deref(),
        )
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn sync() -> Result<String, String> {
    let runner = cli::cli_executor();
    runner.executable.sync().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn execute(
    app: tauri::AppHandle,
    verb: String,
    executable_id: String,
    args: Vec<String>,
    params: Option<std::collections::HashMap<String, String>>,
) -> Result<(), String> {
    let runner = cli::cli_executor();
    let args: Vec<&str> = args.iter().map(|s| s.as_str()).collect();
    runner
        .executable
        .execute::<()>(app, &verb, &executable_id, &args, params)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn reload_window(app: tauri::AppHandle) -> Result<(), String> {
    app.get_webview_window("main")
        .ok_or_else(|| "Main window not found".to_string())?
        .reload()
        .map_err(|e| e.to_string())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_log::Builder::new().build())
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_opener::init())
        .setup(|_app| Ok(()))
        .invoke_handler(tauri::generate_handler![
            check_flow_binary,
            sync,
            execute,
            list_executables,
            get_executable,
            get_workspace,
            set_workspace,
            list_workspaces,
            get_config,
            reload_window,
            set_config_theme,
            set_config_workspace_mode,
            set_config_log_mode,
            set_config_namespace,
            set_config_timeout,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
