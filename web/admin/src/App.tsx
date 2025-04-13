// src/App.tsx
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { useState, useEffect } from "react";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import ContentList from "./pages/content/ContentList";
import ContentEditor from "./pages/content/ContentEditor";
import MediaLibrary from "./pages/media/MediaLibrary";
import UserManagement from "./pages/users/UserManagement";
import WorkspaceSettings from "./pages/settings/WorkspaceSettings";
import Profile from "./pages/Profile";
import Layout from "./components/Layout";
import axios from "axios";
import AuthContext from "./context/AuthContext";
import ContentTypeList from "./pages/content-types/ContentTypeList";
import ContentTypeEditor from "./pages/content-types/ContentTypeEditor";
import NotFound from "./pages/NotFound";

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [user, setUser] = useState<any>(null);
  const [workspaces, setWorkspaces] = useState<any[]>([]);
  const [currentWorkspace, setCurrentWorkspace] = useState<any>(null);

  useEffect(() => {
    const token = localStorage.getItem("accessToken");

    if (token) {
      axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;
      fetchCurrentUser();
    } else {
      setIsLoading(false);
    }
  }, []);

  const fetchCurrentUser = async () => {
    try {
      const response = await axios.get("/api/me");
      setUser(response.data.data.user);
      setWorkspaces(response.data.data.workspaces);

      if (response.data.data.workspaces.length > 0) {
        setCurrentWorkspace(response.data.data.workspaces[0]);
      }

      // src/App.tsx (continued)
      setIsAuthenticated(true);
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
    } catch (error) {
      localStorage.removeItem("accessToken");
      localStorage.removeItem("refreshToken");
      delete axios.defaults.headers.common["Authorization"];
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  };

  const login = (tokens: { accessToken: string; refreshToken: string }) => {
    localStorage.setItem("accessToken", tokens.accessToken);
    localStorage.setItem("refreshToken", tokens.refreshToken);
    axios.defaults.headers.common[
      "Authorization"
    ] = `Bearer ${tokens.accessToken}`;
    fetchCurrentUser();
  };

  const logout = () => {
    localStorage.removeItem("accessToken");
    localStorage.removeItem("refreshToken");
    delete axios.defaults.headers.common["Authorization"];
    setIsAuthenticated(false);
    setUser(null);
    setWorkspaces([]);
    setCurrentWorkspace(null);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-primary-600 font-semibold text-xl">Loading...</div>
      </div>
    );
  }

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        user,
        workspaces,
        currentWorkspace,
        setCurrentWorkspace,
        login,
        logout,
      }}
    >
      <Router>
        <Routes>
          <Route
            path="/login"
            element={!isAuthenticated ? <Login /> : <Navigate to="/" />}
          />

          <Route
            element={isAuthenticated ? <Layout /> : <Navigate to="/login" />}
          >
            <Route path="/" element={<Dashboard />} />
            <Route path="/content" element={<ContentList />} />
            <Route path="/content/new" element={<ContentEditor />} />
            <Route path="/content/:id" element={<ContentEditor />} />
            <Route path="/media" element={<MediaLibrary />} />
            <Route path="/content-types" element={<ContentTypeList />} />
            <Route path="/content-types/new" element={<ContentTypeEditor />} />
            <Route path="/content-types/:id" element={<ContentTypeEditor />} />
            <Route path="/users" element={<UserManagement />} />
            <Route path="/settings" element={<WorkspaceSettings />} />
            <Route path="/profile" element={<Profile />} />
          </Route>

          <Route path="*" element={<NotFound />} />
        </Routes>
      </Router>
    </AuthContext.Provider>
  );
}

export default App;
