// src/pages/media/MediaLibrary.tsx
import { useState, useEffect, useContext, useRef } from "react";
import axios from "axios";
import AuthContext from "../../context/AuthContext";

export default function MediaLibrary() {
  const { currentWorkspace } = useContext(AuthContext);
  const [mediaItems, setMediaItems] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [uploading, setUploading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(12);

  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (currentWorkspace?.id) {
      loadMedia();
    }
  }, [currentWorkspace, page, limit]);

  const loadMedia = async () => {
    setIsLoading(true);
    setError("");

    try {
      const offset = (page - 1) * limit;
      const response = await axios.get(
        `/api/media?workspace_id=${currentWorkspace.id}&limit=${limit}&offset=${offset}`
      );

      if (response.data.success) {
        setMediaItems(response.data.data.media);
        setTotal(response.data.data.total);
      } else {
        setError(response.data.error || "Failed to load media");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to load media");
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0) {
      return;
    }

    setUploading(true);
    setError("");
    setSuccess("");

    const formData = new FormData();
    formData.append("workspace_id", currentWorkspace.id.toString());
    formData.append("file", e.target.files[0]);
    formData.append("name", e.target.files[0].name);

    try {
      const response = await axios.post("/api/media", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (response.data.success) {
        setSuccess("File uploaded successfully");
        loadMedia(); // Reload media list
        if (fileInputRef.current) {
          fileInputRef.current.value = "";
        }
      } else {
        setError(response.data.error || "Failed to upload file");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to upload file");
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm("Are you sure you want to delete this media?")) {
      return;
    }

    try {
      const response = await axios.delete(`/api/media/${id}`);

      if (response.data.success) {
        setMediaItems(mediaItems.filter((m) => m.id !== id));
        setTotal((prev) => prev - 1);
        setSuccess("Media deleted successfully");
      } else {
        setError(response.data.error || "Failed to delete media");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to delete media");
    }
  };

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <div className="px-4 py-5 sm:px-6">
        <h1 className="text-xl font-semibold text-gray-900">Media Library</h1>
      </div>

      <div className="px-4 py-5 sm:px-6 border-t border-gray-200">
        <div className="flex items-center space-x-4">
          <input
            type="file"
            ref={fileInputRef}
            onChange={handleFileChange}
            className="hidden"
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading}
            className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 disabled:bg-primary-400"
          >
            {uploading ? "Uploading..." : "Upload New File"}
          </button>
          <div className="text-sm text-gray-500">
            Upload images, documents, and other files for your content.
          </div>
        </div>

        {error && (
          <div className="mt-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md text-sm">
            {error}
          </div>
        )}

        {success && (
          <div className="mt-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-md text-sm">
            {success}
          </div>
        )}
      </div>

      {isLoading ? (
        <div className="px-4 py-8 text-center text-gray-500">
          Loading media...
        </div>
      ) : mediaItems.length === 0 ? (
        <div className="px-4 py-8 text-center text-gray-500">
          No media files found. Upload your first file to get started.
        </div>
      ) : (
        <div className="px-4 py-5 sm:px-6 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {mediaItems.map((media) => (
            <div
              key={media.id}
              className="border border-gray-200 rounded-md overflow-hidden"
            >
              <div className="h-40 bg-gray-100 flex items-center justify-center">
                {media.mime_type?.startsWith("image/") ? (
                  <img
                    src={media.file_path}
                    alt={media.name}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="p-4 text-center">
                    <div className="bg-gray-200 rounded-md p-4 mb-2">
                      <svg
                        className="h-6 w-6 text-gray-500 mx-auto"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                        ></path>
                      </svg>
                    </div>
                    <span className="text-xs text-gray-500">
                      {media.mime_type}
                    </span>
                  </div>
                )}
              </div>
              <div className="p-3">
                <div className="text-sm font-medium text-gray-900 truncate">
                  {media.name}
                </div>
                <div className="text-xs text-gray-500 mt-1">
                  {(media.size / 1024).toFixed(2)} KB
                </div>
                <div className="flex justify-between mt-2">
                  <a
                    href={media.file_path}
                    target="_blank"
                    rel="noreferrer"
                    className="text-xs text-primary-600 hover:text-primary-900"
                  >
                    View
                  </a>
                  <button
                    onClick={() => handleDelete(media.id)}
                    className="text-xs text-red-600 hover:text-red-900"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
          <div className="flex-1 flex justify-between items-center">
            <div>
              <p className="text-sm text-gray-700">
                Showing{" "}
                <span className="font-medium">
                  {Math.min((page - 1) * limit + 1, total)}
                </span>{" "}
                to{" "}
                <span className="font-medium">
                  {Math.min(page * limit, total)}
                </span>{" "}
                of <span className="font-medium">{total}</span> results
              </p>
            </div>
            <div className="flex space-x-2">
              <button
                onClick={() => setPage((prev) => Math.max(prev - 1, 1))}
                disabled={page === 1}
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:bg-gray-100 disabled:text-gray-400"
              >
                Previous
              </button>
              <button
                onClick={() =>
                  setPage((prev) => Math.min(prev + 1, totalPages))
                }
                disabled={page === totalPages}
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:bg-gray-100 disabled:text-gray-400"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
