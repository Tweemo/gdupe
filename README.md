# Media Merge

Combine several folders of media into one clean folder: exact duplicates are
removed (across images, video, and audio) and visually-similar images are
grouped into subfolders that you confirm before exporting. The result is
downloaded as `merged.zip`.

- **Frontend:** Next.js (App Router, TypeScript, Tailwind) in `web/`
- **Backend:** Go HTTP server in `server/`

## How it works

1. Select one or more folders in the browser. Files upload to the server.
2. The server hashes every file (SHA-256) and removes byte-exact duplicates,
   then perceptually hashes images and clusters visually-similar ones.
3. You review/adjust the proposed similarity groups.
4. The server builds and returns `merged.zip`:

   ```
   merged/
     image-group-001/   # confirmed similar images
     images/            # unique images with no similar match
     video/  audio/  other/
   ```

**Note:** HEIC/HEIF files are exact-deduped and placed by type but are skipped
for visual grouping (no robust pure-Go decoder).

## Running locally

Backend (defaults to `:8080`):

```bash
cd server
go run ./cmd/server
```

Frontend (defaults to talking to `http://localhost:8080`):

```bash
cd web
npm install      # first time only
npm run dev      # http://localhost:3000
```

In split dev the frontend (`:3000`) reaches the API (`:8080`) via
`web/.env.development`. Override the API base with `NEXT_PUBLIC_API_BASE` and the
server address with `MEDIA_MERGE_ADDR` / work directory with `MEDIA_MERGE_WORKDIR`.

## Run with Docker

The frontend is built to static files and served by the Go server, so the whole
app runs as a single container on one port:

```bash
docker build -t media-merge .
docker run --rm -p 8080:8080 media-merge
```

Open http://localhost:8080. The page and the API share one origin, so no
API-URL configuration is needed. Uploaded files and merged zips live in the
container's temp dir and are cleared when the container stops.

## Tests

```bash
cd server && go test ./...     # backend unit + HTTP integration tests
cd web && npm run build        # typecheck + lint + production build
```
