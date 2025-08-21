import ReactDOM from "react-dom/client";
import App from "./App";
import { Layout } from "./layout/Layout";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <Layout>
    <App />
  </Layout>
);
