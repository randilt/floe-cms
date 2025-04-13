/* eslint-disable @typescript-eslint/no-explicit-any */
// src/context/AuthContext.tsx
import { createContext } from "react";

interface AuthContextType {
  isAuthenticated: boolean;
  user: any;
  workspaces: any[];
  currentWorkspace: any;
  setCurrentWorkspace: (workspace: any) => void;
  login: (tokens: { accessToken: string; refreshToken: string }) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  user: null,
  workspaces: [],
  currentWorkspace: null,
  setCurrentWorkspace: () => {},
  login: () => {},
  logout: () => {},
});

export default AuthContext;
