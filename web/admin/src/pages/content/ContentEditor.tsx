// src/pages/content/ContentEditor.tsx
import { useState, useEffect, useContext } from "react";
import { useParams, useNavigate } from "react-router-dom";
import MarkdownEditor from "react-markdown-editor-lite";
import ReactMarkdown from "react-markdown";
import "react-markdown-editor-lite/lib/index.css";
import axios from "axios";
import AuthContext from "../../context/AuthContext";

export default function ContentEditor() {
  const { id } = useParams();
  const { currentWorkspace, workspaces, setCurrentWorkspace } =
    useContext(AuthContext);
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [contentTypes, setContentTypes] = useState<any[]>([]);

  const [content, setContent] = useState({
    title: "",
    slug: "",
    body: "",
    status: "draft",
    content_type_id: "",
    workspace_id: currentWorkspace?.id || "",
    meta_data: "",
  });

  useEffect(() => {
    if (currentWorkspace?.id) {
      loadContentTypes();
      if (id && id !== "new") {
        loadContent(id);
      }
    }
  }, [currentWorkspace, id]);

  const loadContentTypes = async () => {
    try {
      const response = await axios.get(
        `/api/content-types?workspace_id=${currentWorkspace.id}`
      );
      if (response.data.success) {
        setContentTypes(response.data.data);

        // Set default content type if available
        if (response.data.data.length > 0 && !content.content_type_id) {
          setContent((prev) => ({
            ...prev,
            content_type_id: response.data.data[0].id,
          }));
        }
      }
    } catch (err: any) {
      setError("Failed to load content types");
      console.error(err);
    }
  };

  const loadContent = async (contentId: string) => {
    setIsLoading(true);
    setError("");

    try {
      const response = await axios.get(
        `/api/workspaces/${currentWorkspace.id}/content/${contentId}`
      );
      if (response.data.success) {
        // If content belongs to a different workspace than the current one,
        // make sure we handle it properly
        const workspaceId = response.data.data.workspace_id;

        // Check if this content belongs to one of the user's workspaces
        const workspaceExists = workspaces.some((ws) => ws.id === workspaceId);

        if (!workspaceExists) {
          setError("You don't have access to this content's workspace");
          return;
        }

        // If the content belongs to a different workspace than currently selected,
        // consider switching workspaces automatically
        if (workspaceId !== currentWorkspace?.id) {
          const workspace = workspaces.find((ws) => ws.id === workspaceId);
          if (workspace) {
            // Option 1: Switch workspace automatically
            setCurrentWorkspace(workspace);

            // Option 2: Or ask the user if they want to switch
            // if (confirm(`This content belongs to workspace "${workspace.name}". Switch to this workspace?`)) {
            //   setCurrentWorkspace(workspace);
            // } else {
            //   navigate('/content');
            //   return;
            // }
          }
        }

        setContent({
          title: response.data.data.title || "",
          slug: response.data.data.slug || "",
          body: response.data.data.body || "",
          status: response.data.data.status || "draft",
          content_type_id: response.data.data.content_type_id || "",
          workspace_id: response.data.data.workspace_id,
          meta_data: response.data.data.meta_data || "",
        });
      } else {
        setError(response.data.error || "Failed to load content");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to load content");
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (
    e: React.ChangeEvent<
      HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
    >
  ) => {
    const { name, value } = e.target;
    setContent((prev) => ({ ...prev, [name]: value }));

    // Auto-generate slug from title
    if (name === "title" && (!content.slug || content.slug === "")) {
      const slug = value
        .toLowerCase()
        .replace(/[^\w\s-]/g, "")
        .replace(/\s+/g, "-")
        .replace(/--+/g, "-")
        .trim();

      setContent((prev) => ({ ...prev, slug }));
    }
  };

  const handleEditorChange = ({ text }: { text: string }) => {
    setContent((prev) => ({ ...prev, body: text }));
  };

  const handleSubmit = async (e: React.FormEvent, publish: boolean = false) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setIsSaving(true);

    const contentToSave = {
      ...content,
      status: publish ? "published" : content.status,
    };

    try {
      let response;

      if (id && id !== "new") {
        // This is editing an existing post
        response = await axios.put(
          `/api/workspaces/${currentWorkspace.id}/content/${id}`,
          contentToSave
        );
      } else {
        // This is creating a new post
        response = await axios.post(
          `/api/workspaces/${currentWorkspace.id}/content`,
          contentToSave
        );
      }

      if (response.data.success) {
        setSuccess("Content saved successfully");
        navigate("/content")
      } else {
        setError(response.data.error || "Failed to save content");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to save content");
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="text-center py-8">
        <div className="text-primary-600 font-semibold text-xl">
          Loading content...
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white shadow rounded-lg p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">
          {id && id !== "new" ? "Edit Content" : "Create New Content"}
        </h1>
        <div className="flex space-x-2">
          <button
            onClick={() => navigate("/content")}
            className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={(e) => handleSubmit(e, false)}
            disabled={isSaving}
            className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 disabled:bg-primary-400"
          >
            {isSaving ? "Saving..." : "Save Draft"}
          </button>
          <button
            onClick={(e) => handleSubmit(e, true)}
            disabled={isSaving}
            className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-green-600 hover:bg-green-700 disabled:bg-green-400"
          >
            {isSaving ? "Publishing..." : "Publish"}
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md text-sm">
          {error}
        </div>
      )}

      {success && (
        <div className="mb-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-md text-sm">
          {success}
        </div>
      )}

      <form onSubmit={(e) => handleSubmit(e, false)}>
        <div className="space-y-6">
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
            <div>
              <label
                htmlFor="title"
                className="block text-sm font-medium text-gray-700"
              >
                Title
              </label>
              <input
                type="text"
                name="title"
                id="title"
                required
                value={content.title}
                onChange={handleInputChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
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
                value={content.slug}
                onChange={handleInputChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
            <div>
              <label
                htmlFor="content_type_id"
                className="block text-sm font-medium text-gray-700"
              >
                Content Type
              </label>
              <select
                name="content_type_id"
                id="content_type_id"
                value={content.content_type_id}
                onChange={handleInputChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              >
                <option value="">Select a content type</option>
                {contentTypes.map((type) => (
                  <option key={type.id} value={type.id}>
                    {type.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label
                htmlFor="status"
                className="block text-sm font-medium text-gray-700"
              >
                Status
              </label>
              <select
                name="status"
                id="status"
                value={content.status}
                onChange={handleInputChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              >
                <option value="draft">Draft</option>
                <option value="published">Published</option>
                <option value="archived">Archived</option>
              </select>
            </div>
          </div>

          <div>
            <label
              htmlFor="body"
              className="block text-sm font-medium text-gray-700"
            >
              Content (Markdown)
            </label>
            <div className="mt-1 border border-gray-300 rounded-md">
              <MarkdownEditor
                value={content.body}
                onChange={handleEditorChange}
                renderHTML={(text) => <ReactMarkdown>{text}</ReactMarkdown>}
                style={{ height: "400px" }}
              />
            </div>
          </div>

          <div>
            <label
              htmlFor="meta_data"
              className="block text-sm font-medium text-gray-700"
            >
              Meta Data (JSON)
            </label>
            <textarea
              name="meta_data"
              id="meta_data"
              rows={4}
              value={content.meta_data}
              onChange={handleInputChange}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
            />
          </div>
        </div>
      </form>
    </div>
  );
}
