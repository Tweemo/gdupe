import type { AnalyzeResult, SimilarityGroup } from "./types";

// Empty default = same-origin relative requests (e.g. "/api/sessions"), which
// is how the production build is served (Go serves the static frontend and the
// API from one origin). For split local dev, web/.env.development points this
// at the standalone Go server on :8080.
const BASE = process.env.NEXT_PUBLIC_API_BASE ?? "";

export async function createSession(): Promise<string> {
  const res = await fetch(`${BASE}/api/sessions`, { method: "POST" });
  if (!res.ok) throw new Error("Could not start a session");
  const body = (await res.json()) as { sessionId: string };
  return body.sessionId;
}

// uploadFiles streams the selected files to the server, reporting 0..1 progress.
// Each file's webkitRelativePath is carried as the multipart part filename so
// the server can reconstruct the folder structure.
export function uploadFiles(
  sessionId: string,
  files: File[],
  onProgress?: (fraction: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const form = new FormData();
    for (const f of files) {
      form.append("files", f, f.webkitRelativePath || f.name);
    }
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${BASE}/api/sessions/${sessionId}/upload`);
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) onProgress(e.loaded / e.total);
    };
    xhr.onload = () =>
      xhr.status >= 200 && xhr.status < 300
        ? resolve()
        : reject(new Error(`Upload failed (${xhr.status})`));
    xhr.onerror = () => reject(new Error("Upload failed"));
    xhr.send(form);
  });
}

export async function analyze(sessionId: string): Promise<AnalyzeResult> {
  const res = await fetch(`${BASE}/api/sessions/${sessionId}/analyze`, {
    method: "POST",
  });
  if (!res.ok) throw new Error("Analysis failed");
  return (await res.json()) as AnalyzeResult;
}

export function thumbnailUrl(sessionId: string, fileId: string): string {
  return `${BASE}/api/sessions/${sessionId}/thumbnail/${fileId}`;
}

export async function exportZip(
  sessionId: string,
  groups: SimilarityGroup[],
): Promise<void> {
  const res = await fetch(`${BASE}/api/sessions/${sessionId}/export`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ groups }),
  });
  if (!res.ok) throw new Error("Export failed");
}

export function downloadUrl(sessionId: string): string {
  return `${BASE}/api/sessions/${sessionId}/download`;
}

export async function deleteSession(sessionId: string): Promise<void> {
  await fetch(`${BASE}/api/sessions/${sessionId}`, { method: "DELETE" });
}
