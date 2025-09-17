import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Route, Switch } from "wouter";
import "./App.css";
import { AppProvider } from "./hooks/useAppContext.tsx";
import { NotifierProvider } from "./hooks/useNotifier";
import { AppShell } from "./layout";
import { PageWrapper } from "./components/PageWrapper.tsx";
import {Settings, Welcome, Data, Workspaces} from "./pages";
import { Workspace } from "./pages/Workspace/Workspace.tsx";
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

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <NotifierProvider>
        <AppProvider>
          <AppShell>
            <Switch>
              <Route path="/">
                <PageWrapper>
                  <Welcome />
                </PageWrapper>
              </Route>
              <Route
                  path="/workspaces"
                  component={Workspaces}
              />
              <Route
                path="/workspace/:workspaceName"
                component={Workspace}
              />
              <Route
                path="/executable/:executableId"
                component={ExecutableRoute}
              />
              <Route path="/logs">
                <PageWrapper>
                  <Text>Logs view coming soon...</Text>
                </PageWrapper>
              </Route>
              <Route path="/vault" component={Data} />
              <Route path="/cache" component={Data} />
              <Route path="/settings" component={Settings} />
              <Route>
                <PageWrapper>
                  <Welcome welcomeMessage="404: What you are looking for couldn't be found" />
                </PageWrapper>
              </Route>
            </Switch>
          </AppShell>
        </AppProvider>
      </NotifierProvider>
    </QueryClientProvider>
  );
}
export default App;
