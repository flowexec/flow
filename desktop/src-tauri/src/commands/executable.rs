use crate::commands::command_executor::CommandExecutor;
use crate::commands::core::{CommandError, CommandResult};
use crate::types::enriched::Executable;
use serde::Deserialize;

#[derive(Deserialize, Debug)]
struct ExecutableResponse {
    executables: Vec<Executable>,
}

pub struct ExecutableCommands<E: CommandExecutor + 'static> {
    executor: std::sync::Arc<E>,
}

impl<E: CommandExecutor + 'static> ExecutableCommands<E> {
    pub fn new(executor: std::sync::Arc<E>) -> Self {
        Self { executor }
    }

    pub async fn sync(&self) -> CommandResult<String> {
        self.executor.execute::<()>(&["sync"]).await
    }

    pub async fn list(
        &self,
        workspace: Option<&str>,
        namespace: Option<&str>,
        tags: Option<&[&str]>,
        verb: Option<&str>,
        filter: Option<&str>,
    ) -> CommandResult<Vec<Executable>> {
        let mut args = vec!["browse", "--list"];

        if let Some(ws) = workspace {
            if !ws.is_empty() {
                args.extend_from_slice(&["--workspace", ws]);
            }
        }

        if let Some(ns) = namespace {
            if !ns.is_empty() {
                args.extend_from_slice(&["--namespace", ns]);
            } else {
                args.push("--all");
            }
        } else {
            args.push("--all");
        }
        
        if let Some(tags) = tags {
            for tag in tags {
                if !tag.is_empty() {
                    args.extend_from_slice(&["--tag", tag]);
                }
            }
        }
        
        if let Some(verb) = verb {
            if !verb.is_empty() {
                args.extend_from_slice(&["--verb", verb]);
            }
        }
        
        if let Some(filter) = filter {
            if !filter.is_empty() {
                args.extend_from_slice(&["--filter", filter]);
            }
        }

        let response: ExecutableResponse = self.executor.execute_json(&args).await?;
        Ok(response.executables)
    }

    pub async fn get(&self, exec_ref: &str) -> CommandResult<Executable> {
        let split_ref: Vec<&str> = exec_ref.split(" ").collect();
        match split_ref.len() {
            1 => {
                // Just a verb
                self.executor.execute_json(&["browse", split_ref[0]]).await
            }
            2 => {
                // Verb and ID
                self.executor
                    .execute_json(&["browse", split_ref[0], split_ref[1]])
                    .await
            }
            _ => Err(CommandError::ParseError {
                message: format!("Invalid executable reference format: {}", exec_ref),
                command: format!("{:?}", exec_ref),
                output: String::new(),
            }),
        }
    }

    pub async fn execute<T: for<'de> Deserialize<'de> + Send>(
        &self,
        app: tauri::AppHandle,
        verb: &str,
        executable_id: &str,
        args: &[&str],
        params: Option<std::collections::HashMap<String, String>>,
    ) -> CommandResult<T> {
        self.executor
            .execute_executable(app, verb, executable_id, args, params)
            .await
    }
}
