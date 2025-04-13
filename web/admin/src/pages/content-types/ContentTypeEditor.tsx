// src/pages/content-types/ContentTypeEditor.tsx
import { useState, useEffect, useContext } from "react";
import { useParams, useNavigate } from "react-router-dom";
import axios from "axios";
import AuthContext from "../../context/AuthContext";

export default function ContentTypeEditor() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { currentWorkspace } = useContext(AuthContext);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const [contentType, setContentType] = useState({
    name: "",
    slug: "",
    description: "",
    workspace_id: currentWorkspace?.id || "",
    fields: [],
  });

  const [fields, setFields] = useState<
    Array<{
      name: string;
      type: string;
      required: boolean;
      description: string;
    }>
  >([]);

  useEffect(() => {
    if (currentWorkspace?.id) {
      setContentType((prev) => ({
        ...prev,
        workspace_id: currentWorkspace.id,
      }));

      if (id && id !== "new") {
        loadContentType(id);
      }
    }
  }, [currentWorkspace, id]);

  const loadContentType = async (contentTypeId: string) => {
    setIsLoading(true);
    setError("");

    try {
      const response = await axios.get(`/api/content-types/${contentTypeId}`);
      if (response.data.success) {
        const data = response.data.data;
        setContentType({
          name: data.name || "",
          slug: data.slug || "",
          description: data.description || "",
          workspace_id: data.workspace_id || currentWorkspace?.id || "",
          fields: data.fields || [],
        });
        setFields(data.fields || []);
      } else {
        setError("Failed to load content type");
      }
    } catch (err: any) {
      setError("Failed to load content type");
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setContentType((prev) => ({ ...prev, [name]: value }));

    // Auto-generate slug from name
    if (name === "name" && (!contentType.slug || contentType.slug === "")) {
      const slug = value
        .toLowerCase()
        .replace(/[^\w\s-]/g, "")
        .replace(/\s+/g, "-")
        .replace(/--+/g, "-")
        .trim();

      setContentType((prev) => ({ ...prev, slug }));
    }
  };

  const handleFieldChange = (index: number, field: string, value: any) => {
    const updatedFields = [...fields];
    updatedFields[index] = { ...updatedFields[index], [field]: value };
    setFields(updatedFields);
  };

  const addField = () => {
    setFields([
      ...fields,
      { name: "", type: "text", required: false, description: "" },
    ]);
  };

  const removeField = (index: number) => {
    const updatedFields = [...fields];
    updatedFields.splice(index, 1);
    setFields(updatedFields);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setIsSaving(true);

    const contentTypeData = {
      ...contentType,
      fields: fields,
    };

    try {
      let response;

      if (id && id !== "new") {
        response = await axios.put(`/api/content-types/${id}`, contentTypeData);
      } else {
        response = await axios.post("/api/content-types", contentTypeData);
      }

      if (response.data.success) {
        setSuccess("Content type saved successfully");
        if (id === "new") {
          navigate(`/content-types/${response.data.data.id}`);
        }
      } else {
        setError(response.data.error || "Failed to save content type");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to save content type");
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="text-center py-8">
        <div className="text-primary-600 font-semibold text-xl">Loading...</div>
      </div>
    );
  }

  return (
    <div className="bg-white shadow rounded-lg p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">
          {id && id !== "new" ? "Edit Content Type" : "Create New Content Type"}
        </h1>
        <div className="flex space-x-2">
          <button
            onClick={() => navigate("/content-types")}
            className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={isSaving}
            className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 disabled:bg-primary-400"
          >
            {isSaving ? "Saving..." : "Save"}
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

      <form onSubmit={handleSubmit}>
        <div className="space-y-6">
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
            <div>
              <label
                htmlFor="name"
                className="block text-sm font-medium text-gray-700"
              >
                Name
              </label>
              <input
                type="text"
                name="name"
                id="name"
                required
                value={contentType.name}
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
                value={contentType.slug}
                onChange={handleInputChange}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              />
            </div>
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
              value={contentType.description}
              onChange={handleInputChange}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
            />
          </div>

          <div>
            <div className="flex justify-between items-center mb-2">
              <label className="block text-sm font-medium text-gray-700">
                Fields
              </label>
              <button
                type="button"
                onClick={addField}
                className="px-3 py-1 border border-transparent rounded-md text-xs font-medium text-white bg-primary-600 hover:bg-primary-700"
              >
                Add Field
              </button>
            </div>

            {fields.length === 0 ? (
              <div className="text-center py-4 text-gray-500 border border-dashed border-gray-300 rounded-md">
                No fields defined. Click "Add Field" to create your first field.
              </div>
            ) : (
              <div className="space-y-4">
                {fields.map((field, index) => (
                  <div
                    key={index}
                    className="border border-gray-200 rounded-md p-4"
                  >
                    <div className="flex justify-between items-center mb-3">
                      <h3 className="text-sm font-medium text-gray-700">
                        Field #{index + 1}
                      </h3>
                      <button
                        type="button"
                        onClick={() => removeField(index)}
                        className="text-red-600 hover:text-red-900 text-xs"
                      >
                        Remove
                      </button>
                    </div>
                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                      <div>
                        <label className="block text-xs font-medium text-gray-700">
                          Name
                        </label>
                        <input
                          type="text"
                          value={field.name}
                          onChange={(e) =>
                            handleFieldChange(index, "name", e.target.value)
                          }
                          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                          required
                        />
                      </div>
                      <div>
                        <label className="block text-xs font-medium text-gray-700">
                          Type
                        </label>
                        <select
                          value={field.type}
                          onChange={(e) =>
                            handleFieldChange(index, "type", e.target.value)
                          }
                          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                        >
                          <option value="text">Text</option>
                          <option value="textarea">Textarea</option>
                          <option value="number">Number</option>
                          <option value="boolean">Boolean</option>
                          <option value="date">Date</option>
                          <option value="select">Select</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-xs font-medium text-gray-700">
                          Description
                        </label>
                        <input
                          type="text"
                          value={field.description}
                          onChange={(e) =>
                            handleFieldChange(
                              index,
                              "description",
                              e.target.value
                            )
                          }
                          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                        />
                      </div>
                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          checked={field.required}
                          onChange={(e) =>
                            handleFieldChange(
                              index,
                              "required",
                              e.target.checked
                            )
                          }
                          className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                        />
                        <label className="ml-2 block text-xs font-medium text-gray-700">
                          Required
                        </label>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </form>
    </div>
  );
}
