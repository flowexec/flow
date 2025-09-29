import { Executable } from "./generated/flowfile";

export interface EnrichedExecutable extends Executable {
  id: string;
  ref: string;
  namespace: string | null;
  workspace: string;
  flowfile: string;
  fullDescription: string;
}

export const ExecutionVerbType = "execution";
export const DeactivationVerbType = "deactivation";
export const ConfigurationVerbType = "configuration";
export const DestructionVerbType = "destruction";
export const RetrievalVerbType = "retrieval";
export const UpdateVerbType = "update";
export const ValidationVerbType = "validation";
export const LaunchVerbType = "launch";
export const CreationVerbType = "creation";
export const RestartVerbType = "restart";
export const BuildVerbType = "build";

export function GetUIVerbType(executable: Executable): string | null {
  // This is not an exhaustive list of verbs. It's intended to only capture verbs
  // that should have an icon that is different than the default exec icon.
  switch (executable.verb) {
    case "deactivate":
    case "disable":
    case "stop":
    case "pause":
    case "kill":
    case "terminate":
    case "abort":
      return DeactivationVerbType;
    case "watch":
    case "monitor":
      return "monitoring";
    case "restart":
    case "reboot":
    case "reload":
    case "refresh":
      return RestartVerbType;
    case "install":
    case "setup":
    case "deploy":
    case "update":
    case "upgrade":
    case "patch":
    case "publish":
    case "release":
      return UpdateVerbType;
    case "build":
    case "package":
    case "bundle":
    case "compile":
      return BuildVerbType;
    case "configure":
    case "manage":
    case "set":
    case "edit":
    case "transform":
    case "modify":
      return ConfigurationVerbType;
    case "test":
    case "validate":
    case "check":
    case "verify":
    case "analyze":
    case "scan":
    case "lint":
    case "inspect":
      return ValidationVerbType;
    case "open":
    case "launch":
    case "show":
    case "view":
      return LaunchVerbType;
    case "create":
    case "generate":
    case "add":
    case "new":
    case "init":
      return CreationVerbType;
    case "remove":
    case "delete":
    case "destroy":
    case "erase":
    case "unset":
    case "reset":
    case "clean":
    case "clear":
    case "purge":
    case "tidy":
    case "uninstall":
    case "teardown":
    case "undeploy":
      return DestructionVerbType;
    case "retrieve":
    case "fetch":
    case "get":
    case "request":
      return RetrievalVerbType;
    default:
      return ExecutionVerbType;
  }
}
