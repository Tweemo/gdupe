"use client";

import { useMemo, useState } from "react";
import type { AnalyzeResult, MediaFile, SimilarityGroup } from "@/lib/types";
import { thumbnailUrl } from "@/lib/api";

const UNGROUPED = "__ungrouped__";

interface Props {
  sessionId: string;
  result: AnalyzeResult;
  onConfirm: (groups: SimilarityGroup[]) => void;
}

// GroupReview lets the user adjust the proposed similar-image groups. Every
// image carries a "Move to" selector; reassigning images is enough to split a
// group, merge groups, or pull an image out (ungroup) — all through one control.
export default function GroupReview({ sessionId, result, onConfirm }: Props) {
  const byId = useMemo(() => {
    const m = new Map<string, MediaFile>();
    for (const f of result.kept) m.set(f.id, f);
    return m;
  }, [result.kept]);

  const images = result.kept.filter((f) => f.type === "image");
  const videoCount = result.kept.filter((f) => f.type === "video").length;
  const audioCount = result.kept.filter((f) => f.type === "audio").length;

  // assignment: imageId -> group key (a group id, or UNGROUPED).
  const [assignment, setAssignment] = useState<Record<string, string>>(() => {
    const init: Record<string, string> = {};
    for (const f of images) init[f.id] = UNGROUPED;
    for (const g of result.groups) {
      for (const id of g.fileIds) init[id] = g.id;
    }
    return init;
  });
  const [nextGroup, setNextGroup] = useState(result.groups.length + 1);

  // Current group keys present (excluding ungrouped), ordered.
  const groupKeys = useMemo(() => {
    const keys = new Set<string>();
    for (const id of Object.keys(assignment)) {
      const k = assignment[id];
      if (k !== UNGROUPED) keys.add(k);
    }
    return Array.from(keys).sort();
  }, [assignment]);

  const ungrouped = images.filter((f) => assignment[f.id] === UNGROUPED);

  function move(imageId: string, key: string) {
    setAssignment((a) => ({ ...a, [imageId]: key }));
  }

  function newGroupFor(imageId: string) {
    const key = `new-${nextGroup}`;
    setNextGroup((n) => n + 1);
    move(imageId, key);
  }

  function handleConfirm() {
    // Only real groups (>= 2 images) are sent; singletons fall to images/.
    const groups: SimilarityGroup[] = [];
    let idx = 1;
    for (const key of groupKeys) {
      const fileIds = images
        .filter((f) => assignment[f.id] === key)
        .map((f) => f.id);
      if (fileIds.length >= 2) {
        groups.push({
          id: `image-group-${String(idx).padStart(3, "0")}`,
          fileIds,
        });
        idx++;
      }
    }
    onConfirm(groups);
  }

  function MoveSelect({ imageId }: { imageId: string }) {
    const current = assignment[imageId];
    return (
      <select
        value={current}
        onChange={(e) => {
          if (e.target.value === "__new__") newGroupFor(imageId);
          else move(imageId, e.target.value);
        }}
        className="mt-1 w-full rounded border text-xs px-1 py-0.5"
      >
        <option value={UNGROUPED}>Ungrouped</option>
        {groupKeys.map((k, i) => (
          <option key={k} value={k}>
            Group {i + 1}
          </option>
        ))}
        <option value="__new__">+ New group</option>
      </select>
    );
  }

  function Thumb({ file }: { file: MediaFile }) {
    return (
      <div className="w-28">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={thumbnailUrl(sessionId, file.id)}
          alt={file.relPath}
          className="h-28 w-28 rounded object-cover bg-gray-100"
        />
        <p className="truncate text-xs text-gray-500" title={file.relPath}>
          {file.relPath.split("/").pop()}
        </p>
        <MoveSelect imageId={file.id} />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800">
        Removed <strong>{result.duplicates.length}</strong> set
        {result.duplicates.length === 1 ? "" : "s"} of exact duplicates.{" "}
        {videoCount > 0 && <>· {videoCount} video(s) </>}
        {audioCount > 0 && <>· {audioCount} audio file(s) </>}
        will be included, sorted by type.
      </div>

      {groupKeys.map((key, i) => {
        const members = images.filter((f) => assignment[f.id] === key);
        return (
          <section key={key} className="rounded-lg border p-4">
            <h3 className="mb-3 font-semibold">
              Group {i + 1}{" "}
              <span className="text-sm font-normal text-gray-500">
                ({members.length} image{members.length === 1 ? "" : "s"})
              </span>
            </h3>
            <div className="flex flex-wrap gap-4">
              {members.map((f) => (
                <Thumb key={f.id} file={byId.get(f.id)!} />
              ))}
            </div>
          </section>
        );
      })}

      <section className="rounded-lg border border-dashed p-4">
        <h3 className="mb-3 font-semibold text-gray-600">
          Ungrouped images{" "}
          <span className="text-sm font-normal text-gray-400">
            (each kept individually)
          </span>
        </h3>
        {ungrouped.length === 0 ? (
          <p className="text-sm text-gray-400">None.</p>
        ) : (
          <div className="flex flex-wrap gap-4">
            {ungrouped.map((f) => (
              <Thumb key={f.id} file={f} />
            ))}
          </div>
        )}
      </section>

      <button
        type="button"
        onClick={handleConfirm}
        className="rounded-lg bg-blue-600 px-6 py-3 font-medium text-white hover:bg-blue-700"
      >
        Confirm &amp; build merged folder
      </button>
    </div>
  );
}
