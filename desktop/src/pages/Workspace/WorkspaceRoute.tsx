import { useParams } from "react-router";
import { useWorkspace } from "../../hooks/useWorkspace";
import { LoadingOverlay, Text } from "@mantine/core";
import { PageWrapper } from "../../components/PageWrapper.tsx";
import { Workspace } from "./Workspace";
import { Welcome } from "../Welcome/Welcome";

export function WorkspaceRoute() {
    const { workspaceName } = useParams();
    const { workspace, workspaceError, isWorkspaceLoading } = useWorkspace(workspaceName || "");

    return (
        <PageWrapper>
            {isWorkspaceLoading && <LoadingOverlay visible={isWorkspaceLoading} zIndex={1000} overlayProps={{ radius: "sm", blur: 2 }} />}
            {workspaceError && <Text c="red">Error: {workspaceError.message}</Text>}
            {workspace ? (
                <Workspace workspace={workspace} />
            ) : (
                <Welcome welcomeMessage="Select a workspace to get started." />
            )}
        </PageWrapper>
    );
}