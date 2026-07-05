"use client";

import { useState } from "react";
import FolderPicker, { SelectedFolder } from "@/components/FolderPicker";
import GroupReview from "@/components/GroupReview";
import * as api from "@/lib/api";
import type { AnalyzeResult, SimilarityGroup } from "@/lib/types";

type Step = "select" | "uploading" | "analyzing" | "review" | "exporting" | "done";

export default function Home() {
  const [step, setStep] = useState<Step>("select");
  const [folders, setFolders] = useState<SelectedFolder[]>([]);
  const [progress, setProgress] = useState(0);
  const [sessionId, setSessionId] = useState<string>("");
  const [result, setResult] = useState<AnalyzeResult | null>(null);
  const [error, setError] = useState<string>("");

  const allFiles = folders.flatMap((f) => f.files);

  function addFolder(folder: SelectedFolder) {
    setFolders((prev) =>
      prev.some((f) => f.name === folder.name) ? prev : [...prev, folder],
    );
  }

  function removeFolder(name: string) {
    setFolders((prev) => prev.filter((f) => f.name !== name));
  }

  async function start() {
    setError("");
    try {
      const id = await api.createSession();
      setSessionId(id);

      setStep("uploading");
      await api.uploadFiles(id, allFiles, setProgress);

      setStep("analyzing");
      const res = await api.analyze(id);
      setResult(res);
      setStep("review");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Something went wrong");
      setStep("select");
    }
  }

  async function confirm(groups: SimilarityGroup[]) {
    setError("");
    try {
      setStep("exporting");
      await api.exportZip(sessionId, groups);
      setStep("done");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Export failed");
      setStep("review");
    }
  }

  function reset() {
    if (sessionId) api.deleteSession(sessionId).catch(() => {});
    setStep("select");
    setFolders([]);
    setProgress(0);
    setSessionId("");
    setResult(null);
    setError("");
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-12">
      <h1 className="mb-2 text-3xl font-bold">Media Merge</h1>
      <p className="mb-8 text-gray-500">
        Combine folders of media into one — duplicates removed, similar photos
        grouped.
      </p>

      {error && (
        <div className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {step === "select" && (
        <div className="space-y-6">
          <FolderPicker
            folders={folders}
            onAddFolder={addFolder}
            onRemoveFolder={removeFolder}
          />
          <button
            type="button"
            disabled={allFiles.length === 0}
            onClick={start}
            className="rounded-lg bg-blue-600 px-6 py-3 font-medium text-white hover:bg-blue-700 disabled:opacity-40"
          >
            Upload &amp; analyze{" "}
            {allFiles.length > 0 && `(${allFiles.length} files)`}
          </button>
        </div>
      )}

      {step === "uploading" && (
        <div className="space-y-3">
          <p>Uploading…</p>
          <div className="h-3 w-full overflow-hidden rounded-full bg-gray-200">
            <div
              className="h-full bg-blue-600 transition-all"
              style={{ width: `${Math.round(progress * 100)}%` }}
            />
          </div>
          <p className="text-sm text-gray-500">{Math.round(progress * 100)}%</p>
        </div>
      )}

      {step === "analyzing" && (
        <p className="text-gray-600">
          Analyzing for duplicates and similar images…
        </p>
      )}

      {step === "review" && result && (
        <GroupReview
          sessionId={sessionId}
          result={result}
          onConfirm={confirm}
        />
      )}

      {step === "exporting" && (
        <p className="text-gray-600">Building your merged folder…</p>
      )}

      {step === "done" && (
        <div className="space-y-6">
          <div className="rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-green-800">
            Your merged folder is ready.
          </div>
          <a
            href={api.downloadUrl(sessionId)}
            className="inline-block rounded-lg bg-blue-600 px-6 py-3 font-medium text-white hover:bg-blue-700"
          >
            Download merged.zip
          </a>
          <div>
            <button
              type="button"
              onClick={reset}
              className="text-sm text-gray-500 hover:underline"
            >
              Start over
            </button>
          </div>
        </div>
      )}
    </main>
  );
}
