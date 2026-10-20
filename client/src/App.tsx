import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { SignIn, SignUp } from "@clerk/react";
import { Analytics } from "@vercel/analytics/react";
import { Landing } from "./pages/Landing";
import { RootGuard } from "./pages/RootGuard";
import { Onboarding } from "./pages/Onboarding";
import { Home } from "./pages/Home";
import { Room } from "./pages/Room";
import { ProtectedRoute, RequireAuth } from "./components/ProtectedRoute";
import { CurrentUserProvider } from "./providers/CurrentUserProvider";
import { AuthShell } from "./components/ui/AuthShell";

const clerkAppearance = {
  variables: {
    colorPrimary: "#fb7185",
    colorText: "#e5e1e4",
    colorTextSecondary: "#a58a8c",
    colorBackground: "#1c1b1d",
    colorInputBackground: "#0e0e10",
    colorInputText: "#e5e1e4",
    colorDanger: "#ffb4ab",
    colorSuccess: "#10b981",
    borderRadius: "8px",
    fontFamily: '"Plus Jakarta Sans", ui-sans-serif, system-ui, sans-serif',
  },
};

function App() {
  return (
    <BrowserRouter>
      <Analytics />
      <CurrentUserProvider>
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/enter" element={<RootGuard />} />

          <Route
            path="/signin"
            element={
              <AuthShell>
                <div className="jamzy-clerk w-full">
                  <SignIn
                    appearance={clerkAppearance}
                    fallbackRedirectUrl="/enter"
                    forceRedirectUrl="/enter"
                    signUpUrl="/signup"
                  />
                </div>
              </AuthShell>
            }
          />
          <Route
            path="/signup"
            element={
              <AuthShell>
                <div className="jamzy-clerk w-full">
                  <SignUp
                    appearance={clerkAppearance}
                    fallbackRedirectUrl="/enter"
                    forceRedirectUrl="/enter"
                    signInUrl="/signin"
                  />
                </div>
              </AuthShell>
            }
          />

          <Route
            path="/onboarding"
            element={
              <RequireAuth>
                <Onboarding />
              </RequireAuth>
            }
          />
          <Route
            path="/home"
            element={
              <ProtectedRoute>
                <Home />
              </ProtectedRoute>
            }
          />

          <Route
            path="/room/:roomCode"
            element={
              <ProtectedRoute>
                <Room />
              </ProtectedRoute>
            }
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </CurrentUserProvider>
    </BrowserRouter>
  );
}

export default App;
