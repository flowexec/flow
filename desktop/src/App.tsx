import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createHashRouter, Navigate, RouterProvider } from "react-router";
import "./App.css";
import { AppProvider } from "./hooks/useAppContext.tsx";
import { NotifierProvider } from "./hooks/useNotifier";
import { AppShell } from "./layout";
import { PageWrapper } from "./components/PageWrapper.tsx";
import { Settings, Welcome, Data } from "./pages";
import { WorkspaceRoute } from "./pages/Workspace/WorkspaceRoute.tsx";
import { ExecutableRoute } from "./pages/Executable/ExecutableRoute.tsx";
import { Text } from "@mantine/core";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

const router = createHashRouter([
  {
    element: <AppShell />,
    children: [
      {
        path: "/",
        children: [
          {
            index: true,
            element: (
              <PageWrapper>
                <Welcome welcomeMessage="Hey!" />
              </PageWrapper>
            ),
          },
          {
            path: "workspace/:workspaceName",
            element: <WorkspaceRoute />,
          },
          {
            path: "executable/:executableId",
            element: <ExecutableRoute />,
          },
          {
            path: "logs",
            element: (
              <PageWrapper>
                <Text>Logs view coming soon...</Text>
              </PageWrapper>
            ),
          },
          {
            path: "vault",
            element: <Data />,
          },
          {
            path: "cache",
            element: <Data />,
          },
          {
            path: "settings",
            element: <Settings />,
          },
          {
            path: "*",
            element: <Navigate to="/" replace />,
          },
        ],
      },
    ],
  },
]);

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <NotifierProvider>
        <AppProvider>
          <RouterProvider router={router} />
        </AppProvider>
      </NotifierProvider>
    </QueryClientProvider>
  );
}
export default App;
