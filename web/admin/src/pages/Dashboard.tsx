// src/pages/Dashboard.tsx
import { useState, useEffect, useContext } from "react";
import { Link } from "react-router-dom";
import axios from "axios";
import AuthContext from "../context/AuthContext";

export default function Dashboard() {
  const { user, currentWorkspace } = useContext(AuthContext);
  const [stats, setStats] = useState({
    contentCount: 0,
    draftCount: 0,
    publishedCount: 0,
    mediaCount: 0,
    recentContents: [],
  });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (currentWorkspace?.id) {
      loadDashboardData();
    }
  }, [currentWorkspace]);

  const loadDashboardData = async () => {
    setIsLoading(true);

    try {
      // Get content stats
      const contentsResponse = await axios.get(
        `/api/content?workspace_id=${currentWorkspace.id}&limit=0`
      );
      const publishedResponse = await axios.get(
        `/api/content?workspace_id=${currentWorkspace.id}&status=published&limit=0`
      );
      const draftResponse = await axios.get(
        `/api/content?workspace_id=${currentWorkspace.id}&status=draft&limit=0`
      );
      const mediaResponse = await axios.get(
        `/api/media?workspace_id=${currentWorkspace.id}&limit=0`
      );
      const recentResponse = await axios.get(
        `/api/content?workspace_id=${currentWorkspace.id}&limit=5`
      );

      setStats({
        contentCount: contentsResponse.data.data.total,
        publishedCount: publishedResponse.data.data.total,
        draftCount: draftResponse.data.data.total,
        mediaCount: mediaResponse.data.data.total,
        recentContents: recentResponse.data.data.contents,
      });
    } catch (error) {
      console.error("Error loading dashboard data:", error);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className="text-center py-8">
        <div className="text-primary-600 font-semibold text-xl">
          Loading dashboard...
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="bg-white overflow-hidden shadow rounded-lg">
        <div className="px-4 py-5 sm:px-6">
          <h2 className="text-lg font-medium text-gray-900">
            Welcome, {user?.first_name || "User"}
          </h2>
          <p className="mt-1 text-sm text-gray-500">
            Current workspace:{" "}
            <span className="font-medium">{currentWorkspace?.name}</span>
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {/* Content Count */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-primary-100 rounded-md p-3">
                <svg
                  className="h-6 w-6 text-primary-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                  ></path>
                </svg>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Total Content
                  </dt>
                  <dd>
                    <div className="text-lg font-medium text-gray-900">
                      {stats.contentCount}
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-4 sm:px-6">
            <div className="text-sm">
              <Link
                to="/content"
                className="font-medium text-primary-600 hover:text-primary-500"
              >
                View all
              </Link>
            </div>
          </div>
        </div>

        {/* Published Content */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-green-100 rounded-md p-3">
                <svg
                  className="h-6 w-6 text-green-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                  ></path>
                </svg>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Published
                  </dt>
                  <dd>
                    <div className="text-lg font-medium text-gray-900">
                      {stats.publishedCount}
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-4 sm:px-6">
            <div className="text-sm">
              <Link
                to="/content?status=published"
                className="font-medium text-primary-600 hover:text-primary-500"
              >
                View published
              </Link>
            </div>
          </div>
        </div>

        {/* Draft Content */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-yellow-100 rounded-md p-3">
                <svg
                  className="h-6 w-6 text-yellow-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  ></path>
                </svg>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Drafts
                  </dt>
                  <dd>
                    <div className="text-lg font-medium text-gray-900">
                      {stats.draftCount}
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-4 sm:px-6">
            <div className="text-sm">
              <Link
                to="/content?status=draft"
                className="font-medium text-primary-600 hover:text-primary-500"
              >
                View drafts
              </Link>
            </div>
          </div>
        </div>

        {/* Media Count */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-purple-100 rounded-md p-3">
                <svg
                  className="h-6 w-6 text-purple-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                  ></path>
                </svg>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Media Files
                  </dt>
                  <dd>
                    <div className="text-lg font-medium text-gray-900">
                      {stats.mediaCount}
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-4 sm:px-6">
            <div className="text-sm">
              <Link
                to="/media"
                className="font-medium text-primary-600 hover:text-primary-500"
              >
                View media
              </Link>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Content */}
      <div className="bg-white shadow sm:rounded-lg">
        <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
          <h3 className="text-lg leading-6 font-medium text-gray-900">
            Recent Content
          </h3>
          <Link
            to="/content/new"
            className="px-3 py-1 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700"
          >
            Create New
          </Link>
        </div>
        <div className="border-t border-gray-200">
          <div className="overflow-hidden sm:rounded-lg">
            {stats.recentContents.length === 0 ? (
              <div className="px-4 py-5 text-center text-gray-500 sm:px-6">
                No content created yet.
              </div>
            ) : (
              <ul className="divide-y divide-gray-200">
                {stats.recentContents.map((content: any) => (
                  <li
                    key={content.id}
                    className="px-4 py-4 sm:px-6 hover:bg-gray-50"
                  >
                    <div className="flex items-center justify-between">
                      <div className="truncate">
                        <Link
                          to={`/content/${content.id}`}
                          className="text-sm font-medium text-primary-600 truncate hover:text-primary-900"
                        >
                          {content.title}
                        </Link>
                        <p className="text-sm text-gray-500 truncate">
                          {content.slug}
                        </p>
                      </div>
                      <div className="ml-2 flex-shrink-0 flex">
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
                      </div>
                    </div>
                    <div className="mt-2 flex justify-between">
                      <div className="text-xs text-gray-500">
                        {content.content_type?.name || "No type"} • Updated{" "}
                        {new Date(content.updated_at).toLocaleDateString()}
                      </div>
                      <div className="text-xs text-gray-500">
                        by {content.author?.first_name}{" "}
                        {content.author?.last_name}
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
