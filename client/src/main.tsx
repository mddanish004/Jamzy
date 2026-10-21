import { ClerkProvider } from "@clerk/react";
import { Analytics } from "@vercel/analytics/react";
import { StrictMode } from "react";
import { BrowserRouter } from "react-router-dom";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { clientConfig } from "./config/env";
import { AuthNotConfiguredScreen } from "./components/AuthStatusScreen";

const root = createRoot(document.getElementById("root")!);

if (!clientConfig.clerkPublishableKey) {
  root.render(
    <StrictMode>
      <Analytics />
      <BrowserRouter>
        <AuthNotConfiguredScreen />
      </BrowserRouter>
    </StrictMode>,
  );
} else {
  root.render(
    <StrictMode>
      <Analytics />
      <ClerkProvider
        publishableKey={clientConfig.clerkPublishableKey}
        afterSignOutUrl="/"
      >
        <App />
      </ClerkProvider>
    </StrictMode>,
  );
}
