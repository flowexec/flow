import { Text, LoadingOverlay } from "@mantine/core";
import { useParams } from "react-router";
import { useExecutable } from "../../hooks/useExecutable";
import { PageWrapper } from "../../components/PageWrapper.tsx";
import { Welcome } from "../Welcome/Welcome";
import  { Executable } from "./Executable";

export function ExecutableRoute() {
    const { executableId } = useParams();
    const { executable, executableError, isExecutableLoading } = useExecutable(executableId || "");

    return (
        <PageWrapper>
            {isExecutableLoading && <LoadingOverlay visible={isExecutableLoading} zIndex={1000} overlayProps={{ radius: "sm", blur: 2 }} />}
            {executableError && <Text c="red">Error: {executableError.message}</Text>}
            {executable ? (
                <Executable executable={executable} />
            ) : (
                <Welcome welcomeMessage="Select an executable to get started." />
            )}
        </PageWrapper>
    );
}