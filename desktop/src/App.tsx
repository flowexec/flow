import "@mantine/core/styles.css";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {createHashRouter, Navigate} from "react-router";
import {RouterProvider} from "react-router/dom";
import "./App.css";
import { ThemeProvider } from "./theme/ThemeProvider";
import { AppProvider } from "./hooks/useAppContext.tsx";
import { NotifierProvider } from "./hooks/useNotifier";
import { SettingsProvider } from "./hooks/useSettings";
import {AppShell} from "./layout";
import {PageWrapper} from "./components/PageWrapper.tsx";
import {Settings, Welcome, Data} from "./pages";
import {WorkspaceRoute} from "./pages/Workspace/WorkspaceRoute.tsx";
import {ExecutableRoute} from "./pages/Executable/ExecutableRoute.tsx";
import {Text} from "@mantine/core";

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
    path: "/",
    element: (
        <QueryClientProvider client={queryClient}>
          <NotifierProvider>
            <AppProvider>
              <SettingsProvider>
                <ThemeProvider>
                  <AppShell />
                </ThemeProvider>
              </SettingsProvider>
            </AppProvider>
          </NotifierProvider>
        </QueryClientProvider>
    ),
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
        element: <PageWrapper><Text>Logs view coming soon...</Text></PageWrapper>,
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
]);

function App() {
  return <RouterProvider router={router} />

}
export default App;
