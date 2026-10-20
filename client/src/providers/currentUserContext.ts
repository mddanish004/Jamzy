import { createContext } from "react";
import type { MeResponse, PublicUser } from "../services/types";

export type CurrentUserState =
  | { status: "loading" }
  | { status: "unauthenticated" }
  | { status: "needs-onboarding" }
  | { status: "authenticated"; me: MeResponse }
  | { status: "error"; message: string };

export interface CurrentUserContextValue {
  state: CurrentUserState;
  refresh: () => void;
  applyProfile: (user: PublicUser) => void;
}

export const CurrentUserContext =
  createContext<CurrentUserContextValue | null>(null);
