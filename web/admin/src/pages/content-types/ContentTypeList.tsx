// src/pages/content-types/ContentTypeList.tsx
import { useState, useEffect, useContext } from "react";
import { Link } from "react-router-dom";
import axios from "axios";
import AuthContext from "../../context/AuthContext";

export default function ContentTypeList() {
  const { currentWorkspace } = useContext(AuthContext);
  const [contentTypes, setContentTypes] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (currentWorkspace?.id) {
      loadContentTypes();
    }
  }, [currentWorkspace]);

  const loadContentTypes = async () => {
    setIsLoading(true);
    setError("");

    try {
      const response = await axios.get(
        `/api/content-types?workspace_id=${currentWorkspace.id}`
      );

      if (response.data.success) {
        setContentTypes(response.data.data);
      } else {
        setError(response.data.error || "Failed to load content types");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to load content types");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm("Are you sure you want to delete this content type?")) {
      return;
    }

    try {
      const response = await axios.delete(`/api/content-types/${id}`);

      if (response.data.success) {
        setContentTypes(contentTypes.filter((ct) => ct.id !== id));
      } else {
        alert(response.data.error || "Failed to delete content type");
      }
    } catch (err: any) {
      alert(err.response?.data?.error || "Failed to delete content type");
    }
  };

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
        <h1 className="text-xl font-semibold text-gray-900">Content Types</h1>
        <Link
          to="/content-types/new"
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700"
        >
          Create New
        </Link>
      </div>

      {error && (
        <div className="px-4 py-3 bg-red-50 border-b border-red-200 text-red-700 text-sm">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="px-4 py-8 text-center text-gray-500">
          Loading content types...
        </div>
      ) : contentTypes.length === 0 ? (
        <div className="px-4 py-8 text-center text-gray-500">
          No content types found. Create your first content type to get started.
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Name
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Slug
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Fields
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {contentTypes.map((contentType) => (
                <tr key={contentType.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">
                      {contentType.name}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {contentType.slug}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {contentType.fields?.length || 0} fields
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <Link
                      to={`/content-types/${contentType.id}`}
                      className="text-primary-600 hover:text-primary-900 mr-4"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDelete(contentType.id)}
                      className="text-red-600 hover:text-red-900"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
