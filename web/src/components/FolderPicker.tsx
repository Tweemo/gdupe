"use client";

import { useEffect, useRef } from "react";

export interface SelectedFolder {
  name: string;
  files: File[];
}

interface Props {
  folders: SelectedFolder[];
  onAddFolder: (folder: SelectedFolder) => void;
  onRemoveFolder: (name: string) => void;
}

export default function FolderPicker({
  folders,
  onAddFolder,
  onRemoveFolder,
}: Props) {
  const inputRef = useRef<HTMLInputElement>(null);

  // webkitdirectory / directory aren't part of the standard React typings.
  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.setAttribute("webkitdirectory", "");
      inputRef.current.setAttribute("directory", "");
    }
  }, []);

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const fileList = e.target.files;
    if (!fileList || fileList.length === 0) return;
    const files = Array.from(fileList);
    // The folder name is the first path segment of any file.
    const name = files[0].webkitRelativePath.split("/")[0] || "folder";
    onAddFolder({ name, files });
    e.target.value = ""; // allow re-selecting the same folder
  }

  const totalFiles = folders.reduce((n, f) => n + f.files.length, 0);

  return (
    <div className="space-y-4">
      <button
        type="button"
        onClick={() => inputRef.current?.click()}
        className="rounded-lg border-2 border-dashed border-gray-400 px-6 py-8 w-full text-gray-600 hover:border-blue-500 hover:text-blue-600 transition"
      >
        + Add a folder of media
      </button>
      <input
        ref={inputRef}
        type="file"
        multiple
        className="hidden"
        onChange={handleChange}
      />

      {folders.length > 0 && (
        <ul className="divide-y rounded-lg border">
          {folders.map((f) => (
            <li
              key={f.name}
              className="flex items-center justify-between px-4 py-2"
            >
              <span className="font-medium">{f.name}</span>
              <span className="flex items-center gap-3 text-sm text-gray-500">
                {f.files.length} files
                <button
                  type="button"
                  onClick={() => onRemoveFolder(f.name)}
                  className="text-red-500 hover:underline"
                >
                  remove
                </button>
              </span>
            </li>
          ))}
        </ul>
      )}

      {totalFiles > 0 && (
        <p className="text-sm text-gray-500">
          {totalFiles} files across {folders.length} folder
          {folders.length === 1 ? "" : "s"} selected.
        </p>
      )}
    </div>
  );
}
