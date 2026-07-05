// Mirrors the JSON shapes returned by the Go server (internal/media).

export type MediaType = "image" | "video" | "audio" | "other";

export interface MediaFile {
  id: string;
  relPath: string;
  sourceFolder: string;
  size: number;
  type: MediaType;
}

export interface DuplicateSet {
  hash: string;
  keptFileId: string;
  fileIds: string[];
}

export interface SimilarityGroup {
  id: string;
  fileIds: string[];
}

export interface AnalyzeResult {
  kept: MediaFile[];
  duplicates: DuplicateSet[];
  groups: SimilarityGroup[];
}
