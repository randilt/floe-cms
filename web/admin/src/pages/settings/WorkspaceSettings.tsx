// src/pages/settings/WorkspaceSettings.tsx
import { useState, useEffect, useContext } from "react";
import axios from "axios";
import AuthContext from "../../context/AuthContext";

export default function WorkspaceSettings() {
  const { currentWorkspace, setCurrentWorkspace, workspaces } =
    useContext(AuthContext);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const [workspace, setWorkspace] = useState({
    name: "",
    slug: "",
    description: "",
  });

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newWorkspace, setNewWorkspace] = useState({
    name: "",
    slug: "",
    description: "",
  });

  useEffect(() => {
    if (currentWorkspace) {
      setWorkspace({
        name: currentWorkspace.name || "",
        slug: currentWorkspace.slug || "",
        description: currentWorkspace.description || "",
      });
    }
  }, [currentWorkspace]);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setWorkspace((prev) => ({ ...prev, [name]: value }));
  };

  const handleNewWorkspaceChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setNewWorkspace((prev) => ({ ...prev, [name]: value }));

    // Auto-generate slug from name
    if (name === "name" && (!newWorkspace.slug || newWorkspace.slug === "")) {
      const slug = value
        .toLowerCase()
        .replace(/[^\w\s-]/g, "")
        .replace(/\s+/g, "-")
        .replace(/--+/g, "-")
        .trim();

      setNewWorkspace((prev) => ({ ...prev, slug }));
    }
  };

  const handleUpdateWorkspace = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setIsLoading(true);

    try {
      const response = await axios.put(
        `/api/workspaces/${currentWorkspace.id}`,
        workspace
      );

      if (response.data.success) {
        setSuccess("Workspace updated successfully");
        // Update the current workspace in context
        setCurrentWorkspace({
          ...currentWorkspace,
          ...workspace,
        });
      } else {
        setError(response.data.error || "Failed to update workspace");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to update workspace");
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateWorkspace = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    try {
      const response = await axios.post("/api/workspaces", newWorkspace);

      if (response.data.success) {
        setSuccess("Workspace created successfully");
        setShowCreateModal(false);

        // Need to reload workspaces from API
        // This is a simplified approach - in a real app, you'd want to update the workspaces in context
        window.location.reload();
      } else {
        setError(response.data.error || "Failed to create workspace");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to create workspace");
    }
  };

  return (
    <div className="max-w-3xl mx-auto">
      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
          <h1 className="text-xl font-semibold text-gray-900">
            Workspace Settings
          </h1>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700"
          >
            Create Workspace
          </button>
        </div>

        {error && (
          <div className="px-4 py-3 bg-red-50 border-b border-red-200 text-red-700 text-sm">
            {error}
          </div>
        )}

        {success && (
          <div className="px-4 py-3 bg-green-50 border-b border-green-200 text-green-700 text-sm">
            {success}
          </div>
        )}

        <div className="px-4 py-5 sm:p-6">
          <form onSubmit={handleUpdateWorkspace}>
            <div className="space-y-6">
              <div>
                <label
                  htmlFor="name"
                  className="block text-sm font-medium text-gray-700"
                >
                  Workspace Name
                </label>
                <input
                  type="text"
                  name="name"
                  id="name"
                  required
                  value={workspace.name}
                  onChange={handleInputChange}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                />
              </div>

              <div>
                <label
                  htmlFor="slug"
                  className="block text-sm font-medium text-gray-700"
                >
                  Slug
                </label>
                <input
                  type="text"
                  name="slug"
                  id="slug"
                  required
                  value={workspace.slug}
                  onChange={handleInputChange}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                />
                <p className="mt-1 text-xs text-gray-500">
                  The slug is used in URLs for your content. Use only lowercase
                  letters, numbers, and hyphens.
                </p>
              </div>

              <div>
                <label
                  htmlFor="description"
                  className="block text-sm font-medium text-gray-700"
                >
                  Description
                </label>
                <textarea
                  name="description"
                  id="description"
                  rows={3}
                  value={workspace.description}
                  onChange={handleInputChange}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                />
              </div>

              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={isLoading}
                  className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 disabled:bg-primary-400"
                >
                  {isLoading ? "Saving..." : "Save Changes"}
                </button>
              </div>
            </div>
          </form>
        </div>
      </div>

      {/* Workspace List */}
      <div className="mt-6 bg-white shadow rounded-lg overflow-hidden">
        <div className="px-4 py-5 sm:px-6">
          <h2 className="text-lg font-medium text-gray-900">Your Workspaces</h2>
        </div>

        <div className="border-t border-gray-200">
          <ul className="divide-y divide-gray-200">
            {workspaces.map((ws) => (
              <li
                key={ws.id}
                className="px-4 py-4 flex items-center justify-between"
              >
                <div>
                  <p className="text-sm font-medium text-gray-900">{ws.name}</p>
                  <p className="text-sm text-gray-500">{ws.slug}</p>
                </div>
                <button
                  onClick={() => setCurrentWorkspace(ws)}
                  className={`px-3 py-1 rounded-md text-xs font-medium ${
                    currentWorkspace?.id === ws.id
                      ? "bg-primary-100 text-primary-800"
                      : "text-primary-600 hover:bg-primary-50"
                  }`}
                >
                  {currentWorkspace?.id === ws.id ? "Current" : "Switch"}
                </button>
              </li>
            ))}
          </ul>
        </div>
      </div>

      {/* Create Workspace Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-medium text-gray-900">
                Create New Workspace
              </h2>
              <button
                onClick={() => setShowCreateModal(false)}
                className="text-gray-400 hover:text-gray-500"
              >
                <svg
                  className="h-6 w-6"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  ></path>
                </svg>
              </button>
            </div>

            <form onSubmit={handleCreateWorkspace}>
              <div className="space-y-4">
                <div>
                  <label
                    htmlFor="new-name"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Workspace Name
                  </label>
                  <input
                    type="text"
                    name="name"
                    id="new-name"
                    required
                    value={newWorkspace.name}
                    onChange={handleNewWorkspaceChange}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                  />
                </div>

                <div>
                  <label
                    htmlFor="new-slug"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Slug
                  </label>
                  <input
                    type="text"
                    name="slug"
                    id="new-slug"
                    required
                    value={newWorkspace.slug}
                    onChange={handleNewWorkspaceChange}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                  />
                </div>

                <div>
                  <label
                    htmlFor="new-description"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Description
                  </label>
                  <textarea
                    name="description"
                    id="new-description"
                    rows={3}
                    value={newWorkspace.description}
                    onChange={handleNewWorkspaceChange}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                  />
                </div>
              </div>

              <div className="mt-6 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700"
                >
                  Create Workspace
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
