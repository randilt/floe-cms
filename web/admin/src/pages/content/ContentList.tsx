// src/pages/content/ContentList.tsx
import { useState, useEffect, useContext } from "react";
import { Link } from "react-router-dom";
import axios from "axios";
import AuthContext from "../../context/AuthContext";

export default function ContentList() {
  const { currentWorkspace } = useContext(AuthContext);
  const [contents, setContents] = useState<any[]>([]);
  const [contentTypes, setContentTypes] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);

  const [filters, setFilters] = useState({
    status: "",
    content_type_id: "",
  });

  useEffect(() => {
    if (currentWorkspace?.id) {
      loadContentTypes();
      loadContents();
    }
  }, [currentWorkspace, page, limit, filters]);

  const loadContentTypes = async () => {
    try {
      const response = await axios.get(
        `/api/content-types?workspace_id=${currentWorkspace.id}`
      );
      if (response.data.success) {
        setContentTypes(response.data.data);
      }
    } catch (err: any) {
      console.error("Failed to load content types:", err);
    }
  };

  const loadContents = async () => {
    setIsLoading(true);
    setError("");

    try {
      const offset = (page - 1) * limit;
      let url = `/api/content?workspace_id=${currentWorkspace.id}&limit=${limit}&offset=${offset}`;

      if (filters.status) {
        url += `&status=${filters.status}`;
      }

      if (filters.content_type_id) {
        url += `&content_type_id=${filters.content_type_id}`;
      }

      const response = await axios.get(url);

      if (response.data.success) {
        setContents(response.data.data.contents);
        setTotal(response.data.data.total);
      } else {
        setError(response.data.error || "Failed to load contents");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to load contents");
    } finally {
      setIsLoading(false);
    }
  };

  const handleFilterChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFilters((prev) => ({ ...prev, [name]: value }));
    setPage(1); // Reset to first page when changing filters
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm("Are you sure you want to delete this content?")) {
      return;
    }

    try {
      const response = await axios.delete(`/api/content/${id}`);

      if (response.data.success) {
        // Remove the deleted content from the state
        setContents(contents.filter((c) => c.id !== id));

        // Update total count
        setTotal((prev) => prev - 1);
      } else {
        alert(response.data.error || "Failed to delete content");
      }
    } catch (err: any) {
      alert(err.response?.data?.error || "Failed to delete content");
    }
  };

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
        <h1 className="text-xl font-semibold text-gray-900">Content</h1>
        <Link
          to="/content/new"
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700"
        >
          Create New
        </Link>
      </div>

      <div className="px-4 py-3 border-b border-gray-200 bg-gray-50 sm:px-6">
        <div className="flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-4">
          <div className="w-full sm:w-1/4">
            <label
              htmlFor="status"
              className="block text-xs font-medium text-gray-700"
            >
              Status
            </label>
            <select
              id="status"
              name="status"
              value={filters.status}
              onChange={handleFilterChange}
              className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm rounded-md"
            >
              <option value="">All Statuses</option>
              <option value="draft">Draft</option>
              <option value="published">Published</option>
              <option value="archived">Archived</option>
            </select>
          </div>

          <div className="w-full sm:w-1/4">
            <label
              htmlFor="content_type_id"
              className="block text-xs font-medium text-gray-700"
            >
              Content Type
            </label>
            <select
              id="content_type_id"
              name="content_type_id"
              value={filters.content_type_id}
              onChange={handleFilterChange}
              className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm rounded-md"
            >
              <option value="">All Types</option>
              {contentTypes.map((type) => (
                <option key={type.id} value={type.id}>
                  {type.name}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {error && (
        <div className="px-4 py-3 bg-red-50 border-b border-red-200 text-red-700 text-sm">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="px-4 py-8 text-center text-gray-500">
          Loading contents...
        </div>
      ) : contents.length === 0 ? (
        <div className="px-4 py-8 text-center text-gray-500">
          No content items found.
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
                  Title
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Type
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Status
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Author
                </th>
                <th
                  scope="col"
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Updated
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
              {contents.map((content) => (
                <tr key={content.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">
                      {content.title}
                    </div>
                    <div className="text-sm text-gray-500">{content.slug}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">
                      {content.content_type?.name || "No type"}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                      ${
                        content.status === "published"
                          ? "bg-green-100 text-green-800"
                          : content.status === "draft"
                          ? "bg-yellow-100 text-yellow-800"
                          : "bg-gray-100 text-gray-800"
                      }`}
                    >
                      {content.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {content.author?.first_name} {content.author?.last_name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(content.updated_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <Link
                      to={`/content/${content.id}`}
                      className="text-primary-600 hover:text-primary-900 mr-4"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDelete(content.id)}
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
